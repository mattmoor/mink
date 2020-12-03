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

package constants

const (
	// ImageDigestResult is the name of the Tekton result that is expected
	// to surface the digest of a singular image produces by a given task
	// or pipeline.  This digest is NOT fully qualified, so it should have
	// the form: sha256:deadbeef   (shortened for brevity)
	//
	// Generally this input string is paired with a parameter that directs
	// the task to publish the image to a particular tag, e.g.
	//   ghcr.io/mattmoor/mink-images:latest
	//
	// So the fully qualified digest may be assembled by concatenating these
	// with an @:
	//   ghcr.io/mattmoor/mink-images:latest@sha256:deadbeef
	ImageDigestResult = "mink-image-digest"
)
