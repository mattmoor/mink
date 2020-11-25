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

package command

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	errs "github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	minkFileName = ".mink.yaml"
)

var initExample = fmt.Sprintf(`
  # Creates a .mink.yaml file if one does not already exist.
  %[1]s init
`, ExamplePrefix())

// NewInitCommand implements 'kn-im init' command
func NewInitCommand() *cobra.Command {
	opts := &InitOptions{}

	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Creates a default .mink.yaml file if one does not already exist.",
		Example: initExample,
		PreRunE: opts.Validate,
		RunE:    opts.Execute,
	}

	opts.AddFlags(cmd)

	return cmd
}

// AddFlags implements Interface
func (opts *InitOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("dir", ".", "The directory to look inside and generate the the .mink.yaml file.")
	cmd.Flags().String("dockerfile", "Dockerfile", "The path to the Dockerfile within the build context.")
	cmd.Flags().StringSlice("kaniko-args", nil, "Optional arguments to pass to kaniko for dealing with insecure registries. For details see: https://github.com/GoogleContainerTools/kaniko/blob/master/README.md#additional-flags")
	cmd.Flags().Bool("no-git", false, "Disables adding of the generated .mink.yaml file to git if it is generated")
}

// Validate implements Interface
func (opts *InitOptions) Validate(cmd *cobra.Command, args []string) error {
	opts.Dir = viper.GetString("dir")
	opts.Dockerfile = viper.GetString("dockerfile")
	opts.NoGit = viper.GetBool("no-git")
	return nil
}

// InitOptions implements Interface for the `kn im build` command.
type InitOptions struct {

	// Dockerfile is the default relative path to the Dockerfile within the directory to look for
	Dockerfile string

	// Dir is the directory to look inside for chart folders
	Dir string

	// NoGit disables adding the generated .mink.yaml to git
	NoGit bool

	// MinkEnabled returns the result if a .mink.yaml was generated
	MinkEnabled bool

	// Out the output destination
	Out io.Writer
}

// InitOptions implements Interface
var _ Interface = (*InitOptions)(nil)

// Execute implements Interface
func (opts *InitOptions) Execute(cmd *cobra.Command, args []string) error {
	if opts.Out == nil {
		opts.Out = cmd.OutOrStdout()
	}
	chartDirs, err := opts.findHelmChartDirs()
	if err != nil {
		return errs.Wrapf(err, "failed to find charts")
	}
	if len(chartDirs) == 0 {
		return nil
	}

	image, err := opts.findMinkImage()
	if err != nil {
		return errs.Wrapf(err, "failed to find mink image string")
	}
	if image == "" {
		return nil
	}

	// lets add an image ref to the first chart
	err = opts.addImageToValuesFile(image, chartDirs[0])
	if err != nil {
		return errs.Wrapf(err, "failed to add image to values file")
	}

	minkFile := filepath.Join(opts.Dir, minkFileName)
	exists, err := files.FileExists(minkFile)
	if err != nil {
		return errs.Wrapf(err, "failed to check if file exists %s", minkFile)
	}
	if exists {
		opts.MinkEnabled = true
		return nil
	}
	err = opts.createMinkFile(minkFile, chartDirs)
	if err != nil {
		return errs.Wrapf(err, "failed to ")
	}
	opts.MinkEnabled = true

	if !opts.NoGit {
		err = opts.AddToGit(opts.Dir, minkFileName)
		if err != nil {
			return errs.Wrapf(err, "failed to add mink file to git")
		}
	}
	return nil
}

func (opts *InitOptions) findHelmChartDirs() ([]string, error) {
	var dirs []string
	err := filepath.Walk(opts.Dir, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		name := info.Name()
		if name != "Chart.yaml" {
			return nil
		}
		dir := filepath.Dir(path)
		dirs = append(dirs, dir)
		return nil
	})
	if err != nil {
		return dirs, errs.Wrapf(err, "failed to find chart directories")
	}
	return dirs, nil
}

func (opts *InitOptions) createMinkFile(file string, dirs []string) error {
	buf := strings.Builder{}
	buf.WriteString("# the files containing a mink image build URI such as dockerfile:/// ko:/// or buildpack:///\n")
	buf.WriteString("filename:\n")
	for _, d := range dirs {
		buf.WriteString("- ")
		buf.WriteString(d)
		buf.WriteString("\n")
	}
	buf.WriteString("\n")

	text := buf.String()

	err := ioutil.WriteFile(file, []byte(text), files.DefaultFileWritePermissions)
	if err != nil {
		return errs.Wrapf(err, "failed to save file %s", file)
	}
	fmt.Fprintf(opts.Out, "created file %s\n", file)
	return nil
}

func (opts *InitOptions) findMinkImage() (string, error) {
	// check for a Dockerfile
	f := filepath.Join(opts.Dir, opts.Dockerfile)
	exists, err := files.FileExists(f)
	if err != nil {
		return "", errs.Wrapf(err, "failed to check if file exists %s", f)
	}
	if exists {
		return "dockerfile:///", nil
	}

	// check for build pack
	f = filepath.Join(opts.Dir, "overrides.toml")
	exists, err = files.FileExists(f)
	if err != nil {
		return "", errs.Wrapf(err, "failed to check if file exists %s", f)
	}
	if exists {
		return "buildpack:///", nil
	}
	// TODO detect ko
	return "", nil
}

func (opts *InitOptions) addImageToValuesFile(image string, dir string) error {
	f := filepath.Join(dir, "values.yaml")
	exists, err := files.FileExists(f)
	if err != nil {
		return errs.Wrapf(err, "failed to check if file exists %s", f)
	}
	if !exists {
		return nil
	}

	node, err := yaml.ReadFile(f)
	if err != nil {
		return errs.Wrapf(err, "failed to load file %s", f)
	}

	v, err := node.Pipe(yaml.Lookup("image", "fullName"))
	if err != nil {
		return errs.Wrapf(err, "failed to lookup image.fullName")
	}
	if v != nil {
		text, err := v.String()
		if err != nil {
			return errs.Wrapf(err, "failed to get text for image.fullName")
		}
		if text == image {
			return nil
		}
	}

	err = node.PipeE(yaml.LookupCreate(yaml.ScalarNode, "image", "fullName"), yaml.FieldSetter{StringValue: image})
	if err != nil {
		return errs.Wrapf(err, "failed to set image.fullName to %s", image)
	}
	err = yaml.WriteFile(node, f)
	if err != nil {
		return errs.Wrapf(err, "failed to save file %s", f)
	}
	fmt.Fprintf(opts.Out, "added image %s to file %s\n", image, f)
	return nil
}

// AddToGit adds the file to git in the given dir
func (opts *InitOptions) AddToGit(dir string, name string) error {
	r, err := git.PlainOpen(dir)
	if err != nil {
		return errs.Wrapf(err, "failed to open git dir %s", dir)
	}

	t, err := r.Worktree()
	if err != nil {
		return errs.Wrapf(err, "failed to get work tree")
	}

	_, err = t.Add(name)
	if err != nil {
		return errs.Wrapf(err, "failed to add file %s to git", name)
	}
	fmt.Fprintf(opts.Out, "added file %s to git\n", name)
	return nil
}
