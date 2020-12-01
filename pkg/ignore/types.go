package ignore

import "k8s.io/apimachinery/pkg/util/sets"

// Ignorer  helps identifying if a BundleFile needs to be ignored
type Ignorer interface {
	//Ignore checks file has to be ignored or not, returns true if it needs to be ignored
	Ignore(file *BundleFile) bool
	//Patterns returns the pattern associated with interface
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
	//// Patterns the patterns that needs to be checked with this BundleFile
	Patterns *sets.String
}

var DefaultPatterns = []string{".git"}
