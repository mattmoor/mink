/*
Copyright 2019 The Tekton Authors

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
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"go.uber.org/zap"
)

func run(logger *zap.SugaredLogger, dir string, args ...string) (string, error) {
	c := exec.Command("git", args...)
	var output bytes.Buffer
	c.Stderr = &output
	c.Stdout = &output
	// This is the optional working directory. If not set, it defaults to the current
	// working directory of the process.
	if dir != "" {
		c.Dir = dir
	}
	if err := c.Run(); err != nil {
		logger.Errorf("Error running git %v: %v\n%v", args, err, output.String())
		return "", err
	}
	return output.String(), nil
}

// FetchSpec describes how to initialize and fetch from a Git repository.
type FetchSpec struct {
	URL        string
	Revision   string
	Refspec    string
	Path       string
	Depth      uint
	Submodules bool
	SSLVerify  bool
	HTTPProxy  string
	HTTPSProxy string
	NOProxy    string
}

// Fetch fetches the specified git repository at the revision into path, using the refspec to fetch if provided.
func Fetch(logger *zap.SugaredLogger, spec FetchSpec) error {
	if err := ensureHomeEnv(logger); err != nil {
		return err
	}

	if spec.Path != "" {
		if _, err := run(logger, "", "init", spec.Path); err != nil {
			return err
		}
		if err := os.Chdir(spec.Path); err != nil {
			return fmt.Errorf("failed to change directory with path %s; err: %w", spec.Path, err)
		}
	} else if _, err := run(logger, "", "init"); err != nil {
		return err
	}
	trimmedURL := strings.TrimSpace(spec.URL)
	if _, err := run(logger, "", "remote", "add", "origin", trimmedURL); err != nil {
		return err
	}
	if _, err := run(logger, "", "config", "--global", "http.sslVerify", strconv.FormatBool(spec.SSLVerify)); err != nil {
		logger.Warnf("Failed to set http.sslVerify in git config: %s", err)
		return err
	}
	if spec.Revision == "" {
		spec.Revision = "HEAD"
		if _, err := run(logger, "", "symbolic-ref", spec.Revision, "refs/remotes/origin/HEAD"); err != nil {
			return err
		}
	}

	fetchArgs := []string{"fetch"}
	if spec.Submodules {
		fetchArgs = append(fetchArgs, "--recurse-submodules=yes")
	}
	if spec.Depth > 0 {
		fetchArgs = append(fetchArgs, fmt.Sprintf("--depth=%d", spec.Depth))
	}

	// Fetch the revision and verify with FETCH_HEAD
	fetchParam := []string{spec.Revision}
	checkoutParam := "FETCH_HEAD"

	if spec.Refspec != "" {
		// if refspec is specified, fetch the refspec and verify with provided revision
		fetchParam = strings.Split(spec.Refspec, " ")
		checkoutParam = spec.Revision
	}

	// git-init always creates and checks out an empty master branch. When the user requests
	// "master" as the revision, git-fetch will refuse to update the HEAD of the branch it is
	// currently on. The --update-head-ok parameter tells git-fetch that it is ok to update
	// the current (empty) HEAD on initial fetch.
	// The --force parameter tells git-fetch that its ok to update an existing HEAD in a
	// non-fast-forward manner (though this cannot be possible on initial fetch, it can help
	// when the refspec specifies the same destination twice)
	fetchArgs = append(fetchArgs, "origin", "--update-head-ok", "--force")
	fetchArgs = append(fetchArgs, fetchParam...)
	if _, err := run(logger, spec.Path, fetchArgs...); err != nil {
		return fmt.Errorf("failed to fetch %v: %v", fetchParam, err)
	}
	// After performing a fetch, verify that the item to checkout is actually valid
	if _, err := ShowCommit(logger, checkoutParam, spec.Path); err != nil {
		return fmt.Errorf("error parsing %s after fetching refspec %s", checkoutParam, spec.Refspec)
	}

	if _, err := run(logger, "", "checkout", "-f", checkoutParam); err != nil {
		return err
	}

	commit, err := ShowCommit(logger, "HEAD", spec.Path)
	if err != nil {
		return err
	}
	ref, err := ShowRef(logger, "HEAD", spec.Path)
	if err != nil {
		return err
	}
	logger.Infof("Successfully cloned %s @ %s (%s) in path %s", trimmedURL, commit, ref, spec.Path)
	if spec.Submodules {
		if err := SubmoduleFetch(logger, spec); err != nil {
			return err
		}
	}
	return nil
}

func ShowCommit(logger *zap.SugaredLogger, revision, path string) (string, error) {
	output, err := run(logger, path, "show", "-q", "--pretty=format:%H", revision)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(output, "\n"), nil
}

func ShowRef(logger *zap.SugaredLogger, revision, path string) (string, error) {
	output, err := run(logger, path, "show", "-q", "--pretty=format:%D", revision)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(output, "\n"), nil
}

func SubmoduleFetch(logger *zap.SugaredLogger, spec FetchSpec) error {
	if spec.Path != "" {
		if err := os.Chdir(spec.Path); err != nil {
			return fmt.Errorf("failed to change directory with path %s; err: %w", spec.Path, err)
		}
	}
	if _, err := run(logger, "", "submodule", "init"); err != nil {
		return err
	}
	updateArgs := []string{"submodule", "update", "--recursive"}
	if spec.Depth > 0 {
		updateArgs = append(updateArgs, fmt.Sprintf("--depth=%d", spec.Depth))
	}
	if _, err := run(logger, "", updateArgs...); err != nil {
		return err
	}
	logger.Infof("Successfully initialized and updated submodules in path %s", spec.Path)
	return nil
}

func ensureHomeEnv(logger *zap.SugaredLogger) error {
	// HACK: This is to get git+ssh to work since ssh doesn't respect the HOME
	// env variable.
	homepath, err := homedir.Dir()
	if err != nil {
		logger.Errorf("Unexpected error: getting the user home directory: %v", err)
		return err
	}
	homeenv := os.Getenv("HOME")
	euid := os.Geteuid()
	// Special case the root user/directory
	if euid == 0 {
		if err := os.Symlink(homeenv+"/.ssh", "/root/.ssh"); err != nil {
			// Only do a warning, in case we don't have a real home
			// directory writable in our image
			logger.Warnf("Unexpected error: creating symlink: %v", err)
		}
	} else if homeenv != "" && homeenv != homepath {
		if _, err := os.Stat(homepath + "/.ssh"); os.IsNotExist(err) {
			if err := os.Symlink(homeenv+"/.ssh", homepath+"/.ssh"); err != nil {
				// Only do a warning, in case we don't have a real home
				// directory writable in our image
				logger.Warnf("Unexpected error: creating symlink: %v", err)
			}
		}
	}
	return nil
}
