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

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/mattmoor/mink/pkg/bundles"
)

var (
	// BaseImageString holds a reference to a built image of ./cmd/kontext-expander
	// See ./hack/build-flags.sh for how this is replaced at link-time.
	BaseImageString = "ghcr.io/mattmoor/github.com/mattmoor/mink/cmd/kontext-expander:latest"
	// BaseImage is where we publish ./cmd/kontext-expander
	BaseImage, _ = name.ParseReference(BaseImageString)
)

func bundle(directory string) (v1.Layer, error) {
	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := filepath.Walk(directory,
		func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Chase symlinks.
			info, err := os.Stat(path)
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

// Bundle packages up the given directory as a self-extracting container image based
// on BaseImage and publishes it to tag.
func Bundle(ctx context.Context, directory string, tag name.Tag) (name.Digest, error) {
	layer, err := bundle(directory)
	if err != nil {
		return name.Digest{}, err
	}

	return bundles.Map(ctx, BaseImage, tag, func(ctx context.Context, img v1.Image) (v1.Image, error) {
		return mutate.AppendLayers(img, layer)
	})
}
