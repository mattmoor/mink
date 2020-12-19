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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dprotaso/go-yit"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/mattmoor/mink/pkg/builds"
	"github.com/mattmoor/mink/pkg/builds/ko"
	minkcli "github.com/mattmoor/mink/pkg/cli"
	"github.com/mattmoor/mink/pkg/constants"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/cli/pkg/options"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/pool"
)

var resolveExample = fmt.Sprintf(`
  # Build and publish references within foo.yaml, and print a new version to STDOUT.
  %[1]s resolve -f config/foo.yaml

  # Build and publish references within yaml files recursively under config/, and
  # print a new version to STDOUT.
  %[1]s resolve -Rf config/

  # Customize the builder used for buildpack:/// builds
  %[1]s resolve -f config/ --builder foo:latest

  # Customize the name of the overrides file used for buildpack:/// builds
  %[1]s resolve -f config/ --overrides another-name.toml

  # Customize the name of Dockerfiles to use for dockerfile:/// builds
  %[1]s resolve -f config/ --dockerfile Dockerfile.production`, ExamplePrefix())

// NewResolveCommand implements 'kn-im resolve' command
func NewResolveCommand(ctx context.Context) *cobra.Command {
	opts := &ResolveOptions{
		BaseBuildOptions: BaseBuildOptions{BundleOptions: BundleOptions{ctx: ctx}},
	}

	cmd := &cobra.Command{
		Use:     "resolve -f FILE",
		Short:   "Build publish and resolve image references within a collection of yaml files.",
		Example: resolveExample,
		PreRunE: opts.Validate,
		RunE:    opts.Execute,
	}

	opts.AddFlags(cmd)

	return cmd
}

type builder func(context.Context, name.Digest, *url.URL) (name.Digest, error)

// ResolveOptions implements Interface for the `kn im resolve` command.
type ResolveOptions struct {
	// Inherit all of the base build options.
	BaseBuildOptions

	// Inherit the dockerfile and buildpack options.
	dockerfileOptions
	buildpackOptions

	Filenames []string
	Recursive bool

	Parallelism int

	builders map[string]builder
}

// ResolveOptions implements Interface
var _ Interface = (*ResolveOptions)(nil)

// AddFlags implements Interface
func (opts *ResolveOptions) AddFlags(cmd *cobra.Command) {
	// Add the bundle flags to our surface.
	opts.BaseBuildOptions.AddFlags(cmd)
	opts.dockerfileOptions.AddFlags(cmd)
	opts.buildpackOptions.AddFlags(cmd)

	// Based on the same flags in kubectl / ko
	cmd.Flags().StringSliceP("filename", "f", nil,
		"Filename, directory, or URL to files to use to create the resource")
	cmd.Flags().BoolP("recursive", "R", false,
		"Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.")
	cmd.Flags().IntP("parallelism", "P", 20, "How many parallel builds to run at once.")
}

// Validate implements Interface
func (opts *ResolveOptions) Validate(cmd *cobra.Command, args []string) error {
	// Validate the bundle arguments.
	if err := opts.BaseBuildOptions.Validate(cmd, args); err != nil {
		return err
	}
	if err := opts.dockerfileOptions.Validate(cmd, args); err != nil {
		return err
	}
	if err := opts.buildpackOptions.Validate(cmd, args); err != nil {
		return err
	}

	opts.Filenames = viper.GetStringSlice("filename")
	if len(opts.Filenames) == 0 {
		return minkcli.ErrMissingFlag("filename")
	}
	opts.Recursive = viper.GetBool("recursive")

	opts.Parallelism = viper.GetInt("parallelism")
	if opts.Parallelism <= 0 {
		return minkcli.ErrInvalidValue("parallelism",
			"must be greater than 0, but got: %d", opts.Parallelism)
	}

	opts.builders = map[string]builder{
		"dockerfile": opts.db,
		"buildpack":  opts.bp,
		"ko":         opts.ko,
		"task":       opts.task,
		"pipeline":   opts.pipeline,
	}

	return nil
}

// Execute implements Interface
func (opts *ResolveOptions) Execute(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("'im bundle' does not take any arguments")
	}

	// Handle ctrl+C
	return opts.execute(opts.GetContext(cmd), cmd)
}

// execute is the workhorse of execute, but factored to support composition
// with apply (provides its own ctx)
func (opts *ResolveOptions) execute(ctx context.Context, cmd *cobra.Command) error {
	// Bundle up the source context in an image.
	sourceDigest, err := opts.bundle(ctx)
	if err != nil {
		return err
	}

	// Turn the files into yaml nodes.
	files := opts.EnumerateFiles()
	blocks := make([]*yaml.Node, 0, len(files))
	for _, f := range files {
		bs, err := opts.ResolveFile(ctx, f)
		if err != nil {
			return err
		}
		blocks = append(blocks, bs...)
	}

	// Turn all of the images references in the yaml nodes into digests.
	if err := opts.ResolveReferences(ctx, blocks, sourceDigest); err != nil {
		return err
	}

	// Encode the resulting yaml
	e := yaml.NewEncoder(cmd.OutOrStdout())
	e.SetIndent(2)
	for _, doc := range blocks {
		if err := e.Encode(doc); err != nil {
			return fmt.Errorf("failed to encode output: %w", err)
		}
	}
	e.Close()
	return nil
}

// EnumerateFiles is based heavily on pkg/kubectl
func (opts *ResolveOptions) EnumerateFiles() (files []string) {
	seen := sets.NewString()
	for _, paths := range opts.Filenames {
		// Just pass through '-' as it is indicative of stdin.
		if paths == "-" {
			files = append(files, paths)
			continue
		}
		// For each of the "filenames" we are passed (file or directory) start a
		// "Walk" to enumerate all of the contained files recursively.
		err := filepath.Walk(paths, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if seen.Has(path) {
				return nil
			}

			// If this is a directory, skip it if it isn't the current directory we are
			// processing (unless we are in recursive mode).
			if fi.IsDir() {
				if path != paths && !opts.Recursive {
					return filepath.SkipDir
				}
				return nil
			}

			// Don't check extension if the filepath was passed explicitly
			if path != paths {
				switch filepath.Ext(path) {
				case ".json", ".yaml":
					// Process these.
				default:
					return nil
				}
			}

			files = append(files, path)
			seen.Insert(path)
			return nil
		})
		if err != nil {
			log.Fatal("Error enumerating files:", err)
		}
	}
	return files
}

// ResolveFile is based heavily on ko's resolveFile
func (opts *ResolveOptions) ResolveFile(ctx context.Context, f string) (blocks []*yaml.Node, err error) {
	var b []byte
	if f == "-" {
		b, err = ioutil.ReadAll(os.Stdin)
	} else {
		b, err = ioutil.ReadFile(f)
	}
	if err != nil {
		return nil, err
	}

	// The loop is to support multi-document yaml files.
	// This is handled by using a yaml.Decoder and reading objects until io.EOF, see:
	// https://godoc.org/gopkg.in/yaml.v3#Decoder.Decode
	decoder := yaml.NewDecoder(bytes.NewBuffer(b))
	for {
		var doc yaml.Node
		if err := decoder.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		blocks = append(blocks, &doc)
	}

	return
}

// ResolveReferences is based heavily on ko's ImageReferences
func (opts *ResolveOptions) ResolveReferences(ctx context.Context, docs []*yaml.Node, source name.Digest) error {
	// First, walk the input objects and collect a list of supported references
	refs := make(map[string][]*yaml.Node)

	for _, doc := range docs {
		it := opts.refsFromDoc(doc)

		for node, ok := it(); ok; node, ok = it() {
			ref := strings.TrimSpace(node.Value)
			refs[ref] = append(refs[ref], node)
		}
	}

	errg, ctx := pool.NewWithContext(ctx, opts.Parallelism, opts.Parallelism)

	// Next, perform parallel builds for each of the supported references.
	var sm sync.Map
	for ref := range refs {
		ref := ref

		// Parse the reference and use the scheme to determine
		// the builder to apply.
		u, err := url.Parse(ref)
		if err != nil {
			return err
		}
		builder, ok := opts.builders[u.Scheme]
		if !ok {
			continue
		}

		errg.Go(func() error {
			digest, err := builder(ctx, source, u)
			if err != nil {
				return err
			}
			sm.Store(ref, digest.String())
			return nil
		})
	}
	if err := errg.Wait(); err != nil {
		return err
	}

	// Walk the tags and update them with their digest.
	for ref, nodes := range refs {
		digest, ok := sm.Load(ref)

		if !ok {
			return fmt.Errorf("resolved reference to %q not found", ref)
		}

		for _, node := range nodes {
			node.Value = digest.(string)
		}
	}
	return nil
}

func (opts *ResolveOptions) db(ctx context.Context, source name.Digest, u *url.URL) (name.Digest, error) {
	if u.Host != "" {
		return name.Digest{}, fmt.Errorf(
			"unexpected host in %q reference, got: %s (did you mean %s:/// instead of %s://?)",
			u.Scheme, u.Host, u.Scheme, u.Scheme)
	}

	// Create the equivalent `mink build` invocation.
	bo := BuildOptions{
		BaseBuildOptions:  opts.BaseBuildOptions,
		dockerfileOptions: opts.dockerfileOptions,
	}
	bo.Dockerfile = filepath.Join(u.Path, opts.Dockerfile)

	// Buffer the output, so we can display it on failures.
	buf := &bytes.Buffer{}

	// Run the produced Build definition to completion, streaming logs to stdout, and
	// returning the digest of the produced image.
	digest, err := bo.build(ctx, source, buf)
	if err != nil {
		log.Print(buf.String())
		return name.Digest{}, err
	}
	return digest, nil
}

func (opts *ResolveOptions) bp(ctx context.Context, source name.Digest, u *url.URL) (name.Digest, error) {
	if u.Host != "" {
		return name.Digest{}, fmt.Errorf(
			"unexpected host in %q reference, got: %s (did you mean %s:/// instead of %s://?)",
			u.Scheme, u.Host, u.Scheme, u.Scheme)
	}

	// Create the equivalent `mink buildpack` invocation.
	bpo := BuildpackOptions{
		BaseBuildOptions: opts.BaseBuildOptions,
		buildpackOptions: opts.buildpackOptions,
	}
	bpo.DescriptorFile = filepath.Join(u.Path, opts.DescriptorFile)

	// Buffer the output, so we can display it on failures.
	buf := &bytes.Buffer{}

	digest, err := bpo.build(ctx, source, buf)
	if err != nil {
		log.Print(buf.String())
		return name.Digest{}, err
	}
	return digest, nil
}

func (opts *ResolveOptions) task(ctx context.Context, source name.Digest, u *url.URL) (name.Digest, error) {
	// Create the equivalent `mink build` invocation.
	bo := &RunTaskOptions{
		RunOptions: RunOptions{
			BaseBuildOptions: opts.BaseBuildOptions,
			resource:         u.Scheme,
		},
	}

	return opts.run(ctx, source, u, &bo.RunOptions, bo.buildCmd)
}

func (opts *ResolveOptions) pipeline(ctx context.Context, source name.Digest, u *url.URL) (name.Digest, error) {
	// Create the equivalent `mink build` invocation.
	bo := &RunPipelineOptions{
		RunOptions: RunOptions{
			BaseBuildOptions: opts.BaseBuildOptions,
			resource:         u.Scheme,
		},
	}

	return opts.run(ctx, source, u, &bo.RunOptions, bo.buildCmd)
}

type buildCommander func(context.Context, string, signatureDetector) (*cobra.Command, error)

func (opts *ResolveOptions) run(ctx context.Context, source name.Digest, u *url.URL, bo *RunOptions, bc buildCommander) (name.Digest, error) {
	// TODO(mattmoor): Introduce an optional duck for this?
	if u.Path != "" {
		return name.Digest{}, fmt.Errorf(
			"unexpected path in %q reference, got: %s",
			u.Scheme, u.Path)
	}

	var digest name.Digest

	// We take one positional argument, pass that as the task name.
	taskCmd, err := bc(ctx, u.Host, func(cmd *cobra.Command, params []v1beta1.ParamSpec, results sets.String) []Processor {
		paramNames := make(sets.String, len(params))
		for _, param := range params {
			paramNames.Insert(param.Name)
		}

		requiredResults := sets.NewString(constants.ImageDigestResult)
		requiredParams := sets.NewString(constants.SourceBundleParam, constants.ImageTargetParam)

		missingResults := requiredResults.Difference(results)
		missingParams := requiredParams.Difference(paramNames)

		switch {
		case len(missingParams) > 0 && len(missingResults) > 0:
			return []Processor{ValidationErrorProcessor(
				"%s %q is missing required parameter(s): %v and result(s): %v",
				u.Scheme, u.Host, missingParams, missingResults,
			)}
		case len(missingParams) > 0:
			return []Processor{ValidationErrorProcessor(
				"%s %q is missing required parameter(s): %v",
				u.Scheme, u.Host, missingParams,
			)}
		case len(missingResults) > 0:
			return []Processor{ValidationErrorProcessor(
				"%s %q is missing required result(s): %v",
				u.Scheme, u.Host, missingResults,
			)}
		}

		// TODO(mattmoor): Consider an optional duck to pass through the path part (vs. requiring querystring)

		var tag name.Tag
		return []Processor{processParams(cmd, params), &ProcessorFuncs{
			PreRunFunc: func(params []v1beta1.ParamSpec) ([]v1beta1.Param, error) {
				// This was uploaded for us.
				bo.references = append(bo.references, source)

				var err error
				tag, err = bo.tag(imageNameContext{URL: *u})
				if err != nil {
					return nil, err
				}
				bo.references = append(bo.references, tag)

				return []v1beta1.Param{{
					Name:  constants.SourceBundleParam,
					Value: *v1beta1.NewArrayOrString(source.String()),
				}, {
					Name:  constants.ImageTargetParam,
					Value: *v1beta1.NewArrayOrString(tag.String()),
				}}, nil
			},
			PostRunFunc: func(results []v1beta1.TaskRunResult) (err error) {
				for _, r := range results {
					if r.Name != constants.ImageDigestResult {
						continue
					}
					digest, err = name.NewDigest(tag.String() + "@" + r.Value)
					return err
				}
				return fmt.Errorf("unable to find result %q", constants.ImageDigestResult)
			},
		}}
	})
	if err != nil {
		return name.Digest{}, err
	}

	// Pass the querystring as args to the task.
	args := make([]string, 0, len(u.Query()))
	for k, vs := range u.Query() {
		for _, v := range vs {
			args = append(args, fmt.Sprintf("--%s=%s", k, v))
		}
	}
	taskCmd.SetArgs(args)

	// Buffer the output, so we can display it on failures.
	buf := &bytes.Buffer{}
	taskCmd.SetOutput(buf)

	if err := taskCmd.Execute(); err != nil {
		log.Print(buf.String())
		return name.Digest{}, err
	}

	return digest, nil
}

func (opts *ResolveOptions) ko(ctx context.Context, source name.Digest, u *url.URL) (name.Digest, error) {
	tag, err := opts.tag(imageNameContext{
		URL: *u,
	})
	if err != nil {
		return name.Digest{}, err
	}

	tr := ko.Build(ctx, source, tag, ko.Options{
		ImportPath: u.String(),
	})
	tr.Namespace = Namespace()

	// Buffer the output, so we can display it on failures.
	buf := &bytes.Buffer{}

	// Run the produced Build definition to completion, streaming logs to stdout, and
	// returning the digest of the produced image.
	digest, err := builds.Run(ctx, tag.String(), tr, &options.LogOptions{
		ActivityTimeout: activityTimeout,
		Params:          &cli.TektonParams{},
		Stream: &cli.Stream{
			// Send Out to stderr so we can capture the digest for composition.
			Out: buf,
			Err: buf,
		},
		Follow: true,
	}, builds.WithTaskServiceAccount(ctx, opts.ServiceAccount, tag, source))
	if err != nil {
		log.Print(buf.String())
		return name.Digest{}, err
	}
	return digest, nil
}

func (opts *ResolveOptions) refsFromDoc(doc *yaml.Node) yit.Iterator {
	ps := make([]yit.Predicate, 0, len(opts.builders))

	for k := range opts.builders {
		ps = append(ps, yit.WithPrefix(k+"://"))
	}

	return yit.FromNode(doc).
		RecurseNodes().
		Filter(yit.StringValue).
		Filter(yit.Union(ps...))
}
