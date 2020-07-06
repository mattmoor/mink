// +build tools

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

package tools

import (
	_ "knative.dev/pkg/hack"
	_ "knative.dev/test-infra/scripts"

	_ "knative.dev/serving/config/core/configmaps"
	_ "knative.dev/serving/test/conformance/ingress"
	_ "knative.dev/serving/test/test_images/flaky"
	_ "knative.dev/serving/test/test_images/grpc-ping"
	_ "knative.dev/serving/test/test_images/helloworld"
	_ "knative.dev/serving/test/test_images/httpproxy"
	_ "knative.dev/serving/test/test_images/runtime"
	_ "knative.dev/serving/test/test_images/timeout"
	_ "knative.dev/serving/test/test_images/wsserver"

	_ "github.com/mikefarah/yq/v3"
	_ "github.com/projectcontour/contour/cmd/contour"
	_ "github.com/projectcontour/contour/examples/contour"
)
