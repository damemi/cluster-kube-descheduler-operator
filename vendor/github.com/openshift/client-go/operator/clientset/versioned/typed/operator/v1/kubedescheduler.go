// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	v1 "github.com/openshift/api/operator/v1"
	scheme "github.com/openshift/client-go/operator/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// KubeDeschedulersGetter has a method to return a KubeDeschedulerInterface.
// A group's client should implement this interface.
type KubeDeschedulersGetter interface {
	KubeDeschedulers() KubeDeschedulerInterface
}

// KubeDeschedulerInterface has methods to work with KubeDescheduler resources.
type KubeDeschedulerInterface interface {
	Create(ctx context.Context, kubeDescheduler *v1.KubeDescheduler, opts metav1.CreateOptions) (*v1.KubeDescheduler, error)
	Update(ctx context.Context, kubeDescheduler *v1.KubeDescheduler, opts metav1.UpdateOptions) (*v1.KubeDescheduler, error)
	UpdateStatus(ctx context.Context, kubeDescheduler *v1.KubeDescheduler, opts metav1.UpdateOptions) (*v1.KubeDescheduler, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.KubeDescheduler, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.KubeDeschedulerList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.KubeDescheduler, err error)
	KubeDeschedulerExpansion
}

// kubeDeschedulers implements KubeDeschedulerInterface
type kubeDeschedulers struct {
	client rest.Interface
}

// newKubeDeschedulers returns a KubeDeschedulers
func newKubeDeschedulers(c *OperatorV1Client) *kubeDeschedulers {
	return &kubeDeschedulers{
		client: c.RESTClient(),
	}
}

// Get takes name of the kubeDescheduler, and returns the corresponding kubeDescheduler object, and an error if there is any.
func (c *kubeDeschedulers) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.KubeDescheduler, err error) {
	result = &v1.KubeDescheduler{}
	err = c.client.Get().
		Resource("kubedeschedulers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KubeDeschedulers that match those selectors.
func (c *kubeDeschedulers) List(ctx context.Context, opts metav1.ListOptions) (result *v1.KubeDeschedulerList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.KubeDeschedulerList{}
	err = c.client.Get().
		Resource("kubedeschedulers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kubeDeschedulers.
func (c *kubeDeschedulers) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("kubedeschedulers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a kubeDescheduler and creates it.  Returns the server's representation of the kubeDescheduler, and an error, if there is any.
func (c *kubeDeschedulers) Create(ctx context.Context, kubeDescheduler *v1.KubeDescheduler, opts metav1.CreateOptions) (result *v1.KubeDescheduler, err error) {
	result = &v1.KubeDescheduler{}
	err = c.client.Post().
		Resource("kubedeschedulers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kubeDescheduler).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a kubeDescheduler and updates it. Returns the server's representation of the kubeDescheduler, and an error, if there is any.
func (c *kubeDeschedulers) Update(ctx context.Context, kubeDescheduler *v1.KubeDescheduler, opts metav1.UpdateOptions) (result *v1.KubeDescheduler, err error) {
	result = &v1.KubeDescheduler{}
	err = c.client.Put().
		Resource("kubedeschedulers").
		Name(kubeDescheduler.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kubeDescheduler).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *kubeDeschedulers) UpdateStatus(ctx context.Context, kubeDescheduler *v1.KubeDescheduler, opts metav1.UpdateOptions) (result *v1.KubeDescheduler, err error) {
	result = &v1.KubeDescheduler{}
	err = c.client.Put().
		Resource("kubedeschedulers").
		Name(kubeDescheduler.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kubeDescheduler).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the kubeDescheduler and deletes it. Returns an error if one occurs.
func (c *kubeDeschedulers) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("kubedeschedulers").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kubeDeschedulers) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("kubedeschedulers").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched kubeDescheduler.
func (c *kubeDeschedulers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.KubeDescheduler, err error) {
	result = &v1.KubeDescheduler{}
	err = c.client.Patch(pt).
		Resource("kubedeschedulers").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}