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

func TestIgnorableEnum(t *testing.T) {
	ignorableEnumsTest := map[string]struct {
		enum Ignorable
		want string
	}{
		"no":         {enum: No, want: "No"},
		"transitive": {enum: Transitive, want: "Transitive"},
		"current":    {enum: Current, want: "Current"},
		"default":    {enum: Ignorable(55), want: "Unknown"},
	}

	for name, tc := range ignorableEnumsTest {
		t.Run(name, func(t *testing.T) {
			s := tc.enum.String()
			if s != tc.want {
				t.Errorf("Want %s but got %s  ", tc.want, s)
			}
		})
	}
}

func TestDefaultScanDir(t *testing.T) {
	if wd, err := os.Getwd(); err != nil {
		t.Fatal("os.Getwd() =", err)
	} else {
		ignoreTests := map[string]struct {
			dir  string
			want map[string]Ignorable
		}{
			"rootdir": {
				dir: filepath.Join(wd, "testdata", "dir1"),
				want: map[string]Ignorable{
					".":            No,
					"node_modules": Transitive,
					"baz":          No,
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
							t.Errorf("File %s is Expected %#v but got %#v  ", path, v, ignorable)
						}
					}
				}
			})
		}
	}
}
