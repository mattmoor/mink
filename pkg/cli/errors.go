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

import "fmt"

// FlagError implements error to provide good diagnostic messages when
// CLI flags are used improperly.
type FlagError struct {
	Flag    string
	Message string
}

var _ error = (*FlagError)(nil)

// Error implements error
func (fe *FlagError) Error() string {
	return fmt.Sprintf("%q: %s", "--"+fe.Flag, fe.Message)
}

// ErrMissingFlag creates an error indicating a required flag is missing.
func ErrMissingFlag(f string) *FlagError {
	return &FlagError{
		Flag:    f,
		Message: "is a required flag",
	}
}

// ErrInvalidValue creates an error indicating a flag received a bad value.
func ErrInvalidValue(f string, format string, values ...interface{}) *FlagError {
	return &FlagError{
		Flag:    f,
		Message: fmt.Sprintf(format, values...),
	}
}
