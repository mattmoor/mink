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
func (fb *VSphereBinding) Validate(ctx context.Context) *apis.FieldError {
	err := fb.Spec.Validate(ctx).ViaField("spec")
	if fb.Spec.Subject.Namespace != "" && fb.Namespace != fb.Spec.Subject.Namespace {
		err = err.Also(apis.ErrInvalidValue(fb.Spec.Subject.Namespace, "spec.subject.namespace"))
	}
	return err
}

// Validate implements apis.Validatable
func (fbs *VSphereBindingSpec) Validate(ctx context.Context) *apis.FieldError {
	return fbs.Subject.Validate(ctx).ViaField("subject").Also(fbs.VAuthSpec.Validate(ctx))
}

// Validate implements apis.Validatable
func (vas *VAuthSpec) Validate(ctx context.Context) (err *apis.FieldError) {
	if vas.Address.Host == "" {
		err = err.Also(apis.ErrMissingField("address.host"))
	}
	if vas.SecretRef.Name == "" {
		err = err.Also(apis.ErrMissingField("secretRef.name"))
	}
	return err
}
