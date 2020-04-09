/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	"context"

	"knative.dev/pkg/apis"
)

// Validate implements apis.Validatable
func (fb *VSphereSource) Validate(ctx context.Context) *apis.FieldError {
	return fb.Spec.Validate(ctx).ViaField("spec")
}

// Validate implements apis.Validatable
func (fbs *VSphereSourceSpec) Validate(ctx context.Context) *apis.FieldError {
	return fbs.Sink.Validate(ctx).ViaField("sink").Also(fbs.VAuthSpec.Validate(ctx))
}
