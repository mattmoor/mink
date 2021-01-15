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
	"path/filepath"
	"regexp"
	"strings"
	"text/scanner"
)

// fileIgnorePattern holds the ignorable patterns
type fileIgnorePattern struct {
	paths   []string
	regExpr string
	invert  bool
}

// sanitize the pattern to make it more file path friendly
func sanitizePattern(pattern string) string {

	// removes the all unwanted spaces from the beginning and end of the string
	// e.g. "\n\t\nfoo bar\n\t" will be trimmed to "foo bar"
	pattern = strings.TrimSpace(pattern)

	// Ignore comments and skip empty lines
	if strings.HasPrefix(pattern, "#") || pattern == "" {
		return ""
	}

	// The Patterns can start with ! symbolizing inversion of the pattern
	// When ! is seen remove the ! to clean up the pattern for Path separators
	// e.g. Patterns line !README.md
	invert := strings.HasPrefix(pattern, invertPrefix)
	if invert {
		pattern = strings.TrimPrefix(pattern, invertPrefix)
	}

	if len(pattern) > 0 {
		pattern = filepath.Clean(pattern)
		pattern = filepath.ToSlash(pattern)
		// As the patterns will be scanned relative to root directory
		// remove the leading slashes
		if len(pattern) > 1 && strings.HasPrefix(pattern, string(os.PathSeparator)) {
			pattern = strings.TrimPrefix(pattern, string(os.PathSeparator))
		}
	} else {
		return ""
	}
	// prefix the clean pattern with invertPrefix
	if invert {
		pattern = invertPrefix + pattern
	}

	return pattern
}

// toFileIgnorePattern is used to perform normalization on the pattern like
// - clean up the path to be good file path
// - make sure paths start with /
// - check if the patterns has inversions i.e !foo kind of things
// - compile the pattern to regular expression
func toFileIgnorePattern(directory, pattern string) fileIgnorePattern {
	// clean the pattern to be well formed Go path
	pattern = filepath.Clean(pattern)
	// make sure the path starts with /
	pattern = filepath.FromSlash(pattern)

	ignorePattern := fileIgnorePattern{}

	// check if it has inverts and remove them before creating paths
	if strings.HasPrefix(pattern, "!") {
		pattern = strings.TrimPrefix(pattern, "!")
		ignorePattern.invert = true
	}

	// add the parent directories if and only if the pattern has a /
	// otherwise the path is deemed to be under root
	if strings.Contains(pattern, string(os.PathSeparator)) {
		ptokens := strings.Split(pattern, string(os.PathSeparator))
		ignorePattern.paths = ptokens[:len(ptokens)-1]
	}

	ignorePattern.regExpr = toRegExpr(directory, pattern)

	return ignorePattern
}

// make each pattern as valid regular expression that can be compared
// with file path
func toRegExpr(directory, pattern string) (regExpr string) {
	regExpr = "^" + directory + string(os.PathSeparator)

	var charScanner scanner.Scanner
	charScanner.Init(strings.NewReader(pattern))

	for charScanner.Peek() != scanner.EOF {
		ch := charScanner.Next()
		switch ch {
		case '*':
			if charScanner.Peek() == '*' {
				charScanner.Next()
				// handles **
				if charScanner.Peek() == os.PathSeparator {
					charScanner.Next()
				}
				// once you moved out of path separator, treat any other character using
				// regexp wildcard
				regExpr += ".*"
			} else {
				regExpr += ".*"
			}
		case '?':
			regExpr += ".?"
		case '$', '.':
			// escape any regexp meta characters
			regExpr += regexp.QuoteMeta(string(ch))
		default:
			regExpr += string(ch)
		}
	}
	regExpr += "$"

	return regExpr
}
