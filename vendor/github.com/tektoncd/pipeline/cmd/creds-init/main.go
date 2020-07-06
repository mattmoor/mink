/*
Copyright 2019 The Tekton Authors

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
	"flag"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline"
	"github.com/tektoncd/pipeline/pkg/credentials"
	"github.com/tektoncd/pipeline/pkg/credentials/dockercreds"
	"github.com/tektoncd/pipeline/pkg/credentials/gitcreds"
	"go.uber.org/zap"
)

func main() {
	flag.Parse()

	prod, _ := zap.NewProduction()
	logger := prod.Sugar()

	// ignore atomic level because we are not watching this config for any updates
	defer func() {
		_ = logger.Sync()
	}()

	builders := []credentials.Builder{dockercreds.NewBuilder(), gitcreds.NewBuilder()}
	for _, c := range builders {
		if err := c.Write(pipeline.CredsDir); err != nil {
			logger.Fatalf("Error initializing credentials: %v", err)
		}
	}
	logger.Infof("Credentials initialized.")
}
