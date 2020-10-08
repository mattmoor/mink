// Copyright Project Contour Authors
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

package v2

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_v2_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_api_v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_api_v2_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	envoy_config_filter_http_ext_authz_v2 "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/ext_authz/v2"
	lua "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/lua/v2"
	http "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/projectcontour/contour/internal/dag"
	"github.com/projectcontour/contour/internal/envoy"
	"github.com/projectcontour/contour/internal/protobuf"
	"github.com/projectcontour/contour/internal/sorter"
	"github.com/projectcontour/contour/internal/timeout"
)

type HTTPVersionType = http.HttpConnectionManager_CodecType

const (
	HTTPVersionAuto HTTPVersionType = http.HttpConnectionManager_AUTO
	HTTPVersion1    HTTPVersionType = http.HttpConnectionManager_HTTP1
	HTTPVersion2    HTTPVersionType = http.HttpConnectionManager_HTTP2
	HTTPVersion3    HTTPVersionType = http.HttpConnectionManager_HTTP3
)

// ProtoNamesForVersions returns the slice of ALPN protocol names for the give HTTP versions.
func ProtoNamesForVersions(versions ...HTTPVersionType) []string {
	protocols := map[HTTPVersionType]string{
		HTTPVersion1: "http/1.1",
		HTTPVersion2: "h2",
		HTTPVersion3: "",
	}
	defaultVersions := []string{"h2", "http/1.1"}
	wantedVersions := map[HTTPVersionType]struct{}{}

	if versions == nil {
		return defaultVersions
	}

	for _, v := range versions {
		wantedVersions[v] = struct{}{}
	}

	var alpn []string

	// Check for versions in preference order.
	for _, v := range []HTTPVersionType{HTTPVersionAuto, HTTPVersion2, HTTPVersion1} {
		if _, ok := wantedVersions[v]; ok {
			if v == HTTPVersionAuto {
				return defaultVersions
			}

			log.Printf("wanted %d -> %s", v, protocols[v])
			alpn = append(alpn, protocols[v])
		}
	}

	return alpn
}

// CodecForVersions determines a single Envoy HTTP codec constant
// that support all the given HTTP protocol versions.
func CodecForVersions(versions ...HTTPVersionType) HTTPVersionType {
	switch len(versions) {
	case 1:
		return versions[0]
	case 0:
		// Default is to autodetect.
		return HTTPVersionAuto
	default:
		// If more than one version is allowed, autodetect and let ALPN sort it out.
		return HTTPVersionAuto
	}
}

// TLSInspector returns a new TLS inspector listener filter.
func TLSInspector() *envoy_api_v2_listener.ListenerFilter {
	return &envoy_api_v2_listener.ListenerFilter{
		Name: wellknown.TlsInspector,
	}
}

// ProxyProtocol returns a new Proxy Protocol listener filter.
func ProxyProtocol() *envoy_api_v2_listener.ListenerFilter {
	return &envoy_api_v2_listener.ListenerFilter{
		Name: wellknown.ProxyProtocol,
	}
}

// Listener returns a new v2.Listener for the supplied address, port, and filters.
func Listener(name, address string, port int, lf []*envoy_api_v2_listener.ListenerFilter, filters ...*envoy_api_v2_listener.Filter) *v2.Listener {
	l := &v2.Listener{
		Name:            name,
		Address:         SocketAddress(address, port),
		ListenerFilters: lf,
		SocketOptions:   TCPKeepaliveSocketOptions(),
	}
	if len(filters) > 0 {
		l.FilterChains = append(
			l.FilterChains,
			&envoy_api_v2_listener.FilterChain{
				Filters: filters,
			},
		)
	}
	return l
}

type httpConnectionManagerBuilder struct {
	routeConfigName               string
	metricsPrefix                 string
	accessLoggers                 []*accesslog.AccessLog
	requestTimeout                timeout.Setting
	connectionIdleTimeout         timeout.Setting
	streamIdleTimeout             timeout.Setting
	maxConnectionDuration         timeout.Setting
	connectionShutdownGracePeriod timeout.Setting
	filters                       []*http.HttpFilter
	codec                         HTTPVersionType // Note the zero value is AUTO, which is the default we want.
}

// RouteConfigName sets the name of the RDS element that contains
// the routing table for this manager.
func (b *httpConnectionManagerBuilder) RouteConfigName(name string) *httpConnectionManagerBuilder {
	b.routeConfigName = name
	return b
}

// MetricsPrefix sets the prefix used for emitting metrics from the
// connection manager. Note that this prefix is externally visible in
// monitoring tools, so it is subject to compatibility concerns.
//
// See https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/stats#config-http-conn-man-stats
func (b *httpConnectionManagerBuilder) MetricsPrefix(prefix string) *httpConnectionManagerBuilder {
	b.metricsPrefix = prefix
	return b
}

// Codec sets the HTTP codec for the manager. The default is AUTO.
func (b *httpConnectionManagerBuilder) Codec(codecType HTTPVersionType) *httpConnectionManagerBuilder {
	b.codec = codecType
	return b
}

// AccessLoggers sets the access logging configuration.
func (b *httpConnectionManagerBuilder) AccessLoggers(loggers []*accesslog.AccessLog) *httpConnectionManagerBuilder {
	b.accessLoggers = loggers
	return b
}

// RequestTimeout sets the active request timeout on the connection manager.
func (b *httpConnectionManagerBuilder) RequestTimeout(timeout timeout.Setting) *httpConnectionManagerBuilder {
	b.requestTimeout = timeout
	return b
}

// ConnectionIdleTimeout sets the idle timeout on the connection manager.
func (b *httpConnectionManagerBuilder) ConnectionIdleTimeout(timeout timeout.Setting) *httpConnectionManagerBuilder {
	b.connectionIdleTimeout = timeout
	return b
}

// StreamIdleTimeout sets the stream idle timeout on the connection manager.
func (b *httpConnectionManagerBuilder) StreamIdleTimeout(timeout timeout.Setting) *httpConnectionManagerBuilder {
	b.streamIdleTimeout = timeout
	return b
}

// MaxConnectionDuration sets the max connection duration on the connection manager.
func (b *httpConnectionManagerBuilder) MaxConnectionDuration(timeout timeout.Setting) *httpConnectionManagerBuilder {
	b.maxConnectionDuration = timeout
	return b
}

// ConnectionShutdownGracePeriod sets the drain timeout on the connection manager.
func (b *httpConnectionManagerBuilder) ConnectionShutdownGracePeriod(timeout timeout.Setting) *httpConnectionManagerBuilder {
	b.connectionShutdownGracePeriod = timeout
	return b
}

func (b *httpConnectionManagerBuilder) DefaultFilters() *httpConnectionManagerBuilder {
	b.filters = append(b.filters,
		&http.HttpFilter{
			Name: wellknown.Gzip,
		},
		&http.HttpFilter{
			Name: wellknown.GRPCWeb,
		},
		&http.HttpFilter{
			Name: wellknown.CORS,
		},
		&http.HttpFilter{
			Name: wellknown.Router,
		},
	)

	return b
}

// AddFilter appends f to the list of filters for this HTTPConnectionManager. f
// may by nil, in which case it is ignored.
func (b *httpConnectionManagerBuilder) AddFilter(f *http.HttpFilter) *httpConnectionManagerBuilder {
	if f == nil {
		return b
	}

	if len(b.filters) > 0 {
		lastIndex := len(b.filters) - 1
		// If the router filter is last, keep it at the end
		// of the filter chain when we add the new filter.
		if b.filters[lastIndex].Name == wellknown.Router {
			b.filters = append(b.filters[:lastIndex], f, b.filters[lastIndex])
			return b
		}
	}

	b.filters = append(b.filters, f)

	return b
}

// Validate runs builtin validation rules against the current builder state.
func (b *httpConnectionManagerBuilder) Validate() error {
	filterNames := map[string]struct{}{}

	for _, f := range b.filters {
		filterNames[f.Name] = struct{}{}
	}

	// If there's no router filter, requests won't be forwarded.
	if _, ok := filterNames[wellknown.Router]; !ok {
		return fmt.Errorf("missing required filter %q", wellknown.Router)
	}

	return nil
}

// Get returns a new http.HttpConnectionManager filter, constructed
// from the builder settings.
//
// See https://www.envoyproxy.io/docs/envoy/latest/api-v2/config/filter/network/http_connection_manager/v2/http_connection_manager.proto.html
func (b *httpConnectionManagerBuilder) Get() *envoy_api_v2_listener.Filter {
	// For now, failing validation is a programmer error that
	// the caller can't reasonably recover from. A caller that can
	// handle this should validate manually.
	if err := b.Validate(); err != nil {
		panic(err.Error())
	}

	cm := &http.HttpConnectionManager{
		CodecType: b.codec,
		RouteSpecifier: &http.HttpConnectionManager_Rds{
			Rds: &http.Rds{
				RouteConfigName: b.routeConfigName,
				ConfigSource:    ConfigSource("contour"),
			},
		},
		HttpFilters: b.filters,
		CommonHttpProtocolOptions: &envoy_api_v2_core.HttpProtocolOptions{
			IdleTimeout: envoy.Timeout(b.connectionIdleTimeout),
		},
		HttpProtocolOptions: &envoy_api_v2_core.Http1ProtocolOptions{
			// Enable support for HTTP/1.0 requests that carry
			// a Host: header. See #537.
			AcceptHttp_10: true,
		},
		UseRemoteAddress: protobuf.Bool(true),
		NormalizePath:    protobuf.Bool(true),

		// issue #1487 pass through X-Request-Id if provided.
		PreserveExternalRequestId: true,
		MergeSlashes:              true,

		RequestTimeout:    envoy.Timeout(b.requestTimeout),
		StreamIdleTimeout: envoy.Timeout(b.streamIdleTimeout),
		DrainTimeout:      envoy.Timeout(b.connectionShutdownGracePeriod),
	}

	// Max connection duration is infinite/disabled by default in Envoy, so if the timeout setting
	// indicates to either disable or use default, don't pass a value at all. Note that unlike other
	// Envoy timeouts, explicitly passing a 0 here *would not* disable the timeout; it needs to be
	// omitted entirely.
	if !b.maxConnectionDuration.IsDisabled() && !b.maxConnectionDuration.UseDefault() {
		cm.CommonHttpProtocolOptions.MaxConnectionDuration = protobuf.Duration(b.maxConnectionDuration.Duration())
	}

	if len(b.accessLoggers) > 0 {
		cm.AccessLog = b.accessLoggers
	}

	// If there's no explicit metrics prefix, default it to the
	// route config name.
	if b.metricsPrefix != "" {
		cm.StatPrefix = b.metricsPrefix
	} else {
		cm.StatPrefix = b.routeConfigName
	}

	return &envoy_api_v2_listener.Filter{
		Name: wellknown.HTTPConnectionManager,
		ConfigType: &envoy_api_v2_listener.Filter_TypedConfig{
			TypedConfig: protobuf.MustMarshalAny(cm),
		},
	}
}

// HTTPConnectionManager creates a new HTTP Connection Manager filter
// for the supplied route, access log, and client request timeout.
func HTTPConnectionManager(routename string, accesslogger []*accesslog.AccessLog, requestTimeout time.Duration) *envoy_api_v2_listener.Filter {
	return HTTPConnectionManagerBuilder().
		RouteConfigName(routename).
		MetricsPrefix(routename).
		AccessLoggers(accesslogger).
		RequestTimeout(timeout.DurationSetting(requestTimeout)).
		DefaultFilters().
		Get()
}

func HTTPConnectionManagerBuilder() *httpConnectionManagerBuilder {
	return &httpConnectionManagerBuilder{}
}

// TCPProxy creates a new TCPProxy filter.
func TCPProxy(statPrefix string, proxy *dag.TCPProxy, accesslogger []*accesslog.AccessLog) *envoy_api_v2_listener.Filter {
	// Set the idle timeout in seconds for connections through a TCP Proxy type filter.
	// The value of two and a half hours for reasons documented at
	// https://github.com/projectcontour/contour/issues/1074
	// Set to 9001 because now it's OVER NINE THOUSAND.
	idleTimeout := protobuf.Duration(9001 * time.Second)

	switch len(proxy.Clusters) {
	case 1:
		return &envoy_api_v2_listener.Filter{
			Name: wellknown.TCPProxy,
			ConfigType: &envoy_api_v2_listener.Filter_TypedConfig{
				TypedConfig: protobuf.MustMarshalAny(&tcp.TcpProxy{
					StatPrefix: statPrefix,
					ClusterSpecifier: &tcp.TcpProxy_Cluster{
						Cluster: envoy.Clustername(proxy.Clusters[0]),
					},
					AccessLog:   accesslogger,
					IdleTimeout: idleTimeout,
				}),
			},
		}
	default:
		var clusters []*tcp.TcpProxy_WeightedCluster_ClusterWeight
		for _, c := range proxy.Clusters {
			weight := c.Weight
			if weight == 0 {
				weight = 1
			}
			clusters = append(clusters, &tcp.TcpProxy_WeightedCluster_ClusterWeight{
				Name:   envoy.Clustername(c),
				Weight: weight,
			})
		}
		sort.Stable(sorter.For(clusters))
		return &envoy_api_v2_listener.Filter{
			Name: wellknown.TCPProxy,
			ConfigType: &envoy_api_v2_listener.Filter_TypedConfig{
				TypedConfig: protobuf.MustMarshalAny(&tcp.TcpProxy{
					StatPrefix: statPrefix,
					ClusterSpecifier: &tcp.TcpProxy_WeightedClusters{
						WeightedClusters: &tcp.TcpProxy_WeightedCluster{
							Clusters: clusters,
						},
					},
					AccessLog:   accesslogger,
					IdleTimeout: idleTimeout,
				}),
			},
		}
	}
}

// SocketAddress creates a new TCP envoy_api_v2_core.Address.
func SocketAddress(address string, port int) *envoy_api_v2_core.Address {
	if address == "::" {
		return &envoy_api_v2_core.Address{
			Address: &envoy_api_v2_core.Address_SocketAddress{
				SocketAddress: &envoy_api_v2_core.SocketAddress{
					Protocol:   envoy_api_v2_core.SocketAddress_TCP,
					Address:    address,
					Ipv4Compat: true,
					PortSpecifier: &envoy_api_v2_core.SocketAddress_PortValue{
						PortValue: uint32(port),
					},
				},
			},
		}
	}
	return &envoy_api_v2_core.Address{
		Address: &envoy_api_v2_core.Address_SocketAddress{
			SocketAddress: &envoy_api_v2_core.SocketAddress{
				Protocol: envoy_api_v2_core.SocketAddress_TCP,
				Address:  address,
				PortSpecifier: &envoy_api_v2_core.SocketAddress_PortValue{
					PortValue: uint32(port),
				},
			},
		},
	}
}

// Filters returns a []*envoy_api_v2_listener.Filter for the supplied filters.
func Filters(filters ...*envoy_api_v2_listener.Filter) []*envoy_api_v2_listener.Filter {
	if len(filters) == 0 {
		return nil
	}
	return filters
}

// FilterChain retruns a *envoy_api_v2_listener.FilterChain for the supplied filters.
func FilterChain(filters ...*envoy_api_v2_listener.Filter) *envoy_api_v2_listener.FilterChain {
	return &envoy_api_v2_listener.FilterChain{
		Filters: filters,
	}
}

// FilterChains returns a []*envoy_api_v2_listener.FilterChain for the supplied filters.
func FilterChains(filters ...*envoy_api_v2_listener.Filter) []*envoy_api_v2_listener.FilterChain {
	if len(filters) == 0 {
		return nil
	}
	return []*envoy_api_v2_listener.FilterChain{
		FilterChain(filters...),
	}
}

func FilterMisdirectedRequests(fqdn string) *http.HttpFilter {
	// When Envoy matches on the virtual host domain, we configure
	// it to match any port specifier (see envoy.VirtualHost),
	// so the Host header (authority) may contain a port that
	// should be ignored. This means that if we don't have a match,
	// we should try again after stripping the port specifier.

	code := `
function envoy_on_request(request_handle)
	local headers = request_handle:headers()
	local host = string.lower(headers:get(":authority"))
	local target = "%s"

	if host ~= target then
		s, e = string.find(host, ":", 1, true)
		if s ~= nil then
			host = string.sub(host, 1, s - 1)
		end

		if host ~= target then
			request_handle:respond(
				{[":status"] = "421"},
				string.format("misdirected request to %%q", headers:get(":authority"))
			)
		end
	end
end
	`

	return &http.HttpFilter{
		Name: "envoy.filters.http.lua",
		ConfigType: &http.HttpFilter_TypedConfig{
			TypedConfig: protobuf.MustMarshalAny(&lua.Lua{
				InlineCode: fmt.Sprintf(code, strings.ToLower(fqdn)),
			}),
		},
	}
}

// FilterExternalAuthz returns an `ext_authz` filter configured with the
// requested parameters.
func FilterExternalAuthz(authzClusterName string, failOpen bool, timeout timeout.Setting) *http.HttpFilter {
	authConfig := envoy_config_filter_http_ext_authz_v2.ExtAuthz{
		Services: &envoy_config_filter_http_ext_authz_v2.ExtAuthz_GrpcService{
			GrpcService: &envoy_api_v2_core.GrpcService{
				TargetSpecifier: &envoy_api_v2_core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &envoy_api_v2_core.GrpcService_EnvoyGrpc{
						ClusterName: authzClusterName,
					},
				},
				Timeout: envoy.Timeout(timeout),
				// We don't need to configure metadata here, since we allow
				// operators to specify authorization context parameters at
				// the virtual host and route.
				InitialMetadata: []*envoy_api_v2_core.HeaderValue{},
			},
		},
		// Pretty sure we always want this. Why have an
		// external auth service if it is not going to affect
		// routing decisions?
		ClearRouteCache:  true,
		FailureModeAllow: failOpen,
		StatusOnError: &envoy_type.HttpStatus{
			Code: envoy_type.StatusCode_Forbidden,
		},
		MetadataContextNamespaces: []string{},
		IncludePeerCertificate:    true,
	}

	// TODO(jpeach): When we move to the Envoy v3 API, propagate the
	// `transport_api_version` from ExtensionServiceSpec ProtocolVersion.

	return &http.HttpFilter{
		Name: "envoy.filters.http.ext_authz",
		ConfigType: &http.HttpFilter_TypedConfig{
			TypedConfig: protobuf.MustMarshalAny(&authConfig),
		},
	}
}

// FilterChainTLS returns a TLS enabled envoy_api_v2_listener.FilterChain.
func FilterChainTLS(domain string, downstream *envoy_api_v2_auth.DownstreamTlsContext, filters []*envoy_api_v2_listener.Filter) *envoy_api_v2_listener.FilterChain {
	fc := &envoy_api_v2_listener.FilterChain{
		Filters: filters,
		FilterChainMatch: &envoy_api_v2_listener.FilterChainMatch{
			ServerNames: []string{domain},
		},
	}
	// Attach TLS data to this listener if provided.
	if downstream != nil {
		fc.TransportSocket = DownstreamTLSTransportSocket(downstream)

	}
	return fc
}

// FilterChainTLSFallback returns a TLS enabled envoy_api_v2_listener.FilterChain conifgured for FallbackCertificate.
func FilterChainTLSFallback(downstream *envoy_api_v2_auth.DownstreamTlsContext, filters []*envoy_api_v2_listener.Filter) *envoy_api_v2_listener.FilterChain {
	fc := &envoy_api_v2_listener.FilterChain{
		Name:    "fallback-certificate",
		Filters: filters,
		FilterChainMatch: &envoy_api_v2_listener.FilterChainMatch{
			TransportProtocol: "tls",
		},
	}
	// Attach TLS data to this listener if provided.
	if downstream != nil {
		fc.TransportSocket = DownstreamTLSTransportSocket(downstream)
	}
	return fc
}

// ListenerFilters returns a []*envoy_api_v2_listener.ListenerFilter for the supplied listener filters.
func ListenerFilters(filters ...*envoy_api_v2_listener.ListenerFilter) []*envoy_api_v2_listener.ListenerFilter {
	return filters
}

func ContainsFallbackFilterChain(filterchains []*envoy_api_v2_listener.FilterChain) bool {
	for _, fc := range filterchains {
		if fc.Name == "fallback-certificate" {
			return true
		}
	}
	return false
}
