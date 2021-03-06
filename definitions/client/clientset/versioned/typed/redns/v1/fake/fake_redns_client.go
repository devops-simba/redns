// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/devops-simba/redns/definitions/client/clientset/versioned/typed/redns/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeDevopsV1 struct {
	*testing.Fake
}

func (c *FakeDevopsV1) DNSDomains() v1.DNSDomainInterface {
	return &FakeDNSDomains{c}
}

func (c *FakeDevopsV1) DNSLoadBalancers() v1.DNSLoadBalancerInterface {
	return &FakeDNSLoadBalancers{c}
}

func (c *FakeDevopsV1) DNSRecords() v1.DNSRecordInterface {
	return &FakeDNSRecords{c}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeDevopsV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
