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

package git

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/mattmoor/mink/pkg/bundles"
	"github.com/mattmoor/mink/pkg/bundles/kontext"
)

func bundle(ctx context.Context, opts Options) (v1.Layer, error) {
	// Filesystem abstraction based on memory
	fs := memfs.New()
	// Git objects storer based on memory
	storer := memory.NewStorage()

	// Clones the repository into the worktree (fs) and storer all the .git
	// content into the storer
	repo, err := git.CloneContext(ctx, storer, fs, &git.CloneOptions{
		URL:           opts.URL,
		ReferenceName: opts.Ref,
		SingleBranch:  true,
	})
	if err != nil {
		return nil, err
	}
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	// Add an entry for the root kontext directory
	// This is to facilitate testing for compatibility with kontext.
	if err := tw.WriteHeader(&tar.Header{
		Name:     kontext.StoragePath,
		Typeflag: tar.TypeDir,
		Mode:     0555,
	}); err != nil {
		return nil, err
	}

	w := object.NewTreeWalker(tree, true /* recursive */, make(map[plumbing.Hash]bool))
	defer w.Close()
	for {
		name, _, err := w.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}

		if err := func() error {
			// Do not chase symlinks.
			info, err := fs.Lstat(name)
			if err != nil {
				return err
			}

			newPath := filepath.Join(kontext.StoragePath, name)

			// Handle directories
			if info.Mode().IsDir() {
				return tw.WriteHeader(&tar.Header{
					Name:     newPath,
					Typeflag: tar.TypeDir,
					Mode:     0555,
				})
			}

			// Handle symlinks
			if info.Mode()&os.ModeSymlink != 0 {
				linkname, err := fs.Readlink(name)
				if err != nil {
					return err
				}
				return tw.WriteHeader(&tar.Header{
					Name:     newPath,
					Typeflag: tar.TypeSymlink,
					Mode:     0555,
					Linkname: linkname,
				})
			}

			// Open the file to copy it into the tarball.
			file, err := fs.Open(name)
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
		}(); err != nil {
			return nil, fmt.Errorf("error processing %q: %w", name, err)
		}
	}

	return tarball.LayerFromReader(bytes.NewBuffer(buf.Bytes()))
}

// Options contains a collection of options for configuring how things are bundled from git.
type Options struct {
	// URL contains the url from which to clone the git repository.
	URL string
	// Ref contains the ref to check out for bundling.
	Ref plumbing.ReferenceName
}

// Bundle packages up the given git repo as a self-extracting container image based
// on BaseImage and publishes it to tag.
func Bundle(ctx context.Context, opts Options, tag name.Tag) (name.Digest, error) {
	layer, err := bundle(ctx, opts)
	if err != nil {
		return name.Digest{}, err
	}

	return bundles.Map(ctx, kontext.BaseImage, tag, func(ctx context.Context, img v1.Image) (v1.Image, error) {
		return mutate.AppendLayers(img, layer)
	})
}
