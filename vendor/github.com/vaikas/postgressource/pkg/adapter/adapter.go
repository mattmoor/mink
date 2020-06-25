/*
Copyright 2019 The Knative Authors
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

// Package adapter implements a sample receive adapter that generates events
// at a regular interval.
package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "database/sql"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	bindingsql "github.com/mattmoor/bindings/pkg/sql"
	"go.uber.org/zap"

	"github.com/lib/pq"
	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"
)

type envConfig struct {
	// Include the standard adapter.EnvConfig used by all adapters.
	adapter.EnvConfig

	// The name of the channel where our events get sent to.
	NotificationChannel string `envconfig:"NOTIFICATION_CHANNEL" required:"true"`
}

func NewEnv() adapter.EnvConfigAccessor { return &envConfig{} }

// Adapter generates events at a regular interval.
type Adapter struct {
	client   cloudevents.Client
	logger   *zap.SugaredLogger
	listener *pq.Listener
	nextID   int
}

type notification struct {
	Table  string `json:"table"`
	Action string `json:"action"`
	Data   string `json:"data"`
}

func (a *Adapter) newEvent(n *pq.Notification) cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetType("dev.vaikas.postgres")
	// Make this into full db + table.
	event.SetSource(n.Channel)

	if err := event.SetData(cloudevents.ApplicationJSON, n.Extra); err != nil {
		a.logger.Errorw("failed to set data")
	}
	a.nextID++
	return event
}

// Start runs the adapter.
// Returns if stopCh is closed or Send() returns an error.
func (a *Adapter) Start(ctx context.Context) error {
	a.logger.Infow("Starting adapter")
	for {
		select {
		case n := <-a.listener.Notify:
			fmt.Println("Received data from channel [", n.Channel, "] :")
			event := a.newEvent(n)
			event.SetType("dev.vaikas.postgres")
			// Make this into db/channel and maybe others?
			event.SetSource(n.Channel)

			if result := a.client.Send(ctx, event); !cloudevents.IsACK(result) {
				a.logger.Infow("failed to send event", zap.String("event", event.String()), zap.Error(result))
				// We got an error but it could be transient, try again next interval.
				continue
			}

			// Prepare notification payload for pretty print
			var prettyJSON bytes.Buffer
			err := json.Indent(&prettyJSON, []byte(n.Extra), "", "\t")
			if err != nil {
				fmt.Println("Error processing JSON: ", err)
				return err
			}
		case <-time.After(30 * time.Second):
			fmt.Println("Received no events for 30 seconds, checking connection")
			go func() {
				a.listener.Ping()
			}()
		case <-ctx.Done():
			a.logger.Info("Shutting down...")
			return nil
		}
	}
}

func NewAdapter(ctx context.Context, aEnv adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
	env := aEnv.(*envConfig)

	connStr, err := bindingsql.ReadKey("connectionstr")
	if err != nil {
		log.Fatal(err)
	}

	errHandler := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			zap.Error(err)
		}
	}

	logging.FromContext(ctx).Infof("Starting to listen for notifications on %q", env.NotificationChannel)
	listener := pq.NewListener(connStr, 10*time.Second, time.Minute, errHandler)
	err = listener.Listen(env.NotificationChannel)
	if err != nil {
		log.Fatal(err)
	}

	return &Adapter{
		client:   ceClient,
		logger:   logging.FromContext(ctx),
		listener: listener,
	}
}
