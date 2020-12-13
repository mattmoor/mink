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
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/mattmoor/mink/pkg/builds/buildpacks"
	"github.com/mattmoor/mink/pkg/builds/dockerfile"
)

var where = flag.String("where", "./examples", "The directory into which we should write examples.")

const generatedHeader = "# DO NOT EDIT THIS IS A GENERATED FILE (see ./hack/update-codegen.sh)\n\n"

func main() {
	flag.Parse()

	outputs := map[string]string{
		"kaniko.yaml":    dockerfile.KanikoTaskString,
		"buildpack.yaml": buildpacks.BuildpackTaskString,
	}

	for k, v := range outputs {
		if err := ioutil.WriteFile(filepath.Join(*where, k), []byte(generatedHeader+v), 0600); err != nil {
			log.Fatalf("Error writing %q: %v", k, err)
		}
	}
}
