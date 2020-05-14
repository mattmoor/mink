// Copyright © 2019 VMware
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dag

import (
	"github.com/projectcontour/contour/internal/annotation"
	"github.com/projectcontour/contour/internal/k8s"

	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/client-go/tools/cache"

	ingressroutev1 "github.com/projectcontour/contour/apis/contour/v1beta1"
	projectcontour "github.com/projectcontour/contour/apis/projectcontour/v1"
	"github.com/sirupsen/logrus"
	serviceapis "sigs.k8s.io/service-apis/api/v1alpha1"
)

// A KubernetesCache holds Kubernetes objects and associated configuration and produces
// DAG values.
type KubernetesCache struct {
	// RootNamespaces specifies the namespaces where root
	// IngressRoutes can be defined. If empty, roots can be defined in any
	// namespace.
	RootNamespaces []string

	// Contour's IngressClass.
	// If not set, defaults to DEFAULT_INGRESS_CLASS.
	IngressClass string

	ingresses            map[k8s.FullName]*v1beta1.Ingress
	ingressroutes        map[k8s.FullName]*ingressroutev1.IngressRoute
	httpproxies          map[k8s.FullName]*projectcontour.HTTPProxy
	secrets              map[k8s.FullName]*v1.Secret
	irdelegations        map[k8s.FullName]*ingressroutev1.TLSCertificateDelegation
	httpproxydelegations map[k8s.FullName]*projectcontour.TLSCertificateDelegation
	services             map[k8s.FullName]*v1.Service
	gatewayclasses       map[k8s.FullName]*serviceapis.GatewayClass
	gateways             map[k8s.FullName]*serviceapis.Gateway
	httproutes           map[k8s.FullName]*serviceapis.HTTPRoute
	tcproutes            map[k8s.FullName]*serviceapis.TcpRoute

	logrus.FieldLogger
}

// matchesIngressClass returns true if the given Kubernetes object
// belongs to the Ingress class that this cache is using.
func (kc *KubernetesCache) matchesIngressClass(obj k8s.Object) bool {

	if !annotation.MatchesIngressClass(obj, kc.IngressClass) {
		kind := k8s.KindOf(obj)
		om := obj.GetObjectMeta()

		kc.WithField("name", om.GetName()).
			WithField("namespace", om.GetNamespace()).
			WithField("kind", kind).
			WithField("ingress.class", annotation.IngressClass(obj)).
			Debug("ignoring object with unmatched ingress class")
		return false
	}

	return true

}

// Insert inserts obj into the KubernetesCache.
// Insert returns true if the cache accepted the object, or false if the value
// is not interesting to the cache. If an object with a matching type, name,
// and namespace exists, it will be overwritten.
func (kc *KubernetesCache) Insert(obj interface{}) bool {
	if obj, ok := obj.(k8s.Object); ok {
		kind := k8s.KindOf(obj)
		for key := range obj.GetObjectMeta().GetAnnotations() {
			// Emit a warning if this is a known annotation that has
			// been applied to an invalid object kind. Note that we
			// only warn for known annotations because we want to
			// allow users to add arbitrary orthogonal annotations
			// to objects that we inspect.
			if annotation.IsKnown(key) && !annotation.ValidForKind(kind, key) {
				// TODO(jpeach): this should be exposed
				// to the user as a status condition.
				om := obj.GetObjectMeta()
				kc.WithField("name", om.GetName()).
					WithField("namespace", om.GetNamespace()).
					WithField("kind", kind).
					WithField("version", "v1").
					WithField("annotation", key).
					Error("ignoring invalid or unsupported annotation")
			}
		}
	}

	switch obj := obj.(type) {
	case *v1.Secret:
		valid, err := isValidSecret(obj)
		if !valid {
			if err != nil {
				om := obj.GetObjectMeta()
				kc.WithField("name", om.GetName()).
					WithField("namespace", om.GetNamespace()).
					WithField("kind", "Secret").
					WithField("version", "v1").
					Error(err)
			}
			return false
		}

		m := k8s.ToFullName(obj)
		if kc.secrets == nil {
			kc.secrets = make(map[k8s.FullName]*v1.Secret)
		}
		kc.secrets[m] = obj
		return kc.secretTriggersRebuild(obj)
	case *v1.Service:
		m := k8s.ToFullName(obj)
		if kc.services == nil {
			kc.services = make(map[k8s.FullName]*v1.Service)
		}
		kc.services[m] = obj
		return kc.serviceTriggersRebuild(obj)
	case *v1beta1.Ingress:
		if kc.matchesIngressClass(obj) {
			m := k8s.ToFullName(obj)
			if kc.ingresses == nil {
				kc.ingresses = make(map[k8s.FullName]*v1beta1.Ingress)
			}
			kc.ingresses[m] = obj
			return true
		}
	case *ingressroutev1.IngressRoute:
		if kc.matchesIngressClass(obj) {
			m := k8s.ToFullName(obj)
			if kc.ingressroutes == nil {
				kc.ingressroutes = make(map[k8s.FullName]*ingressroutev1.IngressRoute)
			}
			kc.ingressroutes[m] = obj
			return true
		}
	case *projectcontour.HTTPProxy:
		if kc.matchesIngressClass(obj) {
			m := k8s.ToFullName(obj)
			if kc.httpproxies == nil {
				kc.httpproxies = make(map[k8s.FullName]*projectcontour.HTTPProxy)
			}
			kc.httpproxies[m] = obj
			return true
		}
	case *ingressroutev1.TLSCertificateDelegation:
		m := k8s.ToFullName(obj)
		if kc.irdelegations == nil {
			kc.irdelegations = make(map[k8s.FullName]*ingressroutev1.TLSCertificateDelegation)
		}
		kc.irdelegations[m] = obj
		return true
	case *projectcontour.TLSCertificateDelegation:
		m := k8s.ToFullName(obj)
		if kc.httpproxydelegations == nil {
			kc.httpproxydelegations = make(map[k8s.FullName]*projectcontour.TLSCertificateDelegation)
		}
		kc.httpproxydelegations[m] = obj
		return true
	case *serviceapis.GatewayClass:
		m := k8s.ToFullName(obj)
		if kc.gatewayclasses == nil {
			kc.gatewayclasses = make(map[k8s.FullName]*serviceapis.GatewayClass)
		}
		// TODO(youngnick): Remove this once service-apis actually have behavior
		// other than being added to the cache.
		kc.WithField("experimental", "service-apis").WithField("name", m.Name).WithField("namespace", m.Namespace).Debug("Adding GatewayClass")
		kc.gatewayclasses[m] = obj
		return true
	case *serviceapis.Gateway:
		m := k8s.ToFullName(obj)
		if kc.gateways == nil {
			kc.gateways = make(map[k8s.FullName]*serviceapis.Gateway)
		}
		// TODO(youngnick): Remove this once service-apis actually have behavior
		// other than being added to the cache.
		kc.WithField("experimental", "service-apis").WithField("name", m.Name).WithField("namespace", m.Namespace).Debug("Adding Gateway")
		kc.gateways[m] = obj
		return true
	case *serviceapis.HTTPRoute:
		m := k8s.ToFullName(obj)
		if kc.httproutes == nil {
			kc.httproutes = make(map[k8s.FullName]*serviceapis.HTTPRoute)
		}
		// TODO(youngnick): Remove this once service-apis actually have behavior
		// other than being added to the cache.
		kc.WithField("experimental", "service-apis").WithField("name", m.Name).WithField("namespace", m.Namespace).Debug("Adding HTTPRoute")
		kc.httproutes[m] = obj
		return true
	case *serviceapis.TcpRoute:
		m := k8s.ToFullName(obj)
		if kc.tcproutes == nil {
			kc.tcproutes = make(map[k8s.FullName]*serviceapis.TcpRoute)
		}
		// TODO(youngnick): Remove this once service-apis actually have behavior
		// other than being added to the cache.
		kc.WithField("experimental", "service-apis").WithField("name", m.Name).WithField("namespace", m.Namespace).Debug("Adding TcpRoute")
		kc.tcproutes[m] = obj
		return true

	default:
		// not an interesting object
		kc.WithField("object", obj).Error("insert unknown object")
		return false
	}

	return false
}

// Remove removes obj from the KubernetesCache.
// Remove returns a boolean indicating if the cache changed after the remove operation.
func (kc *KubernetesCache) Remove(obj interface{}) bool {
	switch obj := obj.(type) {
	default:
		return kc.remove(obj)
	case cache.DeletedFinalStateUnknown:
		return kc.Remove(obj.Obj) // recurse into ourselves with the tombstoned value
	}
}

func (kc *KubernetesCache) remove(obj interface{}) bool {
	switch obj := obj.(type) {
	case *v1.Secret:
		m := k8s.ToFullName(obj)
		_, ok := kc.secrets[m]
		delete(kc.secrets, m)
		return ok
	case *v1.Service:
		m := k8s.ToFullName(obj)
		_, ok := kc.services[m]
		delete(kc.services, m)
		return ok
	case *v1beta1.Ingress:
		m := k8s.ToFullName(obj)
		_, ok := kc.ingresses[m]
		delete(kc.ingresses, m)
		return ok
	case *ingressroutev1.IngressRoute:
		m := k8s.ToFullName(obj)
		_, ok := kc.ingressroutes[m]
		delete(kc.ingressroutes, m)
		return ok
	case *projectcontour.HTTPProxy:
		m := k8s.ToFullName(obj)
		_, ok := kc.httpproxies[m]
		delete(kc.httpproxies, m)
		return ok
	case *ingressroutev1.TLSCertificateDelegation:
		m := k8s.ToFullName(obj)
		_, ok := kc.irdelegations[m]
		delete(kc.irdelegations, m)
		return ok
	case *projectcontour.TLSCertificateDelegation:
		m := k8s.ToFullName(obj)
		_, ok := kc.httpproxydelegations[m]
		delete(kc.httpproxydelegations, m)
		return ok
	case *serviceapis.GatewayClass:
		m := k8s.ToFullName(obj)
		_, ok := kc.gatewayclasses[m]
		// TODO(youngnick): Remove this once service-apis actually have behavior
		// other than being removed from the cache.
		kc.WithField("experimental", "service-apis").WithField("name", m.Name).WithField("namespace", m.Namespace).Debug("Removing GatewayClass")
		delete(kc.gatewayclasses, m)
		return ok
	case *serviceapis.Gateway:
		m := k8s.ToFullName(obj)
		_, ok := kc.gateways[m]
		// TODO(youngnick): Remove this once service-apis actually have behavior
		// other than being removed from the cache.
		kc.WithField("experimental", "service-apis").WithField("name", m.Name).WithField("namespace", m.Namespace).Debug("Removing Gateway")
		delete(kc.gateways, m)
		return ok
	case *serviceapis.HTTPRoute:
		m := k8s.ToFullName(obj)
		_, ok := kc.httproutes[m]
		// TODO(youngnick): Remove this once service-apis actually have behavior
		// other than being removed from the cache.
		kc.WithField("experimental", "service-apis").WithField("name", m.Name).WithField("namespace", m.Namespace).Debug("Removing HTTPRoute")
		delete(kc.httproutes, m)
		return ok
	case *serviceapis.TcpRoute:
		m := k8s.ToFullName(obj)
		_, ok := kc.tcproutes[m]
		// TODO(youngnick): Remove this once service-apis actually have behavior
		// other than being removed from the cache.
		kc.WithField("experimental", "service-apis").WithField("name", m.Name).WithField("namespace", m.Namespace).Debug("Removing TcpRoute")
		delete(kc.tcproutes, m)
		return ok

	default:
		// not interesting
		kc.WithField("object", obj).Error("remove unknown object")
		return false
	}
}

// serviceTriggersRebuild returns true if this service is referenced
// by an Ingress or IngressRoute in this cache.
func (kc *KubernetesCache) serviceTriggersRebuild(service *v1.Service) bool {
	for _, ingress := range kc.ingresses {
		if ingress.Namespace != service.Namespace {
			continue
		}
		if backend := ingress.Spec.Backend; backend != nil {
			if backend.ServiceName == service.Name {
				return true
			}
		}

		for _, rule := range ingress.Spec.Rules {
			http := rule.IngressRuleValue.HTTP
			if http == nil {
				continue
			}
			for _, path := range http.Paths {
				if path.Backend.ServiceName == service.Name {
					return true
				}
			}
		}
	}

	for _, ir := range kc.ingressroutes {
		if ir.Namespace != service.Namespace {
			continue
		}
		for _, route := range ir.Spec.Routes {
			for _, s := range route.Services {
				if s.Name == service.Name {
					return true
				}
			}
		}
		if tcpproxy := ir.Spec.TCPProxy; tcpproxy != nil {
			for _, s := range tcpproxy.Services {
				if s.Name == service.Name {
					return true
				}
			}
		}
	}

	for _, ir := range kc.httpproxies {
		if ir.Namespace != service.Namespace {
			continue
		}
		for _, route := range ir.Spec.Routes {
			for _, s := range route.Services {
				if s.Name == service.Name {
					return true
				}
			}
		}
		if tcpproxy := ir.Spec.TCPProxy; tcpproxy != nil {
			for _, s := range tcpproxy.Services {
				if s.Name == service.Name {
					return true
				}
			}
		}
	}

	return false
}

// secretTriggersRebuild returns true if this secret is referenced by an Ingress
// or IngressRoute object in this cache. If the secret is not in the same namespace
// it must be mentioned by a TLSCertificateDelegation.
func (kc *KubernetesCache) secretTriggersRebuild(secret *v1.Secret) bool {
	if _, isCA := secret.Data[CACertificateKey]; isCA {
		// locating a secret validation usage involves traversing each
		// ingressroute object, determining if there is a valid delegation,
		// and if the reference the secret as a certificate. The DAG already
		// does this so don't reproduce the logic and just assume for the moment
		// that any change to a CA secret will trigger a rebuild.
		return true
	}

	delegations := make(map[string]bool) // targetnamespace/secretname to bool

	// merge ingressroute.TLSCertificateDelegation and projectcontour.TLSCertificateDelegation.
	for _, d := range kc.irdelegations {
		for _, cd := range d.Spec.Delegations {
			for _, n := range cd.TargetNamespaces {
				delegations[n+"/"+cd.SecretName] = true
			}
		}
	}
	for _, d := range kc.httpproxydelegations {
		for _, cd := range d.Spec.Delegations {
			for _, n := range cd.TargetNamespaces {
				delegations[n+"/"+cd.SecretName] = true
			}
		}
	}

	for _, ingress := range kc.ingresses {
		if ingress.Namespace == secret.Namespace {
			for _, tls := range ingress.Spec.TLS {
				if tls.SecretName == secret.Name {
					return true
				}
			}
		}
		if delegations[ingress.Namespace+"/"+secret.Name] {
			for _, tls := range ingress.Spec.TLS {
				if tls.SecretName == secret.Namespace+"/"+secret.Name {
					return true
				}
			}
		}

		if delegations["*/"+secret.Name] {
			for _, tls := range ingress.Spec.TLS {
				if tls.SecretName == secret.Namespace+"/"+secret.Name {
					return true
				}
			}
		}
	}

	for _, ir := range kc.ingressroutes {
		vh := ir.Spec.VirtualHost
		if vh == nil {
			// not a root ingress
			continue
		}
		tls := vh.TLS
		if tls == nil {
			// no tls spec
			continue
		}

		if ir.Namespace == secret.Namespace && tls.SecretName == secret.Name {
			return true
		}
		if delegations[ir.Namespace+"/"+secret.Name] {
			if tls.SecretName == secret.Namespace+"/"+secret.Name {
				return true
			}
		}
		if delegations["*/"+secret.Name] {
			if tls.SecretName == secret.Namespace+"/"+secret.Name {
				return true
			}
		}
	}

	for _, proxy := range kc.httpproxies {
		vh := proxy.Spec.VirtualHost
		if vh == nil {
			// not a root ingress
			continue
		}
		tls := vh.TLS
		if tls == nil {
			// no tls spec
			continue
		}

		if proxy.Namespace == secret.Namespace && tls.SecretName == secret.Name {
			return true
		}
		if delegations[proxy.Namespace+"/"+secret.Name] {
			if tls.SecretName == secret.Namespace+"/"+secret.Name {
				return true
			}
		}
		if delegations["*/"+secret.Name] {
			if tls.SecretName == secret.Namespace+"/"+secret.Name {
				return true
			}
		}
	}

	return false
}
