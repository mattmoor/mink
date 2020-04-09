/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	"context"

	"knative.dev/pkg/apis"
)

// SetDefaults implements apis.Defaultable
func (as *VSphereSource) SetDefaults(ctx context.Context) {
	withNS := apis.WithinParent(ctx, as.ObjectMeta)
	as.Spec.Sink.SetDefaults(withNS)
}
