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
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/tektoncd/cli/pkg/cmd/taskrun"
	"github.com/tektoncd/cli/pkg/options"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tektonclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/pkg/apis"
)

// CancelableOption is a function option that can be used to customize a taskrun in
// certain ways prior to execution.  Each option may return a cancel function, which
// can be used to clean up any temporary artifacts created in support of this task run.
type CancelableOption func(context.Context, *tknv1beta1.TaskRun) (context.CancelFunc, error)

// Run executes the provided TaskRun with the provided options applied, and returns
// the fully-qualified image digest (or error) upon completion.
func Run(ctx context.Context, image string, tr *tknv1beta1.TaskRun, opt *options.LogOptions, opts ...CancelableOption) (name.Digest, error) {
	// TODO(mattmoor): expose masterURL and kubeconfig flags.
	cfg, err := GetConfig("", "")
	if err != nil {
		return name.Digest{}, err
	}
	client, err := tektonclientset.NewForConfig(cfg)
	if err != nil {
		return name.Digest{}, err
	}

	for _, o := range opts {
		cancel, err := o(ctx, tr)
		if err != nil {
			return name.Digest{}, err
		}
		defer cancel()
	}

	tr, err = client.TektonV1beta1().TaskRuns(tr.Namespace).Create(ctx, tr, metav1.CreateOptions{})
	if err != nil {
		return name.Digest{}, err
	}

	// TODO(mattmoor): From here down assumes opt.Follow, but if we want to have
	// a --no-wait or something then we should have an early-out here.
	defer client.TektonV1beta1().TaskRuns(tr.Namespace).Delete(context.Background(), tr.Name, metav1.DeleteOptions{})

	opt.TaskrunName = tr.Name
	if err := streamLogs(ctx, opt); err != nil {
		return name.Digest{}, err
	}

	// Spin waiting for the final status.
	for {
		// See if our context has been cancelled
		select {
		case <-ctx.Done():
			return name.Digest{}, ctx.Err()
		default:
		}

		// Fetch the final state of the build.
		tr, err = client.TektonV1beta1().TaskRuns(tr.Namespace).Get(ctx, tr.Name, metav1.GetOptions{})
		if err != nil {
			return name.Digest{}, err
		}

		// Return an error if the build failed.
		cond := tr.Status.GetCondition(apis.ConditionSucceeded)
		if cond.IsFalse() {
			return name.Digest{}, fmt.Errorf("%s: %s", cond.Reason, cond.Message)
		} else if !cond.IsTrue() {
			continue
		}

		for _, result := range tr.Status.TaskRunResults {
			if result.Name != "IMAGE-DIGEST" {
				continue
			}
			value := strings.TrimSpace(result.Value)

			// Extract the IMAGE-DIGEST result.
			digest, err := name.NewDigest(image + "@" + value)
			if err != nil {
				return name.Digest{}, err
			}

			return digest, nil
		}
	}
}

func streamLogs(ctx context.Context, opt *options.LogOptions) error {
	// TODO(mattmoor): This should take a context so that it can be cancelled.
	errCh := make(chan error)
	go func() {
		defer close(errCh)
		errCh <- taskrun.Run(opt)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WithServiceAccount is used to adjust the TaskRun to execute as a particular
// service account, as specified by the user.  It supports a special "me" sentinel
// which configures a temporary ServiceAccount infused with the local credentials
// for the container registry hosting the image we will publish to (and to which
// the source is published).
func WithServiceAccount(sa string, tag name.Tag) CancelableOption {
	cfg, err := GetConfig("", "")
	if err != nil {
		log.Fatalf("GetConfig() = %v", err)
	}
	client := kubernetes.NewForConfigOrDie(cfg)

	return func(ctx context.Context, tr *tknv1beta1.TaskRun) (context.CancelFunc, error) {
		if sa != "me" {
			tr.Spec.ServiceAccountName = sa
			return func() {}, nil
		}

		// Fetch the user's auth for the provided build target
		authenticator, err := authn.DefaultKeychain.Resolve(tag)
		if err != nil {
			return nil, err
		}
		auth, err := authenticator.Authorization()
		if err != nil {
			return nil, err
		}

		// Create a secret and service account for this build.
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: tr.GenerateName,
				Namespace:    tr.Namespace,
				Annotations: map[string]string{
					// Use the funny form so that it works with DockerHub.
					"tekton.dev/docker-0": "https://" + tag.RegistryStr() + "/v1/",
				},
			},
			Type: corev1.SecretTypeBasicAuth,
			StringData: map[string]string{
				corev1.BasicAuthUsernameKey: auth.Username,
				corev1.BasicAuthPasswordKey: auth.Password,
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
			Secrets: []corev1.ObjectReference{{
				Name: secret.Name,
			}},
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

// GetConfig is forked out of sharedmain because linking knative.dev/pkg/metrics spews logs.
func GetConfig(masterURL, kubeconfig string) (*rest.Config, error) {
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
	// If we have an explicit indication of where the kubernetes config lives, read that.
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	}
	// If not, try the in-cluster config.
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory.
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags("", filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not create a valid kubeconfig")
}
