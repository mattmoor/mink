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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type bundleFile struct {
	rootDir      string
	patterns     []string
	name         string
	path         string
	excludedDirs *[]string
	isDir        bool
}

func checkIfParentExcluded(files *[]string, file string) bool {
	for _, f := range *files {
		if f == file {
			return true
		}
	}
	return false
}

func (f *bundleFile) handleIgnoreableFile() bool {
	isIgnoreable := f.isIgnorable()
	//add only parent ignored directories as those are the ones needed
	if f.isDir && isIgnoreable {
		*f.excludedDirs = append(*f.excludedDirs, f.path)
	}
	return isIgnoreable
}

func hasDockerIgnore(dir string) (bool, string, error) {
	name := filepath.Join(dir, ".dockerignore")
	finfo, err := os.Stat(name)

	if os.IsNotExist(err) {
		return false, name, nil
	}

	if err != nil {
		return false, name, err
	}

	return !finfo.IsDir(), name, err
}

func (f *bundleFile) isIgnorable() bool {
	var ignoredMatch bool

	for _, p := range f.patterns {
		var negative bool
		var mp string

		//Check if parent is ignorable
		parentDir := filepath.Dir(f.path)
		isParentExcluded := checkIfParentExcluded(f.excludedDirs, parentDir)

		if strings.HasPrefix(p, "!") {
			mp = filepath.Join(f.rootDir, strings.TrimPrefix(p, "!"))
			negative = true
		} else {
			//Add the parent directory to the path
			mp = filepath.Join(f.rootDir, p)
		}

		if isParentExcluded && !negative {
			continue
		}

		mp = filepath.Clean(mp)
		matched, err := filepath.Match(mp, f.path)

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

	return ignoredMatch
}

func ignorablePatterns(ignoreFile string) ([]string, error) {
	var ignorables []string

	fr, err := os.Open(ignoreFile)

	switch {
	case os.IsNotExist(err):
		return ignorables, nil
	case err != nil:
		return nil, err
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
		lineNo++

		//Ignore comments
		if strings.HasPrefix("#", pattern) {
			continue
		}

		pattern = strings.TrimSpace(pattern)

		//Ignore Empty lines
		if pattern == "" {
			continue
		}

		//Normalize - clean and slash
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
		//add invert to the cleaned pattern
		if invert {
			pattern = "!" + pattern
		}

		ignorables = append(ignorables, pattern)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s : %v", ignoreFile, err)
	}

	defer fr.Close()

	return ignorables, err
}
