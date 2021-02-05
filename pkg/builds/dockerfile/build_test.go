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

package dockerfile

import (
	"reflect"
	"testing"
)

func TestRemoveKanikoContext(t *testing.T) {
	testCases := []struct {
		args          []string
		expectedValue string
		expectedArgs  []string
	}{
		{
			args:          []string{"sample", "argument"},
			expectedValue: "",
			expectedArgs:  []string{"sample", "argument"},
		},
		{
			args:          []string{"--context", "myctx", "sample", "argument"},
			expectedValue: "myctx",
			expectedArgs:  []string{"sample", "argument"},
		},
		{
			args:          []string{"sample", "--context", "myctx", "argument"},
			expectedValue: "myctx",
			expectedArgs:  []string{"sample", "argument"},
		},
		{
			args:          []string{"sample", "argument", "--context", "myctx"},
			expectedValue: "myctx",
			expectedArgs:  []string{"sample", "argument"},
		},
		{
			args:          []string{"--context=myctx", "sample", "argument"},
			expectedValue: "myctx",
			expectedArgs:  []string{"sample", "argument"},
		},
		{
			args:          []string{"sample", "--context=myctx", "argument"},
			expectedValue: "myctx",
			expectedArgs:  []string{"sample", "argument"},
		},
		{
			args:          []string{"sample", "argument", "--context=myctx"},
			expectedValue: "myctx",
			expectedArgs:  []string{"sample", "argument"},
		},
	}

	name := "context"
	for _, tc := range testCases {
		value, args := RemoveArgument(tc.args, name)

		if value != tc.expectedValue {
			t.Errorf("got value %s expected %s\n", value, tc.expectedValue)
		}

		if !reflect.DeepEqual(args, tc.expectedArgs) {
			t.Errorf("resulting arguments %v expected %v\n", args, tc.expectedArgs)
		}
	}
}
