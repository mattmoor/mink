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

package xds

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/sirupsen/logrus"
)

// Resource represents a source of proto.Messages that can be registered
// for interest.
type Resource interface {
	// Contents returns the contents of this resource.
	Contents() []proto.Message

	// Query returns an entry for each resource name supplied.
	Query(names []string) []proto.Message

	// Register registers ch to receive a value when Notify is called.
	Register(chan int, int, ...string)

	// TypeURL returns the typeURL of messages returned from Values.
	TypeURL() string
}

type grpcStream interface {
	Context() context.Context
	Send(*v2.DiscoveryResponse) error
	Recv() (*v2.DiscoveryRequest, error)
}

// counter holds an atomically incrementing counter.
type counter uint64

func (c *counter) next() uint64 {
	return atomic.AddUint64((*uint64)(c), 1)
}

var connections counter

// NewContourServer creates an internally implemented Server that streams the
// provided set of Resource objects. The returned Server implements the xDS
// State of the World (SotW) variant.
func NewContourServer(log logrus.FieldLogger, resources ...Resource) Server {
	c := contourServer{
		FieldLogger: log,
		resources:   map[string]Resource{},
	}

	for i, r := range resources {
		c.resources[r.TypeURL()] = resources[i]
	}

	return &c
}

type contourServer struct {
	// Since we only implement the streaming state of the world
	// protocol, embed the default null implementations to handle
	// the unimplemented gRPC endpoints.
	discovery.UnimplementedAggregatedDiscoveryServiceServer
	discovery.UnimplementedSecretDiscoveryServiceServer
	v2.UnimplementedRouteDiscoveryServiceServer
	v2.UnimplementedEndpointDiscoveryServiceServer
	v2.UnimplementedClusterDiscoveryServiceServer
	v2.UnimplementedListenerDiscoveryServiceServer

	logrus.FieldLogger
	resources map[string]Resource
}

// stream processes a stream of DiscoveryRequests.
func (s *contourServer) stream(st grpcStream) error {
	// Bump connection counter and set it as a field on the logger.
	log := s.WithField("connection", connections.next())

	// Notify whether the stream terminated on error.
	done := func(log *logrus.Entry, err error) error {
		if err != nil {
			log.WithError(err).Error("stream terminated")
		} else {
			log.Info("stream terminated")
		}

		return err
	}

	ch := make(chan int, 1)

	// internally all registration values start at zero so sending
	// a last that is less than zero will guarantee that each stream
	// will generate a response immediately, then wait.
	last := -1
	ctx := st.Context()

	// now stick in this loop until the client disconnects.
	for {
		// first we wait for the request from Envoy, this is part of
		// the xDS protocol.
		req, err := st.Recv()
		if err != nil {
			return done(log, err)
		}

		// note: redeclare log in this scope so the next time around the loop all is forgotten.
		log := log.WithField("version_info", req.VersionInfo).WithField("response_nonce", req.ResponseNonce)
		if req.Node != nil {
			log = log.WithField("node_id", req.Node.Id).WithField("node_version", fmt.Sprintf("v%d.%d.%d", req.Node.GetUserAgentBuildVersion().Version.MajorNumber, req.Node.GetUserAgentBuildVersion().Version.MinorNumber, req.Node.GetUserAgentBuildVersion().Version.Patch))
		}

		if status := req.ErrorDetail; status != nil {
			// if Envoy rejected the last update log the details here.
			// TODO(dfc) issue 1176: handle xDS ACK/NACK
			log.WithField("code", status.Code).Error(status.Message)
		}

		// from the request we derive the resource to stream which have
		// been registered according to the typeURL.
		r, ok := s.resources[req.TypeUrl]
		if !ok {
			return done(log, fmt.Errorf("no resource registered for typeURL %q", req.TypeUrl))
		}

		log = log.WithField("resource_names", req.ResourceNames).WithField("type_url", req.TypeUrl)
		log.Info("stream_wait")

		// now we wait for a notification, if this is the first request received on this
		// connection last will be less than zero and that will trigger a response immediately.
		r.Register(ch, last, req.ResourceNames...)
		select {
		case last = <-ch:
			// boom, something in the cache has changed.
			// TODO(dfc) the thing that has changed may not be in the scope of the filter
			// so we're going to be sending an update that is a no-op. See #426

			var resources []proto.Message
			switch len(req.ResourceNames) {
			case 0:
				// no resource hints supplied, return the full
				// contents of the resource
				resources = r.Contents()
			default:
				// resource hints supplied, return exactly those
				resources = r.Query(req.ResourceNames)
			}

			any := make([]*any.Any, 0, len(resources))
			for _, r := range resources {
				a, err := ptypes.MarshalAny(r)
				if err != nil {
					return done(log, err)
				}

				any = append(any, a)
			}

			resp := &v2.DiscoveryResponse{
				VersionInfo: strconv.Itoa(last),
				Resources:   any,
				TypeUrl:     r.TypeURL(),
				Nonce:       strconv.Itoa(last),
			}

			if err := st.Send(resp); err != nil {
				return done(log, err)
			}

		case <-ctx.Done():
			return done(log, ctx.Err())
		}
	}
}

func (s *contourServer) StreamClusters(srv v2.ClusterDiscoveryService_StreamClustersServer) error {
	return s.stream(srv)
}

func (s *contourServer) StreamEndpoints(srv v2.EndpointDiscoveryService_StreamEndpointsServer) error {
	return s.stream(srv)
}

func (s *contourServer) StreamListeners(srv v2.ListenerDiscoveryService_StreamListenersServer) error {
	return s.stream(srv)
}

func (s *contourServer) StreamRoutes(srv v2.RouteDiscoveryService_StreamRoutesServer) error {
	return s.stream(srv)
}

func (s *contourServer) StreamSecrets(srv discovery.SecretDiscoveryService_StreamSecretsServer) error {
	return s.stream(srv)
}
