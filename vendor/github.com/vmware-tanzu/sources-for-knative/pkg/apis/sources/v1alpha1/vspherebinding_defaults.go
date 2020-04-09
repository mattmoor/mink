/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	"context"
)

// SetDefaults implements apis.Defaultable
func (as *VSphereBinding) SetDefaults(ctx context.Context) {
	if as.Spec.Subject.Namespace == "" {
		// Default the subject's namespace to our namespace.
		as.Spec.Subject.Namespace = as.Namespace
	}
}
