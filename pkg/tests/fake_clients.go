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
