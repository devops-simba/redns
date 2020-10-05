// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/devops-simba/redns/controller/pkg/apis/redns/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// DNSLoadBalancerLister helps list DNSLoadBalancers.
// All objects returned here must be treated as read-only.
type DNSLoadBalancerLister interface {
	// List lists all DNSLoadBalancers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.DNSLoadBalancer, err error)
	// Get retrieves the DNSLoadBalancer from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.DNSLoadBalancer, error)
	DNSLoadBalancerListerExpansion
}

// dNSLoadBalancerLister implements the DNSLoadBalancerLister interface.
type dNSLoadBalancerLister struct {
	indexer cache.Indexer
}

// NewDNSLoadBalancerLister returns a new DNSLoadBalancerLister.
func NewDNSLoadBalancerLister(indexer cache.Indexer) DNSLoadBalancerLister {
	return &dNSLoadBalancerLister{indexer: indexer}
}

// List lists all DNSLoadBalancers in the indexer.
func (s *dNSLoadBalancerLister) List(selector labels.Selector) (ret []*v1.DNSLoadBalancer, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.DNSLoadBalancer))
	})
	return ret, err
}

// Get retrieves the DNSLoadBalancer from the index for a given name.
func (s *dNSLoadBalancerLister) Get(name string) (*v1.DNSLoadBalancer, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("dnsloadbalancer"), name)
	}
	return obj.(*v1.DNSLoadBalancer), nil
}