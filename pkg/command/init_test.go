package command_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-helpers/v3/pkg/testhelpers"
	"github.com/mattmoor/mink/pkg/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandInit(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	require.NoError(t, err, "could not create temp dir")

	t.Logf("running tests in %s\n", tmpDir)

	fs, err := ioutil.ReadDir("test_data")

	for _, f := range fs {
		if !f.IsDir() {
			continue
		}
		name := f.Name()
		srcFile := filepath.Join("test_data", name)
		require.DirExists(t, srcFile)

		destDir := filepath.Join(tmpDir, name)
		err = files.CopyDirOverwrite(srcFile, destDir)
		require.NoError(t, err, "failed to copy %s to %s", srcFile, destDir)

		cmd := command.NewInitCommand()
		o := &command.InitOptions{
			Dir:        destDir,
			Dockerfile: "Dockerfile",
		}
		err = o.Execute(cmd, nil)
		require.NoError(t, err, "failed for test %s", name)

		if name == "no-image" {
			assert.NoFileExists(t, filepath.Join(destDir, ".mink.yaml"), "file should not exist for %s", name)
		} else {
			testhelpers.AssertTextFilesEqual(t, filepath.Join(destDir, "expected", ".mink.yaml"), filepath.Join(destDir, ".mink.yaml"), "for test "+name)
			testhelpers.AssertTextFilesEqual(t, filepath.Join(destDir, "expected", "values.yaml"), filepath.Join(destDir, "charts/myapp/values.yaml"), "for test "+name)
		}
	}
}
