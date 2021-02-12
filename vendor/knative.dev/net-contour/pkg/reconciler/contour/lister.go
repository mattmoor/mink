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
	"context"
	"fmt"
	"net/url"
	"strconv"

	"k8s.io/apimachinery/pkg/util/sets"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"knative.dev/net-contour/pkg/reconciler/contour/config"
	network "knative.dev/networking/pkg"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/networking/pkg/ingress"
	"knative.dev/networking/pkg/status"
)

type lister struct {
	ServiceLister   corev1listers.ServiceLister
	EndpointsLister corev1listers.EndpointsLister
}

var _ status.ProbeTargetLister = (*lister)(nil)

// ListProbeTargets implements status.ProbeTargetLister
func (l *lister) ListProbeTargets(ctx context.Context, ing *v1alpha1.Ingress) ([]status.ProbeTarget, error) {
	var results []status.ProbeTarget

	cfg := config.FromContext(ctx)

	port, scheme := int32(80), "http"
	switch cfg.Network.HTTPProtocol {
	case network.HTTPDisabled, network.HTTPRedirected:
		port, scheme = 443, "https"
	}

	visibilityKeys := cfg.Contour.VisibilityKeys
	for key, hosts := range ingress.HostsPerVisibility(ing, visibilityKeys) {
		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key: %w", err)
		}

		service, err := l.ServiceLister.Services(namespace).Get(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get Service: %w", err)
		}

		endpoints, err := l.EndpointsLister.Endpoints(namespace).Get(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get Endpoints: %w", err)
		}

		urls := make([]*url.URL, 0, hosts.Len())
		for _, host := range hosts.UnsortedList() {
			urls = append(urls, &url.URL{
				Scheme: scheme,
				Host:   host,
			})
		}

		portName, err := network.NameForPortNumber(service, port)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup port %d in %s/%s: %w", port, namespace, name, err)
		}
		for _, sub := range endpoints.Subsets {
			podPort, err := network.PortNumberForName(sub, portName)
			if err != nil {
				return nil, fmt.Errorf("failed to lookup port name %q in endpoints subset for %s/%s: %w",
					portName, namespace, name, err)
			}

			pt := status.ProbeTarget{
				PodIPs:  sets.NewString(),
				Port:    strconv.Itoa(int(port)),
				PodPort: strconv.Itoa(int(podPort)),
				URLs:    urls,
			}
			for _, addr := range sub.Addresses {
				pt.PodIPs.Insert(addr.IP)
			}
			results = append(results, pt)
		}
	}

	return results, nil
}
