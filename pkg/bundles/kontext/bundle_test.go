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

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	dockerIgnoreTests := map[string]struct {
		dir  string
		want []string
		size int64
	}{
		"default": {
			dir: filepath.Join(filepath.Dir(wd), "kontext", "testdata"),
			want: []string{
				"/var/run/kontext",
				"/var/run/kontext/dir1",
				"/var/run/kontext/dir1/baz",
				"/var/run/kontext/dir2",
				"/var/run/kontext/dir2/bar",
				"/var/run/kontext/dir3",
				"/var/run/kontext/dir3/.dockerignore",
				"/var/run/kontext/dir3/README.md",
				"/var/run/kontext/dir3/lib",
				"/var/run/kontext/dir3/lib/f1",
				"/var/run/kontext/dir3/lib/f2",
				"/var/run/kontext/dir3/one.md",
				"/var/run/kontext/dir3/pom.xml",
				"/var/run/kontext/dir3/src",
				"/var/run/kontext/dir3/src/main",
				"/var/run/kontext/dir3/src/main/java",
				"/var/run/kontext/dir3/src/main/java/One.java",
				"/var/run/kontext/dir3/src/main/resources",
				"/var/run/kontext/dir3/src/main/resources/application.properties",
				"/var/run/kontext/dir3/target",
				"/var/run/kontext/dir3/target/classes",
				"/var/run/kontext/dir3/target/classes/.gitkeep",
				"/var/run/kontext/dir3/target/foo-runner.jar",
				"/var/run/kontext/dir3/target/foo.jar",
				"/var/run/kontext/dir3/tempA",
				"/var/run/kontext/dir3/tempA/.gitkeep",
				"/var/run/kontext/dir3/tempABC",
				"/var/run/kontext/dir3/tempABC/.gitkeep",
				"/var/run/kontext/dir3/tempB",
				"/var/run/kontext/dir3/tempB/.gitkeep",
				"/var/run/kontext/starignore",
				"/var/run/kontext/starignore/.dockerignore",
				"/var/run/kontext/starignore/README.md",
				"/var/run/kontext/starignore/src",
				"/var/run/kontext/starignore/src/main",
				"/var/run/kontext/starignore/src/main/java",
				"/var/run/kontext/starignore/src/main/java/One.java",
				"/var/run/kontext/starignore/src/main/resources",
				"/var/run/kontext/starignore/src/main/resources/application.properties",
				"/var/run/kontext/starignore/target",
				"/var/run/kontext/starignore/target/classes",
				"/var/run/kontext/starignore/target/classes/.gitkeep",
				"/var/run/kontext/starignore/target/foo-runner.jar",
				"/var/run/kontext/starignore/target/foo.jar",
				"/var/run/kontext/starignore/target/lib",
				"/var/run/kontext/starignore/target/lib/one.jar",
				"/var/run/kontext/starignore/target/quarkus-app",
				"/var/run/kontext/starignore/target/quarkus-app/one.txt",
			},
			size: 1187,
		},
		"noStarExcludes": {
			dir: filepath.Join(filepath.Dir(wd), "kontext", "testdata", "dir3"),
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
			dir: filepath.Join(filepath.Dir(wd), "kontext", "testdata", "starignore"),
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
