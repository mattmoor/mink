/*
Copyright 2021 The Knative Authors

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

package main

import (
	"context"
	"sync"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/logging"

	tkndefaultconfig "github.com/tektoncd/pipeline/pkg/apis/config"
	knedefaultconfig "knative.dev/eventing/pkg/apis/config"
	"knative.dev/eventing/pkg/apis/feature"
	channeldefaultconfig "knative.dev/eventing/pkg/apis/messaging/config"
	pingdefaultconfig "knative.dev/eventing/pkg/apis/sources/config"
	knsdefaultconfig "knative.dev/serving/pkg/apis/config"
)

var (
	decorator func(context.Context) context.Context
	decoSetup sync.Once
)

func newContextDecorator(ctx context.Context, cmw configmap.Watcher) func(ctx context.Context) context.Context {
	// We don't need a copy of this for every webhook, so just set it up
	// once, and return that singleton to each of our webhooks.
	decoSetup.Do(func() {
		// Decorate contexts with the current state of the config.
		knsstore := knsdefaultconfig.NewStore(logging.FromContext(ctx).Named("kns-config-store"))
		knsstore.WatchConfigs(cmw)

		knestore := knedefaultconfig.NewStore(logging.FromContext(ctx).Named("kne-config-store"))
		knestore.WatchConfigs(cmw)

		tknstore := tkndefaultconfig.NewStore(logging.FromContext(ctx).Named("tkn-config-store"))
		tknstore.WatchConfigs(cmw)

		channelStore := channeldefaultconfig.NewStore(logging.FromContext(ctx).Named("channel-config-store"))
		channelStore.WatchConfigs(cmw)

		pingstore := pingdefaultconfig.NewStore(logging.FromContext(ctx).Named("ping-config-store"))
		pingstore.WatchConfigs(cmw)

		featureStore := feature.NewStore(logging.FromContext(ctx).Named("feature-config-store"))
		featureStore.WatchConfigs(cmw)

		decorator = func(ctx context.Context) context.Context {
			ctx = knsstore.ToContext(ctx)
			ctx = knestore.ToContext(ctx)
			ctx = tknstore.ToContext(ctx)
			ctx = channelStore.ToContext(ctx)
			ctx = featureStore.ToContext(ctx)
			return ctx
		}
	})

	return decorator
}
