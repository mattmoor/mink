/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package builds

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/tektoncd/cli/pkg/options"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	pipelineclient "github.com/tektoncd/pipeline/pkg/client/injection/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
)

// CancelableTaskOption is a function option that can be used to customize a taskrun in
// certain ways prior to execution.  Each option may return a cancel function, which
// can be used to clean up any temporary artifacts created in support of this task run.
type CancelableTaskOption func(context.Context, *tknv1beta1.TaskRun) (context.CancelFunc, error)

// RunTask executes the provided TaskRun with the provided options applied, and returns
// the final TaskRun state (or error) upon completion.
func RunTask(ctx context.Context, tr *tknv1beta1.TaskRun, opt *options.LogOptions, opts ...CancelableTaskOption) (*tknv1beta1.TaskRun, error) {
	client := pipelineclient.Get(ctx)

	for _, o := range opts {
		cancel, err := o(ctx, tr)
		if err != nil {
			return nil, err
		}
		defer cancel()
	}

	tr, err := client.TektonV1beta1().TaskRuns(tr.Namespace).Create(ctx, tr, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	// TODO(mattmoor): From here down assumes opt.Follow, but if we want to have
	// a --no-wait or something then we should have an early-out here.
	defer client.TektonV1beta1().TaskRuns(tr.Namespace).Delete(context.Background(), tr.Name, metav1.DeleteOptions{})

	opt.TaskrunName = tr.Name
	if err := streamLogs(ctx, opt); err != nil {
		return nil, err
	}

	// Spin waiting for the final status.
	for {
		// See if our context has been cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Fetch the final state of the build.
		tr, err = client.TektonV1beta1().TaskRuns(tr.Namespace).Get(ctx, tr.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// Return an error if the build failed.
		cond := tr.Status.GetCondition(apis.ConditionSucceeded)
		if cond.IsFalse() {
			return nil, fmt.Errorf("%s: %s", cond.Reason, cond.Message)
		} else if !cond.IsTrue() {
			continue
		}

		return tr, nil
	}
}

// WithTaskServiceAccount is used to adjust the TaskRun to execute as a particular
// service account, as specified by the user.  It supports a special "me" sentinel
// which configures a temporary ServiceAccount infused with the local credentials
// for the container registry hosting the image we will publish to (and to which
// the source is published).
func WithTaskServiceAccount(ctx context.Context, sa string, refs ...name.Reference) CancelableTaskOption {
	client := kubeclient.Get(ctx)

	return func(ctx context.Context, tr *tknv1beta1.TaskRun) (context.CancelFunc, error) {
		if sa != "me" {
			tr.Spec.ServiceAccountName = sa
			return func() {}, nil
		}

		cfg := struct {
			Auths map[string]*authn.AuthConfig `json:"auths"`
		}{
			Auths: make(map[string]*authn.AuthConfig, len(refs)),
		}

		for _, ref := range refs {
			// Fetch the user's auth for the provided build target
			authenticator, err := authn.DefaultKeychain.Resolve(ref.Context())
			if err != nil {
				return nil, err
			}
			auth, err := authenticator.Authorization()
			if err != nil {
				return nil, err
			}
			// Use the funny form so that it works with DockerHub.
			cfg.Auths["https://"+ref.Context().RegistryStr()+"/v1/"] = auth
		}
		b, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}

		// Create a secret and service account for this build.
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: tr.GenerateName,
				Namespace:    tr.Namespace,
			},
			Type: corev1.SecretTypeDockerConfigJson,
			StringData: map[string]string{
				corev1.DockerConfigJsonKey: string(b),
			},
		}
		secret, err = client.CoreV1().Secrets(secret.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
		cleansecret := func() {
			err := client.CoreV1().Secrets(secret.Namespace).Delete(context.Background(), secret.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Printf("WARNING: Secret %q leaked, error cleaning up: %v", secret.Name, err)
			}
		}

		sa := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: tr.GenerateName,
				Namespace:    tr.Namespace,
			},
			// Support pulling source using the user credentials.
			ImagePullSecrets: []corev1.LocalObjectReference{{
				Name: secret.Name,
			}},
		}

		if tr.Spec.TaskSpec != nil {
			// Mount the credentials secret as a volume.
			volumeName := fmt.Sprint("mink-creds-", rand.Uint64()) //nolint:gosec
			tr.Spec.PodTemplate.Volumes = append(tr.Spec.PodTemplate.Volumes, corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{{
							Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: secret.Name,
								},
								Items: []corev1.KeyToPath{{
									Key:  corev1.DockerConfigJsonKey,
									Path: "config.json",
									// Mode defaults to 0644
								}},
							},
						}},
					},
				},
			})

			// How we will mount the credentials into steps.
			vm := corev1.VolumeMount{
				Name:      volumeName,
				MountPath: fmt.Sprint("/var/mink/creds/", rand.Uint64()), //nolint:gosec // Randomize to avoid hardcoding (weak ok)
			}

			for i := range tr.Spec.TaskSpec.Steps {
				for j, env := range tr.Spec.TaskSpec.Steps[i].Env {
					if env.Name == "DOCKER_CONFIG" {
						// When steps specify DOCKER_CONFIG, override it's value and attach our mount.
						tr.Spec.TaskSpec.Steps[i].Env[j].Value = vm.MountPath
						tr.Spec.TaskSpec.Steps[i].VolumeMounts = append(tr.Spec.TaskSpec.Steps[i].VolumeMounts, vm)
						break
					}
				}
			}
		} else {
			sa.Secrets = []corev1.ObjectReference{{
				Name: secret.Name,
			}}
		}

		sa, err = client.CoreV1().ServiceAccounts(sa.Namespace).Create(ctx, sa, metav1.CreateOptions{})
		if err != nil {
			cleansecret()
			return nil, err
		}
		cleansa := func() {
			err := client.CoreV1().ServiceAccounts(sa.Namespace).Delete(context.Background(), sa.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Printf("WARNING: ServiceAccount %q leaked, error cleaning up: %v", sa.Name, err)
			}
		}

		tr.Spec.ServiceAccountName = sa.Name

		return func() {
			cleansa()
			cleansecret()
		}, nil
	}
}
