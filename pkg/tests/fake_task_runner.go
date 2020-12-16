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

package tests

import (
	"context"
	"testing"

	"github.com/mattmoor/mink/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tektonclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"github.com/tektoncd/pipeline/pkg/pod"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

// FakeTaskRunner creates a fake test runner that simulates tekton running tasks using fake clients
func FakeTaskRunner(t *testing.T, ctx context.Context, tektonClient tektonclientset.Interface, ns string, fakeDigests []string) {
	taskRunInterface := tektonClient.TektonV1beta1().TaskRuns(ns)
	watcher, err := taskRunInterface.Watch(ctx, metav1.ListOptions{})
	require.NoError(t, err, "failed to watch TaskRuns in namespace %s", ns)

	ch := watcher.ResultChan()

	counter := 0
	for {
		// See if our context has been cancelled
		select {
		case <-ctx.Done():
			t.Logf("completed the fake task runner")
			return

		case event := <-ch:
			tr, ok := event.Object.(*v1beta1.TaskRun)
			assert.True(t, ok, "invalid object found in watcher %v", event.Object)
			if ok && tr != nil {
				cond := tr.Status.GetCondition(apis.ConditionSucceeded)
				name := tr.Name
				if name == "" {
					name = tr.GenerateName
				}
				if cond == nil || cond.IsFalse() {
					tr2 := &v1beta1.TaskRun{
						TypeMeta:   tr.TypeMeta,
						ObjectMeta: tr.ObjectMeta,
						Spec:       tr.Spec,
						Status:     tr.Status,
					}
					tr = tr2
					pod.MarkStatusSuccess(&tr.Status)

					// lets add a result if there is not one already
					fakeDigest := fakeDigests[counter%len(fakeDigests)]
					found := false
					for _, result := range tr.Status.TaskRunResults {
						if result.Name == constants.ImageDigestResult {
							found = true
						}
					}
					if !found {
						tr.Status.TaskRunResults = append(tr.Status.TaskRunResults, v1beta1.TaskRunResult{
							Name:  constants.ImageDigestResult,
							Value: fakeDigest,
						})
					}
					_, err = taskRunInterface.Update(ctx, tr, metav1.UpdateOptions{})
					if err != nil {
						t.Logf("WARNING: failed to update TaskRun %s to complete: %s\n", name, err.Error())
					} else {
						t.Logf("updated TaskRun %s to complete\n", name)
					}
				} else {
					t.Logf("TaskRun %s is completed\n", name)
				}
			}
		}
	}
}
