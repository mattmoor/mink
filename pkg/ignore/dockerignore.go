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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// dockerIgnoreFile defines the Name of the scanner file
	dockerIgnoreFile = ".dockerignore"
	// invertPrefix is used to prefix the Patterns to symbolize inversions
	invertPrefix = "!"
)

var (
	_ FileIgnorer = &defaultIgnorer{}
)

// dockerIgnorer processes .dockerignore and use the extracted Patterns for ignoring files
type dockerIgnorer struct {
	// root directory to start the scanner
	directory string
	// index of all excludes directories
	excludedDirs sets.String
	// patterns holds all the possible Ignorable patterns
	ignorePatterns []fileIgnorePattern
}

// NewOrDefault builds and returns the new or default FileIgnorer interface
// In all file not found cases the this wil return the default ignorer
func NewOrDefault(dir string) (FileIgnorer, error) {

	ignoreFile := filepath.Join(dir, dockerIgnoreFile)
	var rawPatterns []string

	if _, err := os.Stat(ignoreFile); os.IsNotExist(err) {
		return &defaultIgnorer{
			directory: dir,
		}, nil
	}

	if err := scanAndBuildPatternsList(ignoreFile, &rawPatterns); err != nil {
		return nil, err
	}

	// every new gets clean includes
	ignorePatterns := make([]fileIgnorePattern, 0)

	for _, ip := range rawPatterns {
		ignorePatterns = append(ignorePatterns, toFileIgnorePattern(dir, ip))
	}

	return &dockerIgnorer{
		directory:      dir,
		excludedDirs:   defaultPatterns,
		ignorePatterns: ignorePatterns,
	}, nil
}

// CanIgnore is the directory scanner to scan directory for .dockerignore based
// ignore patterns
func (i *dockerIgnorer) CanIgnore(path string, fi os.FileInfo) (Ignorable, error) {
	// start with assuming nothing is ignored
	ignorable := No

	// dont append the root directory as its always included
	var isRootDir bool
	if isRootDir = path == i.directory; isRootDir {
		return No, nil
	}

	// flag to keep track if file or directory is excluded
	isExcluded := false
	// flag to keep track transitive directories
	// Transitive directories are the ones that are excluded
	// but might have files under them with inversions
	// e.g. with a ignore file like
	// foo
	// !foo/bar/one.txt
	// In above case directory foo is transitive
	hasTransitives := false

	for _, igp := range i.ignorePatterns {
		var re *regexp.Regexp
		var err error

		if re, err = regexp.Compile(igp.regExpr); err != nil {
			isExcluded = false
			continue
		}
		// check if the parent path is an excluded pattern in the list or current pattern
		// matches the parent directory
		if parentDir := filepath.Dir(path); i.excludedDirs.Has(parentDir) || re.MatchString(filepath.Dir(path)) {
			// if the parent directory is not in the list check if it matches
			// any pattern
			if !igp.invert {
				isExcluded = true
				ignorable = Current
				i.excludedDirs.Insert(parentDir)
				continue
			}
		}

		if regxMatches := re.FindAllStringSubmatch(path, -1); len(regxMatches) > 0 {
			isExcluded = true
			// when a file or directory matches the pattern but has inversion
			// then add the file to include list
			if igp.invert {
				ignorable = No
			} else {
				if fi.IsDir() {
					i.excludedDirs.Insert(path)
				}
				ignorable = Current
			}
		}

		if fi.IsDir() && !isRootDir && igp.invert && igp.paths != nil {
			pPath := strings.Join(igp.paths, string(os.PathSeparator))
			pPath = filepath.Join(i.directory, pPath)
			if path == pPath {
				hasTransitives = true
			}
		}
	}

	// the directory is not transitive and not root directory check is being done
	// at last to avoid skipping directories that are not listed in the ignore file
	// with "!" i.e. implicit includes
	if fi.IsDir() && !isRootDir && isExcluded && !hasTransitives {
		ignorable = Transitive
	}

	return ignorable, nil
}

// scanAndBuildPatternsList takes the file typically the .dockerignore file
// splits the file by new line (\n) and normalize them with following rules
// - removes the UTF8 Byte Order Mark (BOM) characters
// - scanner comments line starting with "#"
// - trim spaces in the pattern
// - make the pattern as clean filenames using golang filepath utils
func scanAndBuildPatternsList(ignoreFile string, patterns *[]string) error {
	fr, err := os.Open(ignoreFile)
	if err != nil {
		return err
	}

	defer fr.Close()

	scanner := bufio.NewScanner(fr)

	// UTF8 byte order mark (BOM) which are typically first three bytes of the file with
	// the hexadecimal characters: EF,BB,BF
	utf8bom := []byte{0xEF, 0xBB, 0xBF}

	for lineNo := 0; scanner.Scan(); lineNo++ {
		scannedBytes := scanner.Bytes()

		if lineNo == 0 {
			scannedBytes = bytes.TrimPrefix(scannedBytes, utf8bom)
		}

		if p := sanitizePattern(string(scannedBytes)); p != "" {
			*patterns = append(*patterns, p)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s : %w", ignoreFile, err)
	}

	return nil
}
