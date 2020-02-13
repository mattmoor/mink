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

package contour

import (
	contour "github.com/projectcontour/contour/apis/projectcontour/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	fakecontourclientset "knative.dev/net-contour/pkg/client/clientset/versioned/fake"
	contourlisters "knative.dev/net-contour/pkg/client/listers/projectcontour/v1"
	"knative.dev/pkg/reconciler/testing"
	networking "knative.dev/serving/pkg/apis/networking/v1alpha1"
	fakeservingclientset "knative.dev/serving/pkg/client/clientset/versioned/fake"
	networkinglisters "knative.dev/serving/pkg/client/listers/networking/v1alpha1"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakeservingclientset.AddToScheme,
	fakecontourclientset.AddToScheme,
	fakekubeclientset.AddToScheme,
}

type Listers struct {
	sorter testing.ObjectSorter
}

func NewListers(objs []runtime.Object) Listers {
	scheme := NewScheme()

	ls := Listers{
		sorter: testing.NewObjectSorter(scheme),
	}

	ls.sorter.AddObjects(objs...)

	return ls
}

func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	for _, addTo := range clientSetSchemes {
		addTo(scheme)
	}
	return scheme
}

func (*Listers) NewScheme() *runtime.Scheme {
	return NewScheme()
}

// IndexerFor returns the indexer for the given object.
func (l *Listers) IndexerFor(obj runtime.Object) cache.Indexer {
	return l.sorter.IndexerForObjectType(obj)
}

func (l *Listers) GetServingObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeservingclientset.AddToScheme)
}

func (l *Listers) GetContourObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakecontourclientset.AddToScheme)
}

func (l *Listers) GetKubeObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakekubeclientset.AddToScheme)
}

// GetIngressLister get lister for Ingress resource.
func (l *Listers) GetIngressLister() networkinglisters.IngressLister {
	return networkinglisters.NewIngressLister(l.IndexerFor(&networking.Ingress{}))
}

// GetHTTPProxyLister get lister for HTTPProxy resource.
func (l *Listers) GetHTTPProxyLister() contourlisters.HTTPProxyLister {
	return contourlisters.NewHTTPProxyLister(l.IndexerFor(&contour.HTTPProxy{}))
}

// GetK8sServiceLister get lister for K8s Service resource.
func (l *Listers) GetK8sServiceLister() corev1listers.ServiceLister {
	return corev1listers.NewServiceLister(l.IndexerFor(&corev1.Service{}))
}

// GetEndpointsLister get lister for K8s Endpoints resource.
func (l *Listers) GetEndpointsLister() corev1listers.EndpointsLister {
	return corev1listers.NewEndpointsLister(l.IndexerFor(&corev1.Endpoints{}))
}
