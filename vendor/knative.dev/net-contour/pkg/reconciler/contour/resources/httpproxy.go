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

package resources

import (
	"context"
	"crypto/sha1"
	"fmt"
	"sort"
	"strings"

	v1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/net-contour/pkg/reconciler/contour/config"
	networkingpkg "knative.dev/networking/pkg"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/networking/pkg/ingress"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/network"
	"knative.dev/pkg/ptr"
)

type ServiceInfo struct {
	Port         intstr.IntOrString
	Visibilities []v1alpha1.IngressVisibility

	// TODO(https://github.com/knative-sandbox/net-certmanager/issues/44): Remove this.
	HasPath bool
}

func ServiceNames(ctx context.Context, ing *v1alpha1.Ingress) map[string]ServiceInfo {
	// Build it up using string sets to deduplicate.
	s := map[string]sets.String{}
	p := map[string]intstr.IntOrString{}
	paths := sets.NewString()
	for _, rule := range ing.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			for _, split := range path.Splits {
				set, ok := s[split.ServiceName]
				if !ok {
					set = sets.NewString()
				}
				set.Insert(string(rule.Visibility))
				s[split.ServiceName] = set
				p[split.ServiceName] = split.ServicePort
				if path.Path != "" {
					paths.Insert(split.ServiceName)
				}
			}
		}
	}

	// Then iterate over the map to give the return value the right type.
	s2 := map[string]ServiceInfo{}
	for name, vis := range s {
		visibilities := make([]v1alpha1.IngressVisibility, 0, len(vis))
		for _, v := range vis.List() {
			visibilities = append(visibilities, v1alpha1.IngressVisibility(v))
		}
		s2[name] = ServiceInfo{
			Port:         p[name],
			Visibilities: visibilities,
			HasPath:      paths.Has(name),
		}
	}
	return s2
}

func MakeHTTPProxies(ctx context.Context, ing *v1alpha1.Ingress, serviceToProtocol map[string]string) []*v1.HTTPProxy {
	ing = ing.DeepCopy()
	ingress.InsertProbe(ing)

	hostToTLS := make(map[string]*v1alpha1.IngressTLS, len(ing.Spec.TLS))
	for _, tls := range ing.Spec.TLS {
		for _, host := range tls.Hosts {
			t := tls
			hostToTLS[host] = &t
		}
	}

	var allowInsecure bool
	switch config.FromContext(ctx).Network.HTTPProtocol {
	case networkingpkg.HTTPDisabled, networkingpkg.HTTPRedirected:
		allowInsecure = false
	case networkingpkg.HTTPEnabled:
		allowInsecure = true
	}

	proxies := []*v1.HTTPProxy{}
	for _, rule := range ing.Spec.Rules {
		class := config.FromContext(ctx).Contour.VisibilityClasses[rule.Visibility]

		routes := make([]v1.Route, 0, len(rule.HTTP.Paths))
		for _, path := range rule.HTTP.Paths {
			var top *v1.TimeoutPolicy
			if path.Timeout != nil {
				top = &v1.TimeoutPolicy{
					Response: path.Timeout.Duration.String(),
				}
			} else {
				top = &v1.TimeoutPolicy{
					Response: "infinity",
				}
			}

			var retry *v1.RetryPolicy
			if path.Retries != nil && path.Retries.Attempts > 0 {
				retry = &v1.RetryPolicy{
					NumRetries: int64(path.Retries.Attempts),
				}
				if path.Retries.PerTryTimeout != nil {
					retry.PerTryTimeout = path.Retries.PerTryTimeout.Duration.String()
				}

			}

			preSplitHeaders := &v1.HeadersPolicy{
				Set: make([]v1.HeaderValue, 0, len(path.AppendHeaders)),
			}
			for key, value := range path.AppendHeaders {
				preSplitHeaders.Set = append(preSplitHeaders.Set, v1.HeaderValue{
					Name:  key,
					Value: value,
				})
			}

			if path.RewriteHost != "" {
				preSplitHeaders.Set = append(preSplitHeaders.Set, v1.HeaderValue{
					Name:  "Host",
					Value: path.RewriteHost,
				})
			}

			// This should never be empty due to the InsertProbe
			sort.Slice(preSplitHeaders.Set, func(i, j int) bool {
				return preSplitHeaders.Set[i].Name < preSplitHeaders.Set[j].Name
			})

			svcs := make([]v1.Service, 0, len(path.Splits))
			for _, split := range path.Splits {
				postSplitHeaders := &v1.HeadersPolicy{
					Set: make([]v1.HeaderValue, 0, len(split.AppendHeaders)),
				}
				for key, value := range split.AppendHeaders {
					postSplitHeaders.Set = append(postSplitHeaders.Set, v1.HeaderValue{
						Name:  key,
						Value: value,
					})
				}
				if len(postSplitHeaders.Set) > 0 {
					sort.Slice(postSplitHeaders.Set, func(i, j int) bool {
						return postSplitHeaders.Set[i].Name < postSplitHeaders.Set[j].Name
					})
				} else {
					postSplitHeaders = nil
				}
				var protocol *string
				if proto, ok := serviceToProtocol[split.ServiceName]; ok {
					protocol = ptr.String(proto)
				}
				svcs = append(svcs, v1.Service{
					Name:                 split.ServiceName,
					Port:                 split.ServicePort.IntValue(),
					Weight:               int64(split.Percent),
					RequestHeadersPolicy: postSplitHeaders,
					Protocol:             protocol,
				})
			}

			var conditions []v1.Condition
			if path.Path != "" {
				conditions = append(conditions, v1.Condition{
					// This is technically not accurate since it's not a prefix,
					// but a regular expression, however, all usage is either empty
					// or absolute paths.
					Prefix: path.Path,
				})
			}
			for header, match := range path.Headers {
				conditions = append(conditions, v1.Condition{
					Header: &v1.HeaderCondition{
						Name:  header,
						Exact: match.Exact,
					},
				})
			}

			routes = append(routes, v1.Route{
				Conditions:           conditions,
				TimeoutPolicy:        top,
				RetryPolicy:          retry,
				Services:             svcs,
				EnableWebsockets:     true,
				RequestHeadersPolicy: preSplitHeaders,
				PermitInsecure:       allowInsecure,
			})
		}

		base := v1.HTTPProxy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ing.Namespace,
				Labels: map[string]string{
					GenerationKey: fmt.Sprintf("%d", ing.Generation),
					ParentKey:     ing.Name,
					ClassKey:      class,
				},
				Annotations: map[string]string{
					ClassKey: class,
				},
				OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(ing)},
			},
			Spec: v1.HTTPProxySpec{
				// VirtualHost: filled in below
				Routes: routes,
			},
		}

		for _, originalHost := range rule.Hosts {
			for _, host := range ingress.ExpandedHosts(sets.NewString(originalHost)).List() {
				hostProxy := base.DeepCopy()

				class := class

				// Ideally these would just be marked ClusterLocal :(
				if strings.HasSuffix(originalHost, network.GetClusterDomainName()) {
					class = config.FromContext(ctx).Contour.VisibilityClasses[v1alpha1.IngressVisibilityClusterLocal]
					hostProxy.Annotations[ClassKey] = class
					hostProxy.Labels[ClassKey] = class
				}

				hostProxy.Name = kmeta.ChildName(ing.Name+"-"+class+"-", host)
				hostProxy.Spec.VirtualHost = &v1.VirtualHost{
					Fqdn: host,
				}
				hostProxy.Labels[DomainHashKey] = fmt.Sprintf("%x", sha1.Sum([]byte(host)))

				if tls, ok := hostToTLS[host]; ok {
					// TODO(mattmoor): How do we deal with custom secret schemas?
					hostProxy.Spec.VirtualHost.TLS = &v1.TLS{
						SecretName: fmt.Sprintf("%s/%s", tls.SecretNamespace, tls.SecretName),
					}
				} else if s := config.FromContext(ctx).Contour.DefaultTLSSecret; s != nil {
					hostProxy.Spec.VirtualHost.TLS = &v1.TLS{SecretName: s.String()}
				}

				proxies = append(proxies, hostProxy)
			}
		}
	}

	return proxies
}
