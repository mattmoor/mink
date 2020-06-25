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
	"io"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

var (
	BaseImageString = "docker.io/mattmoor/kontext-expander:latest"
	// BaseImage is where we publish ./cmd/kontext-expander
	BaseImage, _ = name.NewTag(BaseImageString)
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
			if err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return nil, err
	}

	return tarball.LayerFromReader(bytes.NewBuffer(buf.Bytes()))
}

func Bundle(ctx context.Context, directory string, tag name.Tag) (name.Digest, error) {
	// TODO(mattmoor): We can be more clever here to achieve incrementality,
	// but just yolo package stuff for now.
	base, err := remote.Image(BaseImage, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return name.Digest{}, err
	}

	layer, err := bundle(ctx, directory)
	if err != nil {
		return name.Digest{}, err
	}

	result, err := mutate.AppendLayers(base, layer)
	if err != nil {
		return name.Digest{}, err
	}
	hash, err := result.Digest()
	if err != nil {
		return name.Digest{}, err
	}

	if err := remote.Write(tag, result, remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
		return name.Digest{}, err
	}
	return name.NewDigest(tag.String() + "@" + hash.String())
}
