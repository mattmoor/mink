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

package kontext

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

var (
	BaseImageString = "docker.io/mattmoor/kontext-expander:latest"
	// BaseImage is where we publish ./cmd/kontext-expander
	BaseImage, _ = name.ParseReference(BaseImageString)
)

func bundle(ctx context.Context, directory string) (v1.Layer, error) {
	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := filepath.Walk(directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Chase symlinks.
			info, err = os.Stat(path)
			if err != nil {
				return err
			}

			// Compute the path relative to the base path
			relativePath, err := filepath.Rel(directory, path)
			if err != nil {
				return err
			}

			newPath := filepath.Join(StoragePath, relativePath)

			if info.Mode().IsDir() {
				return tw.WriteHeader(&tar.Header{
					Name:     newPath,
					Typeflag: tar.TypeDir,
					Mode:     0555,
				})
			}

			// Open the file to copy it into the tarball.
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// Copy the file into the image tarball.
			if err := tw.WriteHeader(&tar.Header{
				Name:     newPath,
				Size:     info.Size(),
				Typeflag: tar.TypeReg,
				// Use a fixed Mode, so that this isn't sensitive to the directory and umask
				// under which it was created. Additionally, windows can only set 0222,
				// 0444, or 0666, none of which are executable.
				Mode: 0555,
			}); err != nil {
				return err
			}
			_, err = io.Copy(tw, file)
			return err
		})
	if err != nil {
		return nil, err
	}

	return tarball.LayerFromReader(bytes.NewBuffer(buf.Bytes()))
}

type ociThing interface {
	Digest() (v1.Hash, error)
}

type descriptor interface {
	ImageIndex() (v1.ImageIndex, error)
	Image() (v1.Image, error)
}

func appendLayer(mt types.MediaType, baseDesc descriptor, layer v1.Layer) (ociThing, error) {
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

			img, err := mutate.AppendLayers(base, layer)
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

		img, err := mutate.AppendLayers(base, layer)
		if err != nil {
			return nil, err
		}
		return img, nil

	default:
		return nil, fmt.Errorf("Unknown mime type: %v", mt)
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

func Bundle(ctx context.Context, directory string, tag name.Tag) (name.Digest, error) {
	ropt := remote.WithAuthFromKeychain(authn.DefaultKeychain)
	// TODO(mattmoor): We can be more clever here to achieve incrementality,
	// but just yolo package stuff for now.
	mt, baseDesc, err := remoteGet(BaseImage, ropt)
	if err != nil {
		return name.Digest{}, err
	}

	layer, err := bundle(ctx, directory)
	if err != nil {
		return name.Digest{}, err
	}

	oci, err := appendLayer(mt, baseDesc, layer)
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
		return name.Digest{}, fmt.Errorf("Unknown type: %T", oci)
	}

	return name.NewDigest(tag.String() + "@" + hash.String())
}
