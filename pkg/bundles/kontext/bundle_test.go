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
	"context"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type descriptorImpl struct {
	i   v1.Image
	ii  v1.ImageIndex
	err error
}

func (di *descriptorImpl) ImageIndex() (v1.ImageIndex, error) {
	return di.ii, di.err
}

func (di *descriptorImpl) Image() (v1.Image, error) {
	return di.i, di.err
}

func TestBundleLayerIndex(t *testing.T) {
	// Check that if we bundle testdata it has the expected size.
	l, err := bundle("./testdata")
	if err != nil {
		t.Error("bundle() =", err)
	}
	sz, err := l.Size()
	if err != nil {
		t.Error("l.Size() =", err)
	}
	if got, want := sz, int64(204); got != want {
		t.Errorf("Size() = %d, wanted %d", got, want)
	}

	// Now overlay it onto a randomly generated image index with 5 variants.
	ii, err := random.Index(3, 4, 5)
	if err != nil {
		t.Error("random.Index() =", err)
	}
	bundle, err := appendLayer(types.OCIImageIndex, &descriptorImpl{ii: ii}, l)
	if err != nil {
		t.Error("appendLayer() =", err)
	}

	// We expect to get back an image index.
	bii, ok := bundle.(v1.ImageIndex)
	if !ok {
		t.Errorf("appendLayer() = %T, wanted v1.ImageIndex", bundle)
	}
	im, err := bii.IndexManifest()
	if err != nil {
		t.Error("IndexManifest() =", err)
	}

	// We expect that image index to have 5 variants (same as above).
	if got, want := len(im.Manifests), 5; got != want {
		t.Errorf("len(im.Manifests) = %d, wanted %d", got, want)
	}
}

func TestBundleLayerImage(t *testing.T) {
	// Check that if we bundle testdata it has the expected size.
	l, err := bundle("./testdata")
	if err != nil {
		t.Error("bundle() =", err)
	}
	sz, err := l.Size()
	if err != nil {
		t.Error("l.Size() =", err)
	}
	if got, want := sz, int64(204); got != want {
		t.Errorf("Size() = %d, wanted %d", got, want)
	}

	// Now overlay it onto a randomly generated image index with 5 variants.
	i, err := random.Image(3, 4)
	if err != nil {
		t.Error("random.Index() =", err)
	}
	bundle, err := appendLayer(types.OCIManifestSchema1, &descriptorImpl{i: i}, l)
	if err != nil {
		t.Error("appendLayer() =", err)
	}

	// We expect to get back an image index.
	_, ok := bundle.(v1.Image)
	if !ok {
		t.Errorf("appendLayer() = %T, wanted v1.Image", bundle)
	}
}

func TestBundleIndex(t *testing.T) {
	remoteGet = func(name.Reference, ...remote.Option) (types.MediaType, descriptor, error) {
		ii, err := random.Index(3, 4, 5)
		return types.OCIImageIndex, &descriptorImpl{ii: ii}, err
	}
	remoteWriteIndex = func(name.Reference, v1.ImageIndex, ...remote.Option) error {
		return nil
	}
	remoteWrite = func(name.Reference, v1.Image, ...remote.Option) error {
		return nil
	}

	tag, _ := name.NewTag("docker.io/blah/blurg")

	_, err := Bundle(context.Background(), "./testdata", tag)
	if err != nil {
		t.Error("Bundle() =", err)
	}
}

func TestBundleImage(t *testing.T) {
	remoteGet = func(name.Reference, ...remote.Option) (types.MediaType, descriptor, error) {
		i, err := random.Image(3, 4)
		return types.OCIManifestSchema1, &descriptorImpl{i: i}, err
	}
	remoteWriteIndex = func(name.Reference, v1.ImageIndex, ...remote.Option) error {
		return nil
	}
	remoteWrite = func(name.Reference, v1.Image, ...remote.Option) error {
		return nil
	}

	tag, _ := name.NewTag("docker.io/blah/blurg")

	_, err := Bundle(context.Background(), "./testdata", tag)
	if err != nil {
		t.Error("Bundle() =", err)
	}
}
