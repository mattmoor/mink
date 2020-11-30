package kontext

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDirectoryHasDockerIgnore(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}
	dir := filepath.Join(wd, "ignoretestdata", "dir2")

	b, actual, err := hasDockerIgnore(dir)

	if err != nil {
		t.Error("hasDockerIgnore() = ", err)
	}

	if !b {
		t.Errorf("The directory %s has '.dockerignore', but got it does not", dir)
	}

	expected := filepath.Join(wd, "ignoretestdata", "dir2", ".dockerignore")

	if expected != actual {
		t.Errorf("Expecting file path %s but got %s", expected, actual)
	}
}

func TestDirectoryHasNoDockerIgnore(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}
	dir := filepath.Join(wd, "testdata", "dir1")
	b, _, err := hasDockerIgnore(dir)

	if err != nil {
		t.Error("hasDockerIgnore() = ", err)
	}

	if b {
		t.Errorf("The directory %s does not have '.dockerignore', but got it does", dir)
	}

}

func TestDockerIgnoredPatterns(t *testing.T) {
	ignored := []string{"lib", "*.md", "!README.md", "temp?", "target", "!target/*-runner.jar"}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}
	ignoreFile := filepath.Join(wd, "ignoretestdata", "dir2", ".dockerignore")

	patterns, err := ignorablePatterns(ignoreFile)

	if err != nil {
		t.Error("ignorablePatterns() = ", err)
	}

	if !reflect.DeepEqual(ignored, patterns) {
		t.Errorf("Expecting ignore patterns %v but got %v", ignored, patterns)
	}
}

func TestEmptyDockerIgnoredPatterns(t *testing.T) {

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}
	ignoreFile := filepath.Join(wd, "ignoretestdata", "empty", ".dockerignore")

	patterns, err := ignorablePatterns(ignoreFile)

	if err != nil {
		t.Error("ignorablePatterns() = ", err)
	}

	if len(patterns) != 0 {
		t.Errorf("Expecting empty patterns but got %v", patterns)
	}
}

func TestIgnoreables(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}
	dir := filepath.Join(wd, "ignoretestdata", "dir2")
	patterns, err := ignorablePatterns(dir + "/.dockerignore")

	var excludedDirs = &[]string{}

	if err != nil {
		t.Error("isIgnorable() = ", err)
	}

	bundlefile := &bundleFile{
		name:         "lib",
		path:         filepath.Join(dir, "lib"),
		rootDir:      dir,
		patterns:     patterns,
		excludedDirs: excludedDirs,
	}

	b := bundlefile.isIgnorable()

	if !b {
		t.Errorf("%s has to be ignored but its not.", bundlefile.name)
	}

	*excludedDirs = append(*excludedDirs, "lib")

	bundlefile.name = "tempA"
	bundlefile.path = filepath.Join(dir, "tempA")
	bundlefile.name = "tempA"

	b = bundlefile.isIgnorable()

	if !b {
		t.Errorf("%s has to be ignored but its not.", bundlefile.name)
	}

	bundlefile.name = "tempABC"
	bundlefile.path = filepath.Join(dir, "tempABC")
	bundlefile.name = "tempABC"
	*excludedDirs = append(*excludedDirs, "tempA")

	b = bundlefile.isIgnorable()

	if b {
		t.Errorf("%s should not be ignored but actually it is.", bundlefile.name)
	}

	bundlefile.name = "one.md"
	bundlefile.path = filepath.Join(dir, "one.md")
	bundlefile.name = "one.md"

	b = bundlefile.isIgnorable()

	if !b {
		t.Errorf("%s has to be ignored but its not.", bundlefile.name)
	}

	bundlefile.name = "README.md"
	bundlefile.path = filepath.Join(dir, "README.md")
	bundlefile.name = "README.md"

	b = bundlefile.isIgnorable()

	if b {
		t.Errorf("%s should not be ignored but actually it is.", bundlefile.name)
	}

	bundlefile.name = "target"
	bundlefile.path = filepath.Join(dir, "target")
	bundlefile.name = "target"

	b = bundlefile.isIgnorable()

	if !b {
		t.Errorf("%s has to be ignored but its not.", bundlefile.name)
	}

	bundlefile.name = "target/foo-runner.jar"
	bundlefile.path = filepath.Join(dir, "target", "foo-runner.jar")
	bundlefile.name = "foo-runner.jar"
	*excludedDirs = append(*excludedDirs, "target")

	b = bundlefile.isIgnorable()

	if b {
		t.Errorf("%s should not be ignored but actually it is.", bundlefile.name)
	}

}

func difference(slice1, slice2 []string) []string {
	var diff []string = make([]string, 0)

	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			if !found {
				if strings.TrimSpace(s1) != "" {
					diff = append(diff, strings.TrimSpace(s1))
				}
			}
		}
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}
