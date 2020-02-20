/*
Copyright 2019 The Tekton Authors

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

package pod

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	lru "github.com/hashicorp/golang-lru"
	"k8s.io/client-go/kubernetes"
)

const cacheSize = 1024

type entrypointCache struct {
	kubeclient kubernetes.Interface
	lru        *lru.Cache // cache of digest string -> image entrypoint []string
}

// NewEntrypointCache returns a new entrypoint cache implementation that uses
// K8s credentials to pull image metadata from a container image registry.
func NewEntrypointCache(kubeclient kubernetes.Interface) (EntrypointCache, error) {
	lru, err := lru.New(cacheSize)
	if err != nil {
		return nil, err
	}
	return &entrypointCache{
		kubeclient: kubeclient,
		lru:        lru,
	}, nil
}

func (e *entrypointCache) Get(ref name.Reference, namespace, serviceAccountName string) (v1.Image, error) {
	// If image is specified by digest, check the local cache.
	if digest, ok := ref.(name.Digest); ok {
		if img, ok := e.lru.Get(digest.String()); ok {
			return img.(v1.Image), nil
		}
	}

	// If the image wasn't specified by digest, or if the entrypoint
	// wasn't found, we have to consult the remote registry, using
	// imagePullSecrets.
	kc, err := k8schain.New(e.kubeclient, k8schain.Options{
		Namespace:          namespace,
		ServiceAccountName: serviceAccountName,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating k8schain: %v", err)
	}
	mkc := authn.NewMultiKeychain(kc)
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(mkc))
	if err != nil {
		return nil, fmt.Errorf("error getting image manifest: %v", err)
	}
	return img, nil
}

func (e *entrypointCache) Set(d name.Digest, img v1.Image) { e.lru.Add(d.String(), img) }
