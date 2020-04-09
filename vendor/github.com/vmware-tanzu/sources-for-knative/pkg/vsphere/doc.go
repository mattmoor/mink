/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

// Package vsphere holds utilities for bootstrapping a vSphere API client
// from the metadata injected by the VSphereSource.  Within a receive adapter,
// users can write:
//    client, err := vsphere.New(ctx)
// This is modeled after the Bindings pattern.
package vsphere
