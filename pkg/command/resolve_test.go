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
	fakeDigests := []string{"sha256:2e25a0687fe87783ec71298ed06b47307638e13a58a851c770bb6bb03f832b46"}
	expectedImageText := "image: " + image + "@" + fakeDigests[0]

	ctx, _, tektonClient := tests.FakeContextWithClients(t)

	go tests.FakeTaskRunner(t, ctx, tektonClient, ns, fakeDigests)

	var out bytes.Buffer
	o := &command.ResolveOptions{}
	cmd := command.NewResolveCommand()
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
	err = o.Validate(cmd, args)
	require.NoError(t, err, "failed to validate command")

	o.Ctx = ctx

	// lets fake out the bundle writer
	err = bundles.TestWithFakes(func() error {
		return o.Execute(cmd, nil)

	})
	require.NoError(t, err, "failed to run test")

	assert.Contains(t, out.String(), expectedImageText, "output should include the image and manifest")

	t.Logf("the output included text: %s\n", expectedImageText)
	if verbose {
		t.Logf("got output %s\n", out.String())
	}
}
