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

package command_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/mattmoor/mink/pkg/bundles"
	"github.com/mattmoor/mink/pkg/command"
	"github.com/mattmoor/mink/pkg/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	expectedDigest = "sha256:8e65ec4b80519d869e8d600fdf262c6e8cd3f6c7e8382406d9cb039f352a69bc"
	verbose        = false
)

func TestCommandResolve(t *testing.T) {
	bundle := "gcr.io/fake-project/fake-source-bundle:latest"
	ns := "mytest"
	image := "gcr.io/mattmoor/myimage:latest"
	dir := filepath.Join("test_data", "dockerfile")
	fakeDigests := []string{expectedDigest}
	expectedImageText := "image: " + image + "@" + expectedDigest

	ctx, _, tektonClient := tests.FakeContextWithClients(t)

	go tests.FakeTaskRunner(t, ctx, tektonClient, ns, fakeDigests)

	var out bytes.Buffer
	cmd := command.NewResolveCommand(ctx)
	cmd.SetOut(&out)
	args := []string{
		"--namespace", ns,
		"--directory", dir,
		"--bundle", bundle,
		"--image", image,
		"--filename", filepath.Join(dir, "service.yaml"),
	}

	err := cmd.ParseFlags(args)
	require.NoError(t, err, "failed to parse flags")

	// lets validate the arguments
	err = cmd.PreRunE(cmd, args)
	require.NoError(t, err, "failed to validate command")

	// lets fake out the bundle writer
	err = bundles.TestWithFakes(func() error {
		return cmd.RunE(cmd, nil)
	})
	require.NoError(t, err, "failed to run test")

	assert.Contains(t, out.String(), expectedImageText, "output should include the image and manifest")

	t.Logf("the output included text: %s\n", expectedImageText)
	if verbose {
		t.Logf("got output %s\n", out.String())
	}
}
