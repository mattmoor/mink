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

package resources

// These are the label keys that are applied to HTTP proxy resources to facilitate reconciliation.
const (
	// GenerationKey holds the generation of the parent KIngress resource that the HTTPProxy's
	// spec is derived from.  This is updated along with the spec of child HTTPProxy resources
	// and then used to cleanup stale HTTPProxy resources owned by the parent.
	GenerationKey = "contour.networking.knative.dev/generation"
	// ParentKey hold the name of the parent KIngress resource, since OwnerReferences cannot
	// be used in filter expressions.
	ParentKey = "contour.networking.knative.dev/parent"
	// DomainHashKey contains the hash of the fqdn for which this HTTPProxy exists.  We use
	// the hash in place of the actual fqdn because there is a limit on the length of label
	// values.
	DomainHashKey = "contour.networking.knative.dev/domainHash"
)
