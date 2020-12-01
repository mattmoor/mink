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
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func TestBundleLayerIndex(t *testing.T) {
	// Check that if we bundle testdata it has the expected size.
	if l, err := bundle("./testdata"); err != nil {
		t.Error("bundle() =", err)
	} else {
		if sz, err := l.Size(); err != nil {
			t.Error("l.Size() =", err)
		} else if got, want := sz, int64(204); got != want {
			t.Errorf("Size() = %d, wanted %d", got, want)
		}
	}
}

func TestDockerIgnorableBundle(t *testing.T) {

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	dockerIgnoreTests := map[string]struct {
		dir  string
		want []string
		size int64
	}{
		"noStarExcludes": {
			dir: filepath.Join(filepath.Dir(wd), "..", "ignore", "testdata", "dir2"),
			want: []string{
				"/var/run/kontext",
				"/var/run/kontext/.dockerignore",
				"/var/run/kontext/README.md",
				"/var/run/kontext/pom.xml",
				"/var/run/kontext/src",
				"/var/run/kontext/src/main",
				"/var/run/kontext/src/main/java",
				"/var/run/kontext/src/main/java/One.java",
				"/var/run/kontext/src/main/resources",
				"/var/run/kontext/src/main/resources/application.properties",
				"/var/run/kontext/target/foo-runner.jar",
				"/var/run/kontext/tempABC",
				"/var/run/kontext/tempABC/.gitkeep",
			},
			size: 536,
		},
		"starignore": {
			dir: filepath.Join(filepath.Dir(wd), "..", "ignore", "testdata", "starignore"),
			want: []string{
				"/var/run/kontext",
				"/var/run/kontext/README.md",
				"/var/run/kontext/target/foo-runner.jar",
				"/var/run/kontext/target/lib/one.jar",
				"/var/run/kontext/target/quarkus-app/one.txt",
			},
			size: 251,
		},
	}

	for name, tc := range dockerIgnoreTests {
		t.Run(name, func(t *testing.T) {
			//Bundle
			var l v1.Layer

			if l, err = bundle(tc.dir); err != nil {
				t.Error("bundle() = ", err)
			} else if s, err := l.Size(); err != nil {
				t.Error("bundle#l.Size() = ", err)
			} else if s != tc.size {
				t.Errorf("Expected Size = %d, got %d", tc.size, s)
			}

			if r, err := l.Uncompressed(); err != nil {
				t.Error("error reading layer image", err)
			} else {
				var actualBundleFiles []string

				tr := tar.NewReader(r)
				defer r.Close()

				for {
					if h, err := tr.Next(); errors.Is(err, io.EOF) {
						break
					} else if err != nil {
						t.Error("error reading tar", err)
						break
					} else {
						actualBundleFiles = append(actualBundleFiles, h.Name)
					}
				}

				if !cmp.Equal(tc.want, actualBundleFiles) {
					t.Errorf("Diff of want - got\n%s", cmp.Diff(tc.want, actualBundleFiles))
				}
			}
		})
	}
}
