// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	v1 "github.com/devops-simba/redns/controller/pkg/apis/redns/v1"
	scheme "github.com/devops-simba/redns/controller/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// DNSRecordsGetter has a method to return a DNSRecordInterface.
// A group's client should implement this interface.
type DNSRecordsGetter interface {
	DNSRecords() DNSRecordInterface
}

// DNSRecordInterface has methods to work with DNSRecord resources.
type DNSRecordInterface interface {
	Create(ctx context.Context, dNSRecord *v1.DNSRecord, opts metav1.CreateOptions) (*v1.DNSRecord, error)
	Update(ctx context.Context, dNSRecord *v1.DNSRecord, opts metav1.UpdateOptions) (*v1.DNSRecord, error)
	UpdateStatus(ctx context.Context, dNSRecord *v1.DNSRecord, opts metav1.UpdateOptions) (*v1.DNSRecord, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.DNSRecord, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.DNSRecordList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.DNSRecord, err error)
	DNSRecordExpansion
}

// dNSRecords implements DNSRecordInterface
type dNSRecords struct {
	client rest.Interface
}

// newDNSRecords returns a DNSRecords
func newDNSRecords(c *RednsV1Client) *dNSRecords {
	return &dNSRecords{
		client: c.RESTClient(),
	}
}

// Get takes name of the dNSRecord, and returns the corresponding dNSRecord object, and an error if there is any.
func (c *dNSRecords) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.DNSRecord, err error) {
	result = &v1.DNSRecord{}
	err = c.client.Get().
		Resource("dnsrecords").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of DNSRecords that match those selectors.
func (c *dNSRecords) List(ctx context.Context, opts metav1.ListOptions) (result *v1.DNSRecordList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.DNSRecordList{}
	err = c.client.Get().
		Resource("dnsrecords").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested dNSRecords.
func (c *dNSRecords) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("dnsrecords").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a dNSRecord and creates it.  Returns the server's representation of the dNSRecord, and an error, if there is any.
func (c *dNSRecords) Create(ctx context.Context, dNSRecord *v1.DNSRecord, opts metav1.CreateOptions) (result *v1.DNSRecord, err error) {
	result = &v1.DNSRecord{}
	err = c.client.Post().
		Resource("dnsrecords").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(dNSRecord).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a dNSRecord and updates it. Returns the server's representation of the dNSRecord, and an error, if there is any.
func (c *dNSRecords) Update(ctx context.Context, dNSRecord *v1.DNSRecord, opts metav1.UpdateOptions) (result *v1.DNSRecord, err error) {
	result = &v1.DNSRecord{}
	err = c.client.Put().
		Resource("dnsrecords").
		Name(dNSRecord.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(dNSRecord).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *dNSRecords) UpdateStatus(ctx context.Context, dNSRecord *v1.DNSRecord, opts metav1.UpdateOptions) (result *v1.DNSRecord, err error) {
	result = &v1.DNSRecord{}
	err = c.client.Put().
		Resource("dnsrecords").
		Name(dNSRecord.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(dNSRecord).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the dNSRecord and deletes it. Returns an error if one occurs.
func (c *dNSRecords) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("dnsrecords").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *dNSRecords) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("dnsrecords").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched dNSRecord.
func (c *dNSRecords) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.DNSRecord, err error) {
	result = &v1.DNSRecord{}
	err = c.client.Patch(pt).
		Resource("dnsrecords").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}