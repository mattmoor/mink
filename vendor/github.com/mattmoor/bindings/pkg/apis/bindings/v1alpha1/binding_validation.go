/*
Copyright 2019 The Knative Authors.

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

package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"knative.dev/pkg/apis"
)

// Validate implements apis.Validatable
func (fb *GithubBinding) Validate(ctx context.Context) *apis.FieldError {
	return fb.Spec.Validate(ctx).ViaField("spec")
}

// Validate implements apis.Validatable
func (fbs *GithubBindingSpec) Validate(ctx context.Context) *apis.FieldError {
	err := fbs.Subject.Validate(ctx).ViaField("subject")

	if fbs.Secret.Name == "" {
		err = err.Also(apis.ErrMissingField("name").ViaField("secret"))
	}
	return err
}

// Validate implements apis.Validatable
func (fb *SlackBinding) Validate(ctx context.Context) *apis.FieldError {
	return fb.Spec.Validate(ctx).ViaField("spec")
}

// Validate implements apis.Validatable
func (fbs *SlackBindingSpec) Validate(ctx context.Context) *apis.FieldError {
	err := fbs.Subject.Validate(ctx).ViaField("subject")

	if fbs.Secret.Name == "" {
		err = err.Also(apis.ErrMissingField("name").ViaField("secret"))
	}
	return err
}

// Validate implements apis.Validatable
func (fb *TwitterBinding) Validate(ctx context.Context) *apis.FieldError {
	return fb.Spec.Validate(ctx).ViaField("spec")
}

// Validate implements apis.Validatable
func (fbs *TwitterBindingSpec) Validate(ctx context.Context) *apis.FieldError {
	err := fbs.Subject.Validate(ctx).ViaField("subject")

	if fbs.Secret.Name == "" {
		err = err.Also(apis.ErrMissingField("name").ViaField("secret"))
	}
	return err
}

// Validate implements apis.Validatable
func (fb *GoogleCloudSQLBinding) Validate(ctx context.Context) *apis.FieldError {
	return fb.Spec.Validate(ctx).ViaField("spec")
}

// Validate implements apis.Validatable
func (fbs *GoogleCloudSQLBindingSpec) Validate(ctx context.Context) *apis.FieldError {
	err := fbs.Subject.Validate(ctx).ViaField("subject")

	if fbs.Secret.Name == "" {
		err = err.Also(apis.ErrMissingField("name").ViaField("secret"))
	}

	if fbs.Instance == "" {
		err = err.Also(apis.ErrMissingField("instance"))
	}
	// Some rudimentary validation, since I don't know the exact schema.
	parts := strings.Split(fbs.Instance, ":")
	if len(parts) != 3 {
		err = err.Also(apis.ErrInvalidValue(
			fmt.Sprintf(`expected "project:region:name", but got %q`, fbs.Instance), "instance"))
	}
	return err
}

// Validate implements apis.Validatable
func (fb *SQLBinding) Validate(ctx context.Context) *apis.FieldError {
	return fb.Spec.Validate(ctx).ViaField("spec")
}

// Validate implements apis.Validatable
func (fbs *SQLBindingSpec) Validate(ctx context.Context) *apis.FieldError {
	err := fbs.Subject.Validate(ctx).ViaField("subject")

	if fbs.Secret.Name == "" {
		err = err.Also(apis.ErrMissingField("name").ViaField("secret"))
	}
	return err
}
