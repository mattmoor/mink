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

// Package status holds pieces for handling status updates propagated from
// the DAG back to Kubernetes
package status

import (
	"time"

	contour_api_v1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	"github.com/projectcontour/contour/internal/k8s"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ProxyStatus string

const (
	ProxyStatusValid    ProxyStatus = "valid"
	ProxyStatusInvalid  ProxyStatus = "invalid"
	ProxyStatusOrphaned ProxyStatus = "orphaned"

	OrphanedConditionType ConditionType = "Orphaned"
)

// NewCache creates a new Cache for holding status updates.
func NewCache() Cache {
	return Cache{
		proxyUpdates: make(map[types.NamespacedName]*ProxyUpdate),
		entries:      make(map[string]map[types.NamespacedName]CacheEntry),
	}
}

type CacheEntry interface {
	AsStatusUpdate() k8s.StatusUpdate
	ConditionFor(ConditionType) *contour_api_v1.DetailedCondition
}

// Cache holds status updates from the DAG back towards Kubernetes.
// It holds a per-Kind cache, and is intended to be accessed with a
// KindAccessor.
type Cache struct {
	proxyUpdates map[types.NamespacedName]*ProxyUpdate

	// Map of cache entry maps, keyed on Kind.
	entries map[string]map[types.NamespacedName]CacheEntry
}

// Get returns a pointer to a the cache entry if it exists, nil
// otherwise. The return value is shared between all callers, who
// should take care to cooperate.
func (c *Cache) Get(obj k8s.Object) CacheEntry {
	kind := k8s.KindOf(obj)

	if _, ok := c.entries[kind]; !ok {
		c.entries[kind] = make(map[types.NamespacedName]CacheEntry)
	}

	return c.entries[kind][k8s.NamespacedNameOf(obj)]
}

// Put returns an entry to the cache.
func (c *Cache) Put(obj k8s.Object, e CacheEntry) {
	kind := k8s.KindOf(obj)

	if _, ok := c.entries[kind]; !ok {
		c.entries[kind] = make(map[types.NamespacedName]CacheEntry)
	}

	c.entries[kind][k8s.NamespacedNameOf(obj)] = e
}

// ProxyAccessor returns a ProxyUpdate that allows a client to build up a list of
// errors and warnings to go onto the proxy as conditions, and a function to commit the change
// back to the cache when everything is done.
// The commit function pattern is used so that the ProxyUpdate does not need to know anything
// the cache internals.
func (c *Cache) ProxyAccessor(proxy *contour_api_v1.HTTPProxy) (*ProxyUpdate, func()) {
	pu := &ProxyUpdate{
		Fullname:       k8s.NamespacedNameOf(proxy),
		Generation:     proxy.Generation,
		TransitionTime: v1.NewTime(time.Now()),
		Conditions:     make(map[ConditionType]*contour_api_v1.DetailedCondition),
	}

	return pu, func() {
		c.commitProxy(pu)
	}
}

func (c *Cache) commitProxy(pu *ProxyUpdate) {
	if len(pu.Conditions) == 0 {
		return
	}

	_, ok := c.proxyUpdates[pu.Fullname]
	if ok {
		// When we're committing, if we already have a Valid Condition with an error, and we're trying to
		// set the object back to Valid, skip the commit, as we've visited too far down.
		// If this is removed, the status reporting for when a parent delegates to a child that delegates to itself
		// will not work. Yes, I know, problems everywhere. I'm sorry.
		// TODO(youngnick)#2968: This issue has more details.
		if c.proxyUpdates[pu.Fullname].Conditions[ValidCondition].Status == contour_api_v1.ConditionFalse {
			if pu.Conditions[ValidCondition].Status == contour_api_v1.ConditionTrue {
				return
			}
		}
	}
	c.proxyUpdates[pu.Fullname] = pu
}

// GetStatusUpdates returns a slice of StatusUpdates, ready to be sent off
// to the StatusUpdater by the event handler.
// As more kinds are handled by Cache, we'll update this method.
func (c *Cache) GetStatusUpdates() []k8s.StatusUpdate {
	var flattened []k8s.StatusUpdate

	for fullname, pu := range c.proxyUpdates {
		update := k8s.StatusUpdate{
			NamespacedName: fullname,
			Resource:       contour_api_v1.HTTPProxyGVR,
			Mutator:        pu,
		}

		flattened = append(flattened, update)
	}

	for _, byKind := range c.entries {
		for _, e := range byKind {
			flattened = append(flattened, e.AsStatusUpdate())
		}
	}

	return flattened
}

// GetProxyUpdates gets the underlying ProxyUpdate objects
// from the cache, used by various things (`internal/contour/metrics.go` and `internal/dag/status_test.go`)
// to retrieve info they need.
// TODO(youngnick)#2969: This could conceivably be replaced with a Walk pattern.
func (c *Cache) GetProxyUpdates() []*ProxyUpdate {
	var allUpdates []*ProxyUpdate
	for _, pu := range c.proxyUpdates {
		allUpdates = append(allUpdates, pu)
	}
	return allUpdates
}
