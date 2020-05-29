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

package command

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func Kubeconfig() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		return kubeconfig
	}

	if usr, err := user.Current(); err == nil {
		return filepath.Join(usr.HomeDir, ".kube", "config")
	}

	return ""
}

func Namespace() string {
	ns, _, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: Kubeconfig()},
		&clientcmd.ConfigOverrides{ClusterInfo: api.Cluster{Server: ""}}).Namespace()
	if err == nil {
		return ns
	}

	// TODO(mattmoor): Why isn't this fallback naturally happening?
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return "default"
}
