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
	"context"
	"time"

	"github.com/spf13/cobra"
)

// Interface defines the interface that various options types are expected to adhere to.
type Interface interface {
	// GetContext fetches the context for the given command.
	GetContext(cmd *cobra.Command) context.Context

	// AddFlags populates the flags needed by this command.
	AddFlags(cmd *cobra.Command)

	// Validate checks the command line arguments and configuration
	// prior to executing the command.
	Validate(cmd *cobra.Command, args []string) error

	// Execute runs the command.
	Execute(cmd *cobra.Command, args []string) error
}

// activityTimeout is the amount of time to wait for a run to show activity before timing out.
const activityTimeout = 30 * time.Second
