/*
Copyright 2020 The Manba Authors.

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

package v1beta1

import (
	"time"

	v1beta1 "github.com/domgoer/manba-ingress/pkg/apis/configuration/v1beta1"
	scheme "github.com/domgoer/manba-ingress/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ManbaClustersGetter has a method to return a ManbaClusterInterface.
// A group's client should implement this interface.
type ManbaClustersGetter interface {
	ManbaClusters(namespace string) ManbaClusterInterface
}

// ManbaClusterInterface has methods to work with ManbaCluster resources.
type ManbaClusterInterface interface {
	Create(*v1beta1.ManbaCluster) (*v1beta1.ManbaCluster, error)
	Update(*v1beta1.ManbaCluster) (*v1beta1.ManbaCluster, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.ManbaCluster, error)
	List(opts v1.ListOptions) (*v1beta1.ManbaClusterList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ManbaCluster, err error)
	ManbaClusterExpansion
}

// manbaClusters implements ManbaClusterInterface
type manbaClusters struct {
	client rest.Interface
	ns     string
}

// newManbaClusters returns a ManbaClusters
func newManbaClusters(c *ConfigurationV1beta1Client, namespace string) *manbaClusters {
	return &manbaClusters{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the manbaCluster, and returns the corresponding manbaCluster object, and an error if there is any.
func (c *manbaClusters) Get(name string, options v1.GetOptions) (result *v1beta1.ManbaCluster, err error) {
	result = &v1beta1.ManbaCluster{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("manbaclusters").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ManbaClusters that match those selectors.
func (c *manbaClusters) List(opts v1.ListOptions) (result *v1beta1.ManbaClusterList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.ManbaClusterList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("manbaclusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested manbaClusters.
func (c *manbaClusters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("manbaclusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a manbaCluster and creates it.  Returns the server's representation of the manbaCluster, and an error, if there is any.
func (c *manbaClusters) Create(manbaCluster *v1beta1.ManbaCluster) (result *v1beta1.ManbaCluster, err error) {
	result = &v1beta1.ManbaCluster{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("manbaclusters").
		Body(manbaCluster).
		Do().
		Into(result)
	return
}

// Update takes the representation of a manbaCluster and updates it. Returns the server's representation of the manbaCluster, and an error, if there is any.
func (c *manbaClusters) Update(manbaCluster *v1beta1.ManbaCluster) (result *v1beta1.ManbaCluster, err error) {
	result = &v1beta1.ManbaCluster{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("manbaclusters").
		Name(manbaCluster.Name).
		Body(manbaCluster).
		Do().
		Into(result)
	return
}

// Delete takes name of the manbaCluster and deletes it. Returns an error if one occurs.
func (c *manbaClusters) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("manbaclusters").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *manbaClusters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("manbaclusters").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched manbaCluster.
func (c *manbaClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ManbaCluster, err error) {
	result = &v1beta1.ManbaCluster{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("manbaclusters").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
