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
			args:          []string{"dummy", "argument"},
			expectedValue: "",
			expectedArgs:  []string{"dummy", "argument"},
		},
		{
			args:          []string{"--context", "myctx", "dummy", "argument"},
			expectedValue: "myctx",
			expectedArgs:  []string{"dummy", "argument"},
		},
		{
			args:          []string{"dummy", "--context", "myctx", "argument"},
			expectedValue: "myctx",
			expectedArgs:  []string{"dummy", "argument"},
		},
		{
			args:          []string{"dummy", "argument", "--context", "myctx"},
			expectedValue: "myctx",
			expectedArgs:  []string{"dummy", "argument"},
		},
		{
			args:          []string{"--context=myctx", "dummy", "argument"},
			expectedValue: "myctx",
			expectedArgs:  []string{"dummy", "argument"},
		},
		{
			args:          []string{"dummy", "--context=myctx", "argument"},
			expectedValue: "myctx",
			expectedArgs:  []string{"dummy", "argument"},
		},
		{
			args:          []string{"dummy", "argument", "--context=myctx"},
			expectedValue: "myctx",
			expectedArgs:  []string{"dummy", "argument"},
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
