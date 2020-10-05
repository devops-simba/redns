package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DNSRecord is a specification for DNSRecord resource
type DNSRecord struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DNSRecordSpec   `json:"spec"`
	Status DNSRecordStatus `json:"status"`
}

// DNSRecordSpec is the spec for a DNSRecord resource
type DNSRecordSpec struct {
	Domain      string               `json:"domain"`
	Name        string               `json:"name"`
	Type        string               `json:"type"`
	Value       string               `json:"value"`
	Weight      uint16               `json:"weight"`
	TTL         uint16               `json:"ttl"`
	Priority    *uint16              `json:"priority,omitempty"`
	Enabled     bool                 `json:"enabled"`
	HealthCheck DNSRecordHealthCheck `json:"healthCheck"`
}

// DNSRecordStatus is the status for a DNSRecord resource
type DNSRecordStatus struct {
	Generation *int32 `json:"generation"`
}

// DNSRecordHealthCheck is the specification for healthCheck of a DNSRecord resource
type DNSRecordHealthCheck struct {
	Type   string `json:"type"`
	Server string `json:"server"`
}

// DNSRecordList is a list of DNSRecord resources
type DNSRecordList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DNSRecord `json:"items"`
}
