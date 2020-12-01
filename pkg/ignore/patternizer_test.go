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
	"reflect"
	"testing"
)

func TestSanitizePattern(t *testing.T) {

	patternsTest := map[string]struct {
		pattern string
		want    string
	}{
		"comment":             {pattern: "#comment", want: ""},
		"commentWithSpace":    {pattern: "# comment", want: ""},
		"empty":               {pattern: " ", want: ""},
		"newline":             {pattern: "\n", want: ""},
		"tab":                 {pattern: "\t", want: ""},
		"newlineAndTab":       {pattern: "\n\t", want: ""},
		"justInvert":          {pattern: "!", want: ""},
		"inverted":            {pattern: "!README.md", want: "!README.md"},
		"invertedDirOnlyPath": {pattern: "!target/classes", want: "!target/classes"},
		"invertedDirPath":     {pattern: "!target/foo-runner.jar", want: "!target/foo-runner.jar"},
		"dot":                 {pattern: "./README.md", want: "README.md"},
		"slash":               {pattern: "/lib/one", want: "lib/one"},
	}

	for name, tc := range patternsTest {
		t.Run(name, func(t *testing.T) {
			if p := sanitizePattern(tc.pattern); p != tc.want {
				t.Errorf("Expected %#v but got %#v  ", tc.want, p)
			}
		})
	}
}

func TestPatternizer(t *testing.T) {

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	patternTests := map[string]struct {
		dir     string
		pattern string
		want    fileIgnorePattern
	}{
		"README.md": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "README.md",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf("^%s%sREADME\\.md$", filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
		"*.md": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "*.md",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf("^%s%s.*\\.md$", filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
		"lib": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "lib",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf("^%s%slib$", filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
		"lib*": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "lib*",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf("^%s%slib.*$", filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
		"temp?": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "temp?",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf(`^%s%stemp.?$`, filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
		"*": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "*",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf(`^%s%s.*$`, filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
		"**": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "**",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf(`^%s%s.*$`, filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
		"**/": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "**/",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf(`^%s%s.*$`, filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
		"*foo": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "*foo",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf(`^%s%s.*foo$`, filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
		"**foo": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "**foo",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf(`^%s%s.*foo$`, filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
		"**foo/bar": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "**foo/bar",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf(`^%s%s.*foo/bar$`, filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   []string{"**foo"},
				invert:  false,
			},
		},
		"!target/foo-runner": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "!target/foo-runner",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf(`^%s%starget/foo-runner$`, filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   []string{"target"},
				invert:  true,
			},
		},
		"target/lib/*.jar": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "target/lib/*.jar",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf(`^%s%starget/lib/.*\.jar$`, filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   []string{"target", "lib"},
				invert:  false,
			},
		},
		"$jar": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "$jar",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf(`^%s%s\$jar$`, filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
		".jar": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: ".jar",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf(`^%s%s\.jar$`, filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
		"$one.jar": {
			dir:     filepath.Join(wd, "testdata", "dir2"),
			pattern: "$one.jar",
			want: fileIgnorePattern{
				regExpr: fmt.Sprintf(`^%s%s\$one\.jar$`, filepath.Join(wd, "testdata", "dir2"), string(os.PathSeparator)),
				paths:   nil,
				invert:  false,
			},
		},
	}

	for name, tc := range patternTests {
		t.Run(name, func(t *testing.T) {
			if fp := toFileIgnorePattern(tc.dir, tc.pattern); fp.regExpr != tc.want.regExpr {
				t.Errorf("Expected %#v but got %#v  ", tc.want.regExpr, fp.regExpr)
			} else if !reflect.DeepEqual(tc.want.paths, fp.paths) {
				t.Errorf("Expected %#v but got %#v  ", tc.want.paths, fp.paths)
			} else if fp.invert != tc.want.invert {
				t.Errorf("Expected %#v but got %#v  ", tc.want.invert, fp.invert)
			}
		})
	}
}
