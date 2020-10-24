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

package v1

import (
	"context"
	"time"

	v1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	scheme "knative.dev/net-contour/pkg/client/clientset/versioned/scheme"
)

// HTTPProxiesGetter has a method to return a HTTPProxyInterface.
// A group's client should implement this interface.
type HTTPProxiesGetter interface {
	HTTPProxies(namespace string) HTTPProxyInterface
}

// HTTPProxyInterface has methods to work with HTTPProxy resources.
type HTTPProxyInterface interface {
	Create(ctx context.Context, hTTPProxy *v1.HTTPProxy, opts metav1.CreateOptions) (*v1.HTTPProxy, error)
	Update(ctx context.Context, hTTPProxy *v1.HTTPProxy, opts metav1.UpdateOptions) (*v1.HTTPProxy, error)
	UpdateStatus(ctx context.Context, hTTPProxy *v1.HTTPProxy, opts metav1.UpdateOptions) (*v1.HTTPProxy, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.HTTPProxy, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.HTTPProxyList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.HTTPProxy, err error)
	HTTPProxyExpansion
}

// hTTPProxies implements HTTPProxyInterface
type hTTPProxies struct {
	client rest.Interface
	ns     string
}

// newHTTPProxies returns a HTTPProxies
func newHTTPProxies(c *ProjectcontourV1Client, namespace string) *hTTPProxies {
	return &hTTPProxies{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the hTTPProxy, and returns the corresponding hTTPProxy object, and an error if there is any.
func (c *hTTPProxies) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.HTTPProxy, err error) {
	result = &v1.HTTPProxy{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("httpproxies").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of HTTPProxies that match those selectors.
func (c *hTTPProxies) List(ctx context.Context, opts metav1.ListOptions) (result *v1.HTTPProxyList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.HTTPProxyList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("httpproxies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested hTTPProxies.
func (c *hTTPProxies) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("httpproxies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a hTTPProxy and creates it.  Returns the server's representation of the hTTPProxy, and an error, if there is any.
func (c *hTTPProxies) Create(ctx context.Context, hTTPProxy *v1.HTTPProxy, opts metav1.CreateOptions) (result *v1.HTTPProxy, err error) {
	result = &v1.HTTPProxy{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("httpproxies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(hTTPProxy).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a hTTPProxy and updates it. Returns the server's representation of the hTTPProxy, and an error, if there is any.
func (c *hTTPProxies) Update(ctx context.Context, hTTPProxy *v1.HTTPProxy, opts metav1.UpdateOptions) (result *v1.HTTPProxy, err error) {
	result = &v1.HTTPProxy{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("httpproxies").
		Name(hTTPProxy.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(hTTPProxy).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *hTTPProxies) UpdateStatus(ctx context.Context, hTTPProxy *v1.HTTPProxy, opts metav1.UpdateOptions) (result *v1.HTTPProxy, err error) {
	result = &v1.HTTPProxy{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("httpproxies").
		Name(hTTPProxy.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(hTTPProxy).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the hTTPProxy and deletes it. Returns an error if one occurs.
func (c *hTTPProxies) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("httpproxies").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *hTTPProxies) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("httpproxies").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched hTTPProxy.
func (c *hTTPProxies) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.HTTPProxy, err error) {
	result = &v1.HTTPProxy{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("httpproxies").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}