/*
Copyright 2020 The Knative Authors

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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/pflag"
)

type kv struct {
	Name  string `toml:"name"`
	Value string `toml:"value"`
}

type build struct {
	Env []kv `toml:"env"`
}

type project struct {
	Build build `toml:"build"`
}

const platformDir = "/platform/env"

func handleTOML(filename string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		// No project.toml is fine!
		return
	}

	var metadata project
	if _, err := toml.Decode(string(content), &metadata); err != nil {
		log.Fatal("Malformed project.toml: ", err)
	}
	for _, elt := range metadata.Build.Env {
		if err := ioutil.WriteFile(filepath.Join(platformDir, elt.Name), []byte(elt.Value), os.ModePerm); err != nil {
			log.Fatalf("Unable to write %q: %v", elt.Name, err)
		}
		log.Printf("%s=%q", elt.Name, elt.Value)
	}
}

var (
	overrides = pflag.String("overrides", "", "The path to a set of overrides for project.toml")
	env       = pflag.StringSlice("env", nil, "An associative list of KEY=VALUE environment variable overrides.")
)

func handleOverrides() {
	for _, kv := range *env {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			log.Fatal("Unexpected KEY=VALUE:", kv)
		}
		key, value := parts[0], parts[1]
		if err := ioutil.WriteFile(filepath.Join(platformDir, key), []byte(value), os.ModePerm); err != nil {
			log.Fatalf("Unable to write %q: %v", key, err)
		}
		log.Printf("%s=%q", key, value)
	}
}

func main() {
	pflag.Parse()

	if err := os.MkdirAll(platformDir, os.ModePerm); err != nil {
		log.Fatalf("Unable to create %q: %v", platformDir, err)
	}

	handleTOML("./project.toml")

	if *overrides != "" {
		handleTOML(*overrides)
	}
	handleOverrides()
}
