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

package bundles

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type ociThing interface {
	Digest() (v1.Hash, error)
}

type descriptor interface {
	ImageIndex() (v1.ImageIndex, error)
	Image() (v1.Image, error)
}

func doMap(ctx context.Context, mt types.MediaType, baseDesc descriptor, fn Mutator) (ociThing, error) {
	switch mt {
	case types.OCIImageIndex, types.DockerManifestList:
		baseIndex, err := baseDesc.ImageIndex()
		if err != nil {
			return nil, err
		}
		im, err := baseIndex.IndexManifest()
		if err != nil {
			return nil, err
		}

		// Build an image for each child from the base and append it to a new index to produce the result.
		adds := []mutate.IndexAddendum{}
		for _, desc := range im.Manifests {
			base, err := baseIndex.Image(desc.Digest)
			if err != nil {
				return nil, err
			}

			img, err := fn(ctx, base)
			if err != nil {
				return nil, err
			}

			adds = append(adds, mutate.IndexAddendum{
				Add: img,
				Descriptor: v1.Descriptor{
					URLs:        desc.URLs,
					MediaType:   desc.MediaType,
					Annotations: desc.Annotations,
					Platform:    desc.Platform,
				},
			})
		}

		// Construct the image index.
		return mutate.IndexMediaType(mutate.AppendManifests(empty.Index, adds...), mt), nil

	case types.OCIManifestSchema1, types.DockerManifestSchema2:
		base, err := baseDesc.Image()
		if err != nil {
			return nil, err
		}

		img, err := fn(ctx, base)
		if err != nil {
			return nil, err
		}
		return img, nil

	default:
		return nil, fmt.Errorf("unknown mime type: %v", mt)
	}
}

// These exist for the purpose of TESTING
var (
	remoteGet = func(ref name.Reference, opts ...remote.Option) (types.MediaType, descriptor, error) {
		d, err := remote.Get(ref, opts...)
		if err != nil {
			return "", nil, err
		}
		return d.MediaType, d, nil
	}
	remoteWriteIndex = remote.WriteIndex
	remoteWrite      = remote.Write
)

// TestWithFakes temporarily switches the remote writes with fakes for testing
// ideally we'd have a nicer way to do this with a fake container registry host or something
func TestWithFakes(fn func() error) error {
	oldWriteIndex := remoteWriteIndex
	oldWrite := remoteWrite

	remoteWriteIndex = func(ref name.Reference, ii v1.ImageIndex, options ...remote.Option) error {
		fmt.Printf("faked writing an index\n")
		return nil
	}

	remoteWrite = func(ref name.Reference, img v1.Image, options ...remote.Option) error {
		fmt.Printf("faked writing\n")
		return nil
	}

	err := fn()
	remoteWriteIndex = oldWriteIndex
	remoteWrite = oldWrite
	return err
}

// Mutator is the signature of the callback supplied to Map.  This function will be called on each of the
// images that comprise the referenced base, and the function maybe be called once (if an image) or many
// times (is an image index).
type Mutator func(ctx context.Context, img v1.Image) (v1.Image, error)

// Map loads the base reference, applies the Mutator function to all of the
// images contained within, and publishes it to tag.  The digest of the resulting
// image or image index is returned upon success, or an error on failure.
func Map(ctx context.Context, base name.Reference, tag name.Tag, fn Mutator) (name.Digest, error) {
	auth, err := authn.DefaultKeychain.Resolve(base.Context())
	if err != nil {
		return name.Digest{}, err
	}
	ropt := remote.WithAuth(auth)

	// TODO(mattmoor): We can be more clever here to achieve incrementality,
	// but just yolo package stuff for now.
	mt, baseDesc, err := remoteGet(base, ropt)
	if err != nil {
		return name.Digest{}, err
	}

	// If it is going to a different registry, switch auth.
	// Don't do this unconditionally as resolution is ~400ms.
	if tag.RegistryStr() != base.Context().RegistryStr() {
		auth, err := authn.DefaultKeychain.Resolve(tag)
		if err != nil {
			return name.Digest{}, err
		}
		ropt = remote.WithAuth(auth)
	}

	oci, err := doMap(ctx, mt, baseDesc, fn)
	if err != nil {
		return name.Digest{}, err
	}

	hash, err := oci.Digest()
	if err != nil {
		return name.Digest{}, err
	}

	switch oci := oci.(type) {
	case v1.ImageIndex:
		if err := remoteWriteIndex(tag, oci, ropt); err != nil {
			return name.Digest{}, err
		}
	case v1.Image:
		if err := remoteWrite(tag, oci, ropt); err != nil {
			return name.Digest{}, err
		}
	default:
		return name.Digest{}, fmt.Errorf("unknown type: %T", oci)
	}

	return name.NewDigest(tag.String() + "@" + hash.String())
}
