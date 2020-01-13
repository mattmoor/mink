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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	scheme "knative.dev/eventing/pkg/client/clientset/versioned/scheme"
)

// SinkBindingsGetter has a method to return a SinkBindingInterface.
// A group's client should implement this interface.
type SinkBindingsGetter interface {
	SinkBindings(namespace string) SinkBindingInterface
}

// SinkBindingInterface has methods to work with SinkBinding resources.
type SinkBindingInterface interface {
	Create(*v1alpha1.SinkBinding) (*v1alpha1.SinkBinding, error)
	Update(*v1alpha1.SinkBinding) (*v1alpha1.SinkBinding, error)
	UpdateStatus(*v1alpha1.SinkBinding) (*v1alpha1.SinkBinding, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.SinkBinding, error)
	List(opts v1.ListOptions) (*v1alpha1.SinkBindingList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SinkBinding, err error)
	SinkBindingExpansion
}

// sinkBindings implements SinkBindingInterface
type sinkBindings struct {
	client rest.Interface
	ns     string
}

// newSinkBindings returns a SinkBindings
func newSinkBindings(c *SourcesV1alpha1Client, namespace string) *sinkBindings {
	return &sinkBindings{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the sinkBinding, and returns the corresponding sinkBinding object, and an error if there is any.
func (c *sinkBindings) Get(name string, options v1.GetOptions) (result *v1alpha1.SinkBinding, err error) {
	result = &v1alpha1.SinkBinding{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sinkbindings").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SinkBindings that match those selectors.
func (c *sinkBindings) List(opts v1.ListOptions) (result *v1alpha1.SinkBindingList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.SinkBindingList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sinkbindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested sinkBindings.
func (c *sinkBindings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("sinkbindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a sinkBinding and creates it.  Returns the server's representation of the sinkBinding, and an error, if there is any.
func (c *sinkBindings) Create(sinkBinding *v1alpha1.SinkBinding) (result *v1alpha1.SinkBinding, err error) {
	result = &v1alpha1.SinkBinding{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("sinkbindings").
		Body(sinkBinding).
		Do().
		Into(result)
	return
}

// Update takes the representation of a sinkBinding and updates it. Returns the server's representation of the sinkBinding, and an error, if there is any.
func (c *sinkBindings) Update(sinkBinding *v1alpha1.SinkBinding) (result *v1alpha1.SinkBinding, err error) {
	result = &v1alpha1.SinkBinding{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sinkbindings").
		Name(sinkBinding.Name).
		Body(sinkBinding).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *sinkBindings) UpdateStatus(sinkBinding *v1alpha1.SinkBinding) (result *v1alpha1.SinkBinding, err error) {
	result = &v1alpha1.SinkBinding{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sinkbindings").
		Name(sinkBinding.Name).
		SubResource("status").
		Body(sinkBinding).
		Do().
		Into(result)
	return
}

// Delete takes name of the sinkBinding and deletes it. Returns an error if one occurs.
func (c *sinkBindings) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sinkbindings").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *sinkBindings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sinkbindings").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched sinkBinding.
func (c *sinkBindings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SinkBinding, err error) {
	result = &v1alpha1.SinkBinding{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("sinkbindings").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
