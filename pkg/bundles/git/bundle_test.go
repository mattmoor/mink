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
	"context"
	"testing"
)

func TestBundleLayersWithRef(t *testing.T) {
	// Check that if we bundle testdata it has the expected size.
	l, err := bundle(context.Background(), Options{
		URL: "https://github.com/knative/pkg.git",
		Ref: "refs/heads/release-0.10", // Pick something we've frozen
	})
	if err != nil {
		t.Error("bundle() =", err)
	}
	sz, err := l.Size()
	if err != nil {
		t.Error("l.Size() =", err)
	}
	if got, want := sz, int64(15246139); got != want {
		t.Errorf("Size() = %d, wanted %d", got, want)
	}
}

func TestBundleLayersNoRef(t *testing.T) {
	// Check that if we bundle testdata it has the expected size.
	l, err := bundle(context.Background(), Options{
		URL: "https://github.com/knative/pkg.git",
	})
	if err != nil {
		t.Error("bundle() =", err)
	}
	sz, err := l.Size()
	if err != nil {
		t.Error("l.Size() =", err)
	}
	if sz == 0 {
		t.Error("Size() = 0, wanted non-zero")
	}
}
