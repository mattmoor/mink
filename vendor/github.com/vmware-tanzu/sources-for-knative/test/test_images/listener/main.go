/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"knative.dev/pkg/signals"
)

func main() {
	ctx := signals.NewContext()

	client, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatal(err.Error())
	}
	ctx, cancel := context.WithCancel(ctx)
	once := sync.Once{}

	if err := client.StartReceiver(ctx, func(ctx context.Context, event cloudevents.Event) error {
		log.Printf("Received event: %#v", event)
		once.Do(func() {
			// Launch a go routine to avoid blocking.
			go func() {
				// Ten seconds after the first event, exit normally.
				<-time.After(10 * time.Second)
				cancel()
			}()
		})
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	log.Print("Fin!")
}
