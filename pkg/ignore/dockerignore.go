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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"path/filepath"
	"strings"
)

const (
	EscSCL           = "\\"
	DockerIgnoreFile = ".dockerignore"
)

var _ Ignorer = &defaultIgnorer{}

// defaultIgnorer is the default Ignorer which is returned when no .dockerignore file is present
// or error processing .dockerignore
type defaultIgnorer struct {
	patterns *sets.String
}

// dockerIgnorer processes .dockerignore and use the extracted patterns for ignoring files
type dockerIgnorer struct {
	patterns *sets.String
}

// NewOrDefault builds and returns the new or default Ignorer interface
func NewOrDefault(dir string) (Ignorer, error) {

	//Ensure the re-init it every time
	patterns := sets.NewString(DefaultPatterns...)

	ignoreFile := filepath.Join(dir, DockerIgnoreFile)
	_, err := os.Stat(ignoreFile)

	if os.IsNotExist(err) {
		return &defaultIgnorer{
			&patterns,
		}, nil
	}

	fr, err := os.Open(ignoreFile)

	switch {
	case os.IsNotExist(err):
		return &defaultIgnorer{
			&patterns,
		}, nil
	case err != nil:
		return &defaultIgnorer{
			&patterns,
		}, err
	}

	scanner := bufio.NewScanner(fr)

	utf8bom := []byte{0xEF, 0xBB, 0xBF}
	lineNo := 0

	for scanner.Scan() {
		scannedBytes := scanner.Bytes()

		if lineNo == 0 {
			scannedBytes = bytes.TrimPrefix(scannedBytes, utf8bom)
		}

		pattern := string(scannedBytes)
		pattern = strings.TrimSpace(pattern)
		lineNo++

		// Ignore comments
		if strings.HasPrefix("#", pattern) {
			continue
		}

		// Ignore Empty lines
		if pattern == "" {
			continue
		}

		// Normalize - clean and slash
		invert := pattern[0] == '!'
		if invert {
			pattern = strings.TrimSpace(pattern[1:])
		}

		if len(pattern) > 0 {
			pattern = filepath.Clean(pattern)
			pattern = filepath.ToSlash(pattern)
			if len(pattern) > 1 && pattern[0] == '/' {
				pattern = pattern[1:]
			}
		}

		// add invert to the cleaned pattern
		if invert {
			pattern = "!" + pattern
		}

		patterns.Insert(pattern)
	}

	if err := scanner.Err(); err != nil {
		return &defaultIgnorer{
			&patterns,
		}, fmt.Errorf("error reading file %s : %w", ignoreFile, err)
	}

	defer fr.Close()

	return &dockerIgnorer{
		&patterns,
	}, err
}

// Ignore implements Ignorer, for no dockerignore cases, where only .git is ignored
func (i *defaultIgnorer) Ignore(bundleFile *BundleFile) bool {
	return bundleFile.ignore()
}

// Patterns implements Ignorer, returns the pattern associated with the ignorer
func (i *defaultIgnorer) Patterns() *sets.String {
	return i.patterns
}

// Ignore implements Ignorer, for  dockerignore cases
func (id *dockerIgnorer) Ignore(bundleFile *BundleFile) bool {
	return bundleFile.ignore()
}

// Patterns implements Ignorer, returns the pattern associated with the ignorer
func (id *dockerIgnorer) Patterns() *sets.String {
	return id.patterns
}

// ignore is the abstracted private file ignore processing
func (f *BundleFile) ignore() bool {

	var ignoredMatch bool

	for _, p := range f.Patterns.List() {

		var negative bool
		var mp string

		//Check if parent is ignorable
		parentDir := filepath.Dir(f.Path)
		isParentExcluded := f.ExcludedDirs.HasAny(parentDir)

		if strings.HasPrefix(p, "!") {
			negative = true
			mp = filepath.Join(f.RootDir, strings.TrimPrefix(p, "!"))
		} else {
			//Add the parent directory to the Path
			mp = filepath.Join(f.RootDir, p)
		}

		if isParentExcluded && !negative {
			continue
		}

		mp = filepath.Clean(mp)

		//Handle negative directories
		if negative && f.IsDir {
			if "*" == filepath.Base(mp) && "*" != p {
				noStar := strings.TrimSuffix(mp, "*")
				mp = noStar + EscSCL + "\\*"
			}
		}

		matched, err := filepath.Match(mp, f.Path)

		if errors.Is(err, filepath.ErrBadPattern) {
			//if bad pattern then allow the file to be unignored
			ignoredMatch = false
		} else {
			if negative && matched {
				ignoredMatch = false
			} else if isParentExcluded && (negative || matched) {
				//if parent is excluded, its negative and not matched
				ignoredMatch = true
			} else {
				//all other cases
				ignoredMatch = ignoredMatch || matched
			}
		}
	}

	// Collect the excluded directories for checks for its sub dirs or files
	if f.IsDir && ignoredMatch {
		*f.ExcludedDirs = f.ExcludedDirs.Insert(f.Path)
	}

	return ignoredMatch
}
