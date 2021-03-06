// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	rednsv1 "github.com/devops-simba/redns/definitions/apis/redns/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeDNSLoadBalancers implements DNSLoadBalancerInterface
type FakeDNSLoadBalancers struct {
	Fake *FakeDevopsV1
}

var dnsloadbalancersResource = schema.GroupVersionResource{Group: "devops.snapp.ir", Version: "v1", Resource: "dnsloadbalancers"}

var dnsloadbalancersKind = schema.GroupVersionKind{Group: "devops.snapp.ir", Version: "v1", Kind: "DNSLoadBalancer"}

// Get takes name of the dNSLoadBalancer, and returns the corresponding dNSLoadBalancer object, and an error if there is any.
func (c *FakeDNSLoadBalancers) Get(ctx context.Context, name string, options v1.GetOptions) (result *rednsv1.DNSLoadBalancer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(dnsloadbalancersResource, name), &rednsv1.DNSLoadBalancer{})
	if obj == nil {
		return nil, err
	}
	return obj.(*rednsv1.DNSLoadBalancer), err
}

// List takes label and field selectors, and returns the list of DNSLoadBalancers that match those selectors.
func (c *FakeDNSLoadBalancers) List(ctx context.Context, opts v1.ListOptions) (result *rednsv1.DNSLoadBalancerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(dnsloadbalancersResource, dnsloadbalancersKind, opts), &rednsv1.DNSLoadBalancerList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &rednsv1.DNSLoadBalancerList{ListMeta: obj.(*rednsv1.DNSLoadBalancerList).ListMeta}
	for _, item := range obj.(*rednsv1.DNSLoadBalancerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested dNSLoadBalancers.
func (c *FakeDNSLoadBalancers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(dnsloadbalancersResource, opts))
}

// Create takes the representation of a dNSLoadBalancer and creates it.  Returns the server's representation of the dNSLoadBalancer, and an error, if there is any.
func (c *FakeDNSLoadBalancers) Create(ctx context.Context, dNSLoadBalancer *rednsv1.DNSLoadBalancer, opts v1.CreateOptions) (result *rednsv1.DNSLoadBalancer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(dnsloadbalancersResource, dNSLoadBalancer), &rednsv1.DNSLoadBalancer{})
	if obj == nil {
		return nil, err
	}
	return obj.(*rednsv1.DNSLoadBalancer), err
}

// Update takes the representation of a dNSLoadBalancer and updates it. Returns the server's representation of the dNSLoadBalancer, and an error, if there is any.
func (c *FakeDNSLoadBalancers) Update(ctx context.Context, dNSLoadBalancer *rednsv1.DNSLoadBalancer, opts v1.UpdateOptions) (result *rednsv1.DNSLoadBalancer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(dnsloadbalancersResource, dNSLoadBalancer), &rednsv1.DNSLoadBalancer{})
	if obj == nil {
		return nil, err
	}
	return obj.(*rednsv1.DNSLoadBalancer), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeDNSLoadBalancers) UpdateStatus(ctx context.Context, dNSLoadBalancer *rednsv1.DNSLoadBalancer, opts v1.UpdateOptions) (*rednsv1.DNSLoadBalancer, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(dnsloadbalancersResource, "status", dNSLoadBalancer), &rednsv1.DNSLoadBalancer{})
	if obj == nil {
		return nil, err
	}
	return obj.(*rednsv1.DNSLoadBalancer), err
}

// Delete takes name of the dNSLoadBalancer and deletes it. Returns an error if one occurs.
func (c *FakeDNSLoadBalancers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(dnsloadbalancersResource, name), &rednsv1.DNSLoadBalancer{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDNSLoadBalancers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(dnsloadbalancersResource, listOpts)

	_, err := c.Fake.Invokes(action, &rednsv1.DNSLoadBalancerList{})
	return err
}

// Patch applies the patch and returns the patched dNSLoadBalancer.
func (c *FakeDNSLoadBalancers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *rednsv1.DNSLoadBalancer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(dnsloadbalancersResource, name, pt, data, subresources...), &rednsv1.DNSLoadBalancer{})
	if obj == nil {
		return nil, err
	}
	return obj.(*rednsv1.DNSLoadBalancer), err
}
