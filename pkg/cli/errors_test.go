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

package cli

import "testing"

func TestErrMissingFlag(t *testing.T) {
	err := ErrMissingFlag("foo")
	if want, got := `"--foo": is a required flag`, err.Error(); want != got {
		t.Errorf("Error() = %s, wanted %s", got, want)
	}

	err = ErrMissingFlag("bar")
	if want, got := `"--bar": is a required flag`, err.Error(); want != got {
		t.Errorf("Error() = %s, wanted %s", got, want)
	}
}

func TestErrInvalidValue(t *testing.T) {
	// Simple string.
	err := ErrInvalidValue("foo", "blah")
	if want, got := `"--foo": blah`, err.Error(); want != got {
		t.Errorf("Error() = %s, wanted %s", got, want)
	}

	// With formatting
	err = ErrInvalidValue("foo", "a %q b %d c", "d", 3)
	if want, got := `"--foo": a "d" b 3 c`, err.Error(); want != got {
		t.Errorf("Error() = %s, wanted %s", got, want)
	}
}
