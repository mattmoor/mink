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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestExpand(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	// "expand" testdata into a new temporary directory.
	src := filepath.Join(wd, "testdata")
	dest, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("ioutil.TempDir() =", err)
	}
	defer os.RemoveAll(dest)
	if err := os.Chdir(dest); err != nil {
		t.Fatal("os.Chdir() =", err)
	}
	if err := expand(context.Background(), src); err != nil {
		t.Error("expand() =", err)
	}

	// bundle up both directories.
	lSrc, err := bundle(src)
	if err != nil {
		t.Error("bundle() =", err)
	}
	lDest, err := bundle(dest)
	if err != nil {
		t.Error("bundle() =", err)
	}

	// Compute the bundle hashes
	hSrc, err := lSrc.Digest()
	if err != nil {
		t.Error("lSrc.Digest() =", err)
	}
	hDest, err := lDest.Digest()
	if err != nil {
		t.Error("lDest.Digest() =", err)
	}

	// Make sure they match.
	if hSrc != hDest {
		t.Errorf("bundle() = %v, wanted %v", hDest, hSrc)
	}
}
