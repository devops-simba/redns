package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DNSLoadBalancer is a specification for a DNSLoadBalancer resource
type DNSLoadBalancer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DNSLoadBalancerSpec   `json:"spec"`
	Status DNSLoadBalancerStatus `json:"status"`
}

// DNSLoadBalancerSpec is the spec for a DNSLoadBalancer resource
type DNSLoadBalancerSpec struct {
	Domain  string `json:"domain"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

// DNSLoadBalancerStatus is the status for a DNSLoadBalancer resource
type DNSLoadBalancerStatus struct {
	Generation *int32 `json:"generation"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DNSLoadBalancerList is a list of DNSLoadBalancer resources
type DNSLoadBalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DNSLoadBalancer `json:"items"`
}
