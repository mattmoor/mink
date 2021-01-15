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
	"os"

	"k8s.io/apimachinery/pkg/util/sets"
)

// Ignorable  holds the what type of ignore
type Ignorable int

const (
	// No marks the directory or file not to be ignored
	No Ignorable = iota
	// Transitive marks the directory is transitive ignore so not to be skipped
	Transitive
	// Current marks the directory to be not having any transitive dependencies
	// so clean be skipped from walking
	Current
)

// FileIgnorer  helps identifying if a BundleFile needs to be ignored
type FileIgnorer interface {
	// CanIgnore checks file has to be ignored or not, returns true if it needs to be ignored
	CanIgnore(path string, fi os.FileInfo) (Ignorable, error)
}

var (
	_               FileIgnorer = (*defaultIgnorer)(nil)
	defaultPatterns             = sets.NewString(".git", "vendor", "node_modules")
)

// String implements fmt.Stringer
func (i Ignorable) String() string {
	switch i {
	case Current:
		return "Current"
	case Transitive:
		return "Transitive"
	case No:
		return "No"
	default:
		return "Unknown"
	}
}

// defaultIgnorer is the default FileIgnorer which is returned when no .dockerignore file is present
// or error processing .dockerignore
type defaultIgnorer struct {
	// root directory to start the scanner
	directory string
}

// CanIgnore is the directory scanner to scan directory for standard ignore patterns
func (i *defaultIgnorer) CanIgnore(path string, fi os.FileInfo) (Ignorable, error) {
	// start with assuming nothing is ignored
	ignorable := No

	// dont append the root directory as its always included
	if isRootDir := path == i.directory; isRootDir {
		return No, nil
	}

	// if the dir is any of default patterns then skip scanning the dir
	if defaultPatterns.Has(fi.Name()) {
		return Transitive, nil
	}

	return ignorable, nil
}
