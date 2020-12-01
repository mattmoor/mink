package ignore

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryHasDockerIgnore(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	dir := filepath.Join(wd, "testdata", "dir2")

	ignorer, err := NewOrDefault(dir)

	if err != nil {
		t.Error("hasDockerIgnore() = ", err)
	}

	patterns := ignorer.Patterns()

	if patterns.Len() == 0 {
		t.Errorf("The directory %s has '.dockerignore', but got it does not", dir)
	}

	if got, want := patterns.Len(), int(7); got != want {
		t.Errorf("Patterns() = %d, wanted %d", got, want)
	}
}

func TestDirectoryHasNoDockerIgnore(t *testing.T) {
	expected := sets.NewString(DefaultPatterns...)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}
	dir := filepath.Join(wd, "testdata", "dir1")

	ignorer, err := NewOrDefault(dir)

	if err != nil {
		t.Error("hasDockerIgnore() = ", err)
	}

	actual := ignorer.Patterns()

	if got, want := actual.Len(), int(1); got != want {
		t.Errorf("Patterns() = %d, wanted %d", got, want)
	}

	if !expected.Equal(*actual) {
		t.Errorf("Actual %v Patterns differ expected %v by %v", actual, expected, actual.Difference(expected))
	}
}

func TestDockerIgnoredPatterns(t *testing.T) {
	expected := sets.NewString(".git", "lib", "*.md", "!README.md", "temp?", "target", "!target/*-runner.jar")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	dir := filepath.Join(wd, "testdata", "dir2")

	ignorer, err := NewOrDefault(dir)

	if err != nil {
		t.Error("ignorablePatterns() = ", err)
	}

	actual := ignorer.Patterns()

	if got, want := actual.Len(), int(7); got != want {
		t.Errorf("Patterns() = %d, wanted %d", got, want)
	}

	if !expected.Equal(*actual) {
		t.Errorf("Actual %v Patterns differ expected %v by %v", actual, expected, actual.Difference(expected))
	}
}

func TestEmptyDockerIgnoredPatterns(t *testing.T) {
	expected := sets.NewString(".git")
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	dir := filepath.Join(wd, "testdata", "empty")

	ignorer, err := NewOrDefault(dir)

	if err != nil {
		t.Error("ignorablePatterns() = ", err)
	}

	actual := ignorer.Patterns()

	if got, want := actual.Len(), int(1); got != want {
		t.Errorf("Patterns() = %d, wanted %d", got, want)
	}

	if !expected.Equal(*actual) {
		t.Errorf("Actual %v Patterns differ expected %v by %v", actual, expected, actual.Difference(expected))
	}
}

func TestIgnoreables(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	dir := filepath.Join(wd, "testdata", "dir2")

	ignorer, err := NewOrDefault(dir)

	var excludedDirs = sets.String{}
	var incudedDirs = sets.String{}

	if err != nil {
		t.Error("isIgnorable() = ", err)
	}

	bundlefile := &BundleFile{
		Name:         "lib",
		Path:         filepath.Join(dir, "lib"),
		RootDir:      dir,
		ExcludedDirs: &excludedDirs,
		IncludedDirs: &incudedDirs,
		Patterns:     ignorer.Patterns(),
	}

	b := ignorer.Ignore(bundlefile)

	if !b {
		t.Errorf("%s has to be ignored but its not.", bundlefile.Name)
	}

	excludedDirs.Insert("lib")

	bundlefile.Name = "tempA"
	bundlefile.Path = filepath.Join(dir, "tempA")

	b = ignorer.Ignore(bundlefile)

	if !b {
		t.Errorf("%s has to be ignored but its not.", bundlefile.Name)
	}

	bundlefile.Name = "tempABC"
	bundlefile.Path = filepath.Join(dir, "tempABC")

	excludedDirs.Insert("tempA")

	b = ignorer.Ignore(bundlefile)

	if b {
		t.Errorf("%s should not be ignored but actually it is.", bundlefile.Name)
	}

	bundlefile.Name = "one.md"
	bundlefile.Path = filepath.Join(dir, "one.md")

	b = ignorer.Ignore(bundlefile)

	if !b {
		t.Errorf("%s has to be ignored but its not.", bundlefile.Name)
	}

	bundlefile.Name = "README.md"
	bundlefile.Path = filepath.Join(dir, "README.md")

	b = ignorer.Ignore(bundlefile)

	if b {
		t.Errorf("%s should not be ignored but actually it is.", bundlefile.Name)
	}

	bundlefile.Name = "target"
	bundlefile.Path = filepath.Join(dir, "target")

	b = ignorer.Ignore(bundlefile)

	if !b {
		t.Errorf("%s has to be ignored but its not.", bundlefile.Name)
	}

	bundlefile.Name = "target/foo-runner.jar"
	bundlefile.Path = filepath.Join(dir, "target", "foo-runner.jar")
	excludedDirs.Insert("target")

	b = ignorer.Ignore(bundlefile)

	if b {
		t.Errorf("%s should not be ignored but actually it is.", bundlefile.Name)
	}

}
