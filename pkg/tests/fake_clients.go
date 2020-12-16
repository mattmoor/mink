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

	tektonclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	tektonfake "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/fake"
	pipelineclient "github.com/tektoncd/pipeline/pkg/client/injection/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
)

// FakeContextWithClients creates a context and fake kube and pipeline clients for testing
func FakeContextWithClients(_ *testing.T) (context.Context, kubernetes.Interface, tektonclientset.Interface) {
	ctx := context.TODO()
	kubeClient := fake.NewSimpleClientset()
	tektonClient := tektonfake.NewSimpleClientset()
	ctx = context.WithValue(ctx, kubeclient.Key{}, kubeClient)
	ctx = context.WithValue(ctx, pipelineclient.Key{}, tektonClient)
	return ctx, kubeClient, tektonClient
}
