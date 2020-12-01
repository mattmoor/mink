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

import "k8s.io/apimachinery/pkg/util/sets"

// Ignorer  helps identifying if a BundleFile needs to be ignored
type Ignorer interface {
	// Ignore checks file has to be ignored or not, returns true if it needs to be ignored
	Ignore(file *BundleFile) bool
	// Patterns returns the pattern associated with interface
	Patterns() *sets.String
}

// BundleFile  represents the file that will be added to the bundle
type BundleFile struct {
	// RootDir the root directory where the .dockerignore is searched for
	RootDir string
	// Name the basename of the file that needs to be checked for ignorability
	Name string
	// Path the absolute path of the file that needs to be checked for ignorability
	Path string
	// ExcludedDirs a collection of ignored directories
	ExcludedDirs *sets.String
	// IsDir the file that is checked for ignorability is a directory ?
	IsDir bool
	// Patterns the patterns that needs to be checked with this BundleFile
	Patterns *sets.String
	// IncludedDirs Keeps track of directories that are included
	IncludedDirs *sets.String
}

var DefaultPatterns = []string{".git"}
