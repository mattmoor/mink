/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package test

import (
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/pkg/test"

	// Support running e2e on GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/vmware-tanzu/sources-for-knative/pkg/client/clientset/versioned"
	sources "github.com/vmware-tanzu/sources-for-knative/pkg/client/clientset/versioned/typed/sources/v1alpha1"
)

const (
	// To simplify testing against our sample yamls, run our testing
	// in the default namespace.
	Namespace = "default"

	// PollInterval is how frequently e2e tests will poll for updates.
	PollInterval = 1 * time.Second
	// PollTimeout is how long e2e tests will wait for resource updates when polling.
	PollTimeout = 1 * time.Minute
)

type Clients struct {
	KubeClient   *test.KubeClient
	VMWareClient *VMWareClients
}

type VMWareClients struct {
	Bindings sources.VSphereBindingInterface
	Sources  sources.VSphereSourceInterface
}

func NewClients(configPath string, clusterName string, namespace string) (*Clients, error) {
	cfg, err := BuildClientConfig(configPath, clusterName)
	if err != nil {
		return nil, err
	}
	return NewClientsFromConfig(cfg, namespace)
}

func Setup(t test.TLegacy) *Clients {
	t.Helper()
	clients, err := NewClients(test.Flags.Kubeconfig, test.Flags.Cluster, Namespace)
	if err != nil {
		t.Fatal("Couldn't initialize clients", "error", err.Error())
	}
	return clients
}

func NewClientsFromConfig(cfg *rest.Config, namespace string) (*Clients, error) {
	clients := &Clients{}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	clients.KubeClient = &test.KubeClient{Kube: kubeClient}

	cs, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	clients.VMWareClient = &VMWareClients{
		Bindings: cs.SourcesV1alpha1().VSphereBindings(namespace),
		Sources:  cs.SourcesV1alpha1().VSphereSources(namespace),
	}

	return clients, nil
}

func BuildClientConfig(kubeConfigPath string, clusterName string) (*rest.Config, error) {
	overrides := clientcmd.ConfigOverrides{}
	// Override the cluster name if provided.
	if clusterName != "" {
		overrides.Context.Cluster = clusterName
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath},
		&overrides).ClientConfig()
}
