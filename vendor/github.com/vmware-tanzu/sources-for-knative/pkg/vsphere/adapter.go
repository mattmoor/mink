/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package vsphere

import (
	"context"
	"fmt"
	"reflect"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/kvstore"
	"knative.dev/pkg/logging"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
)

type envConfig struct {
	adapter.EnvConfig

	// The name of the configmap to use as our kvstore.
	KVConfigMap string `envconfig:"VSPHERE_KVSTORE_CONFIGMAP" required:"true"`
}

func NewEnvConfig() adapter.EnvConfigAccessor {
	return &envConfig{}
}

// vAdapter implements the vSphereSource adapter to trigger a Sink.
type vAdapter struct {
	Logger    *zap.SugaredLogger
	Namespace string
	Source    string
	VClient   *govmomi.Client
	CEClient  cloudevents.Client
	KVStore   kvstore.Interface
}

func NewAdapter(ctx context.Context, processed adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
	env := processed.(*envConfig)

	logger := logging.FromContext(ctx)

	vClient, err := New(ctx)
	if err != nil {
		logger.Fatalf("Unable to create vSphere client: %v", err)
	}

	source, err := Address(ctx)
	if err != nil {
		logger.Fatalf("Unable to determine source: %v", err)
	}

	store := kvstore.NewConfigMapKVStore(ctx, env.KVConfigMap, env.Namespace, kubeclient.Get(ctx).CoreV1())
	err = store.Init(ctx)
	if err != nil {
		logger.Fatalf("couldn't initialize kv store: %v", err)
	}

	return &vAdapter{
		Logger:    logger,
		Namespace: env.Namespace,
		Source:    source,
		VClient:   vClient,
		CEClient:  ceClient,
		KVStore:   store,
	}
}

// Start implements adapter.Adapter
func (a *vAdapter) Start(ctx context.Context) error {
	manager := event.NewManager(a.VClient.Client)

	managedTypes := []types.ManagedObjectReference{a.VClient.ServiceContent.RootFolder}
	return manager.Events(ctx, managedTypes, 1, true /* tail */, false /* force */, a.sendEvents(ctx))
}

func (a *vAdapter) sendEvents(ctx context.Context) func(moref types.ManagedObjectReference, baseEvents []types.BaseEvent) error {
	return func(moref types.ManagedObjectReference, baseEvents []types.BaseEvent) error {
		for _, be := range baseEvents {
			event := cloudevents.NewEvent(cloudevents.VersionV1)

			event.SetType("com.vmware.vsphere." + reflect.TypeOf(be).Elem().Name())
			event.SetTime(be.GetEvent().CreatedTime)
			event.SetID(fmt.Sprintf("%d", be.GetEvent().Key))
			event.SetSource(a.Source)

			switch e := be.(type) {
			case *types.EventEx:
				event.SetExtension("EventEx", e)
			case *types.ExtendedEvent:
				event.SetExtension("ExtendedEvent", e)
			}
			// TODO(mattmoor): Consider setting the subject

			if err := event.SetData(cloudevents.ApplicationXML, be); err != nil {
				logging.FromContext(ctx).Errorw("failed to set data on event", zap.Error(err))
			}

			result := a.CEClient.Send(ctx, event)
			if !cloudevents.IsACK(result) {
				a.Logger.Errorw("failed to send cloudevent", zap.Error(result))
				return result
			}
		}

		return nil
	}
}
