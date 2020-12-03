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
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/mattmoor/mink/pkg/constants"
)

type image struct {
	Tags   []string `toml:"tags"`
	Digest string   `toml:"digest"`
}

type report struct {
	Image image `toml:"image"`
}

var output = flag.String("output", path.Join("/tekton/results", constants.ImageDigestResult), "Where to write the image digest from report.toml")

func main() {
	flag.Parse()
	content, err := ioutil.ReadFile("./report.toml")
	if err != nil {
		// No project.toml is fine!
		return
	}

	var metadata report
	if _, err := toml.Decode(string(content), &metadata); err != nil {
		log.Fatal("Malformed project.toml: ", err)
	}
	if err := ioutil.WriteFile(*output, []byte(metadata.Image.Digest), os.ModePerm); err != nil {
		log.Fatal("ioutil.WriteFile() =", err)
	}
}
