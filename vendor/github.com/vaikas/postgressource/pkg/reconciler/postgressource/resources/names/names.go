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

package names

import (
	"fmt"
	"strings"

	"knative.dev/pkg/kmeta"

	"github.com/vaikas/postgressource/pkg/apis/sources/v1alpha1"
)

func Deployment(src *v1alpha1.PostgresSource) string {
	return kmeta.ChildName(fmt.Sprintf("postgressource-%s-", src.Name), string(src.GetUID()))
}

func SinkBinding(src *v1alpha1.PostgresSource) string {
	return kmeta.ChildName(fmt.Sprintf("postgressource-%s-sinkbinding", src.Name), string(src.GetUID()))
}

func SQLBinding(src *v1alpha1.PostgresSource) string {
	return kmeta.ChildName(fmt.Sprintf("postgressource-%s-sqlbinding", src.Name), string(src.GetUID()))
}

func RoleBinding(src *v1alpha1.PostgresSource) string {
	return kmeta.ChildName(fmt.Sprintf("postgressource-%s-rolebinding", src.Name), string(src.GetUID()))
}

func ServiceAccount(src *v1alpha1.PostgresSource) string {
	return kmeta.ChildName(fmt.Sprintf("postgressource-%s-sa", src.Name), string(src.GetUID()))
}

// Make postgres compatible name just like we do for k8s (<=63 chars)
// and convert all the - into underscores.
func PostgresName(source *v1alpha1.PostgresSource) string {
	return strings.ReplaceAll(kmeta.ChildName(fmt.Sprintf("postgressource-%s-", source.Name), string(source.GetUID())), "-", "_")

}
