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

package kontext

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"

	"knative.dev/pkg/pool"
)

const (
	// StoragePath is where in the container image the files are placed.
	StoragePath = "/var/run/kontext"
)

func copy(src, dest string) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	return err
}

func expand(ctx context.Context, base string) error {
	targetPath, err := os.Getwd()
	if err != nil {
		return err
	}

	eg, ctx := pool.NewWithContext(ctx, 1000 /* workers */, 1000 /* capacity */)
	if err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == base {
			return nil
		}

		// Add each file to the backlog.
		eg.Go(func() error {
			// If the context is canceled, then bail out early.
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			relativePath := path[len(base)+1:]
			target := filepath.Join(targetPath, relativePath)

			if info.IsDir() {
				return os.MkdirAll(target, os.ModePerm)
			}
			if !info.Mode().IsRegular() {
				log.Printf("Skipping irregular file: %q", relativePath)
				return nil
			}
			if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
				return err
			}
			return copy(path, target)
		})

		return nil
	}); err != nil {
		return err
	}

	// Wait for the work to be done.
	return eg.Wait()
}

// Expand recursively copies the current working directory into StoragePath.
func Expand(ctx context.Context) error {
	return expand(ctx, StoragePath)
}
