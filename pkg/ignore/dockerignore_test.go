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

package ignore

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDockerIgnore(t *testing.T) {

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	ignoreTests := map[string]struct {
		dir  string
		want map[string]Ignorable
	}{
		"dockerignore": {
			dir: filepath.Join(wd, "testdata", "dir2"),
			want: map[string]Ignorable{
				".dockerignore":          No,
				"README.md":              No,
				"one.md":                 Current,
				"src":                    No,
				"src/main":               No,
				"src/main/java":          No,
				"src/main/java/One.java": No,
				"src/main/resources":     No,
				"src/main/resources/application.properties": No,
				"tempABC":               No,
				"pom.xml":               No,
				"lib":                   Transitive,
				"target":                Current,
				"target/classes":        Transitive,
				"target/foo-runner.jar": No,
			},
		},
		"nodockerignore": {
			dir: filepath.Join(wd, "testdata", "dir1"),
			want: map[string]Ignorable{
				"baz":          No,
				"node_modules": Transitive,
				".":            No,
			},
		},
		"emptydockerignore": {
			dir: filepath.Join(wd, "testdata", "empty"),
			want: map[string]Ignorable{
				"vendor":       Transitive,
				"node_modules": Transitive,
			},
		},
		"dockerignorewithstar": {
			dir: filepath.Join(wd, "testdata", "starignore"),
			want: map[string]Ignorable{
				"README.md":                  No,
				"target/foo-runner.jar":      No,
				"target/lib/one.jar":         No,
				"target/quarkus-app/one.txt": No,
			},
		},
		"rootdir": {
			dir: filepath.Join(wd, "testdata", "starignore"),
			want: map[string]Ignorable{
				".":                          No,
				"README.md":                  No,
				"target/foo-runner.jar":      No,
				"target/lib/one.jar":         No,
				"target/quarkus-app/one.txt": No,
			},
		},
		"defaultIgnores": {
			dir: filepath.Join(wd, "testdata", "empty"),
			want: map[string]Ignorable{
				".":            No,
				"node_modules": Transitive,
				"vendor":       Transitive,
			},
		},
	}

	for name, tc := range ignoreTests {
		t.Run(name, func(t *testing.T) {
			var err error
			var ignoreScanner FileIgnorer

			if ignoreScanner, err = NewOrDefault(tc.dir); err != nil {
				t.Error("NewOrDefault()", err)
			}

			for k, v := range tc.want {
				path := filepath.Join(tc.dir, k)

				if fi, err := os.Stat(path); err != nil {
					errCtx := fmt.Sprintf("os.Stat(%q)", path)
					t.Error(errCtx, err)
				} else {
					if ignorable, err := ignoreScanner.CanIgnore(path, fi); err != nil {
						errCtx := fmt.Sprintf("CanIgnore(%q)", path)
						t.Error(errCtx, err)
					} else if v != ignorable {
						t.Errorf("File %s is Expected to be %q but got %q  ", path, v, ignorable)
					}
				}
			}
		})
	}
}
