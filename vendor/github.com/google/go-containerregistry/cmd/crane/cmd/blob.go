// Copyright 2020 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"io"
	"log"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/spf13/cobra"
)

func init() { Root.AddCommand(NewCmdBlob()) }

// NewCmdBlob creates a new cobra.Command for the blob subcommand.
func NewCmdBlob() *cobra.Command {
	return &cobra.Command{
		Use:     "blob BLOB",
		Short:   "Read a blob from the registry",
		Example: "crane blob ubuntu@sha256:4c1d20cdee96111c8acf1858b62655a37ce81ae48648993542b7ac363ac5c0e5 > blob.tar.gz",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			src := args[0]
			layer, err := crane.PullLayer(src, options...)
			if err != nil {
				log.Fatalf("pulling layer %s: %v", src, err)
			}
			blob, err := layer.Compressed()
			if err != nil {
				log.Fatalf("fetching blob %s: %v", src, err)
			}
			if _, err := io.Copy(cmd.OutOrStdout(), blob); err != nil {
				log.Fatalf("copying blob %s: %v", src, err)
			}
		},
	}
}
