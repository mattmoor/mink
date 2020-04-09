/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/vmware-tanzu/sources-for-knative/pkg/vsphere"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/types"
)

type receiver struct {
	manager *tags.Manager
}

func main() {
	ctx := context.Background()

	ceclient, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Instantiate a client for interacting with the vSphere APIs.
	client, err := vsphere.NewREST(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	r := &receiver{manager: tags.NewManager(client)}

	if err := ceclient.StartReceiver(ctx, r.handle); err != nil {
		log.Fatal(err)
	}
}

func (r *receiver) handle(ctx context.Context, event cloudevents.Event) error {
	req := &types.VmCreatedEvent{}
	if err := event.DataAs(&req); err != nil {
		return err
	}
	log.Printf("Tagging VM: %v", req.Vm.Vm)
	return r.manager.AttachTag(ctx, "shrug", req.Vm.Vm)
}
