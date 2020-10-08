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
	// nolint:gosec // No strong cryptography needed.
	"crypto/sha1"
	"fmt"
	"net/http"
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
	Port            intstr.IntOrString
	RawVisibilities sets.String
	// If the Host header sent to this service needs to be rewritten,
	// then track that so we can send it for probing.
	RewriteHost string

	// TODO(https://github.com/knative-sandbox/net-certmanager/issues/44): Remove this.
	HasPath bool
}

func (si *ServiceInfo) Visibilities() (vis []v1alpha1.IngressVisibility) {
	for _, v := range si.RawVisibilities.List() {
		vis = append(vis, v1alpha1.IngressVisibility(v))
	}
	return
}

func ServiceNames(ctx context.Context, ing *v1alpha1.Ingress) map[string]ServiceInfo {
	s := map[string]ServiceInfo{}
	for _, rule := range ing.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			for _, split := range path.Splits {
				si, ok := s[split.ServiceName]
				if !ok {
					si = ServiceInfo{
						Port:            split.ServicePort,
						RawVisibilities: sets.NewString(),
						HasPath:         path.Path != "",
						RewriteHost:     path.RewriteHost,
					}
				}
				si.RawVisibilities.Insert(string(rule.Visibility))
				s[split.ServiceName] = si
			}
		}
	}
	return s
}

func defaultRetryPolicy() *v1.RetryPolicy {
	return &v1.RetryPolicy{
		NumRetries: 2,
		RetryOn: []v1.RetryOn{
			"cancelled",
			"connect-failure",
			"refused-stream",
			"resource-exhausted",
			"retriable-status-codes",

			// In addition to what Istio specifies (above),
			// also retry connection resets.
			"reset",
		},
		RetriableStatusCodes: []uint32{
			http.StatusServiceUnavailable,
		},
	}
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

			// By default retry on connection problems twice.
			// This matches the default behavior of Istio:
			// https://istio.io/latest/docs/concepts/traffic-management/#retries
			// However, in addition to the codes specified by istio
			retry := defaultRetryPolicy()
			if path.DeprecatedRetries != nil && path.DeprecatedRetries.Attempts > 0 {
				retry.NumRetries = int64(path.DeprecatedRetries.Attempts)

				// When retries is specified explicitly, then we retry some http-level failures as well.
				retry.RetryOn = append(retry.RetryOn, "5xx")

				if path.DeprecatedRetries.PerTryTimeout != nil {
					retry.PerTryTimeout = path.DeprecatedRetries.PerTryTimeout.Duration.String()
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

			var conditions []v1.MatchCondition
			if path.Path != "" {
				conditions = append(conditions, v1.MatchCondition{
					// This is technically not accurate since it's not a prefix,
					// but a regular expression, however, all usage is either empty
					// or absolute paths.
					Prefix: path.Path,
				})
			}
			for header, match := range path.Headers {
				conditions = append(conditions, v1.MatchCondition{
					Header: &v1.HeaderMatchCondition{
						Name:  header,
						Exact: match.Exact,
					},
				})
			}

			if len(conditions) > 1 {
				sort.Slice(conditions, func(i, j int) bool {
					hasPrefixLHS := conditions[i].Prefix != ""
					hasPrefixRHS := conditions[j].Prefix != ""
					if hasPrefixLHS && !hasPrefixRHS {
						return true
					}
					if !hasPrefixLHS && hasPrefixRHS {
						return false
					}
					return conditions[i].Header.Name > conditions[j].Header.Name
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
				// nolint:gosec // No strong cryptography needed.
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
