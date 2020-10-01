package definitions

import "github.com/miekg/dns"

const (
	Kind_A     = "A"
	Kind_AAAA  = "AAAA"
	Kind_NS    = "NS"
	Kind_TXT   = "TXT"
	Kind_CNAME = "CNAME"
	Kind_MX    = "MX"
	Kind_SRV   = "SRV"
)

type IDNSAddressRecord interface {
	IsWeighted() bool
	GetItemKind() string
	Length() int
	IsEmpty() bool
	AddressList() []IDNSAddress
	LimitToActive() IDNSAddressRecord
	ToRRList(name string) []dns.RR
}

//region DNS_A_Record
type DNS_A_Record struct {
	Weighted  bool
	Addresses []DNS_A_Address
}

func (this *DNS_A_Record) GetItemKind() string { return Kind_A }
func (this *DNS_A_Record) IsWeighted() bool {
	if this == nil {
		return false
	} else {
		return this.Weighted
	}
}
func (this *DNS_A_Record) Length() int {
	if this == nil {
		return 0
	} else {
		return len(this.Addresses)
	}
}
func (this *DNS_A_Record) IsEmpty() bool {
	return this == nil || len(this.Addresses) == 0
}
func (this *DNS_A_Record) AddressList() []IDNSAddress {
	if this.IsEmpty() {
		return nil
	} else {
		result := make([]IDNSAddress, len(this.Addresses))
		for i := 0; i < len(this.Addresses); i++ {
			result[i] = &this.Addresses[i]
		}
		return result
	}
}
func (this *DNS_A_Record) LimitToActive() IDNSAddressRecord {
	if this == nil {
		return &DNS_A_Record{Weighted: this.Weighted}
	}

	length := len(this.Addresses)
	if length == 0 {
		return &DNS_A_Record{Weighted: this.Weighted}
	}

	addresses := make([]DNS_A_Address, 0, length)
	for i := 0; i < length; i++ {
		if this.Addresses[i].Enabled && this.Addresses[i].Healthy {
			addresses = append(addresses, this.Addresses[i])
		}
	}
	return &DNS_A_Record{Weighted: this.Weighted, Addresses: addresses}
}
func (this *DNS_A_Record) ToRRList(name string) []dns.RR {
	if this == nil {
		return nil
	}

	length := len(this.Addresses)
	if length == 0 {
		return nil
	}

	result := make([]dns.RR, length)
	for i := 0; i < length; i++ {
		result[i] = this.Addresses[i].ToRR(name)
	}
	return result
}

//endregion

//region DNS_AAAA_Record
type DNS_AAAA_Record struct {
	Weighted  bool
	Addresses []DNS_AAAA_Address
}

func (this *DNS_AAAA_Record) GetItemKind() string { return Kind_AAAA }
func (this *DNS_AAAA_Record) IsWeighted() bool {
	if this == nil {
		return false
	} else {
		return this.Weighted
	}
}
func (this *DNS_AAAA_Record) Length() int {
	if this == nil {
		return 0
	} else {
		return len(this.Addresses)
	}
}
func (this *DNS_AAAA_Record) IsEmpty() bool {
	return this == nil || len(this.Addresses) == 0
}
func (this *DNS_AAAA_Record) AddressList() []IDNSAddress {
	if this.IsEmpty() {
		return nil
	} else {
		result := make([]IDNSAddress, len(this.Addresses))
		for i := 0; i < len(this.Addresses); i++ {
			result[i] = &this.Addresses[i]
		}
		return result
	}
}
func (this *DNS_AAAA_Record) LimitToActive() IDNSAddressRecord {
	if this == nil {
		return &DNS_AAAA_Record{Weighted: this.Weighted}
	}

	length := len(this.Addresses)
	if length == 0 {
		return &DNS_AAAA_Record{Weighted: this.Weighted}
	}

	addresses := make([]DNS_AAAA_Address, 0, length)
	for i := 0; i < length; i++ {
		if this.Addresses[i].Enabled && this.Addresses[i].Healthy {
			addresses = append(addresses, this.Addresses[i])
		}
	}
	return &DNS_AAAA_Record{Weighted: this.Weighted, Addresses: addresses}
}
func (this *DNS_AAAA_Record) ToRRList(name string) []dns.RR {
	if this == nil {
		return nil
	}

	length := len(this.Addresses)
	if length == 0 {
		return nil
	}

	result := make([]dns.RR, length)
	for i := 0; i < length; i++ {
		result[i] = this.Addresses[i].ToRR(name)
	}
	return result
}

//endregion

//region DNS_NS_Record
type DNS_NS_Record struct {
	Weighted  bool
	Addresses []DNS_NS_Address
}

func (this *DNS_NS_Record) GetItemKind() string { return Kind_NS }
func (this *DNS_NS_Record) IsWeighted() bool {
	if this == nil {
		return false
	} else {
		return this.Weighted
	}
}
func (this *DNS_NS_Record) Length() int {
	if this == nil {
		return 0
	} else {
		return len(this.Addresses)
	}
}
func (this *DNS_NS_Record) IsEmpty() bool {
	return this == nil || len(this.Addresses) == 0
}
func (this *DNS_NS_Record) AddressList() []IDNSAddress {
	if this.IsEmpty() {
		return nil
	} else {
		result := make([]IDNSAddress, len(this.Addresses))
		for i := 0; i < len(this.Addresses); i++ {
			result[i] = &this.Addresses[i]
		}
		return result
	}
}
func (this *DNS_NS_Record) LimitToActive() IDNSAddressRecord {
	if this == nil {
		return &DNS_NS_Record{Weighted: this.Weighted}
	}

	length := len(this.Addresses)
	if length == 0 {
		return &DNS_NS_Record{Weighted: this.Weighted}
	}

	addresses := make([]DNS_NS_Address, 0, length)
	for i := 0; i < length; i++ {
		if this.Addresses[i].Enabled && this.Addresses[i].Healthy {
			addresses = append(addresses, this.Addresses[i])
		}
	}
	return &DNS_NS_Record{Weighted: this.Weighted, Addresses: addresses}
}
func (this *DNS_NS_Record) ToRRList(name string) []dns.RR {
	if this == nil {
		return nil
	}

	length := len(this.Addresses)
	if length == 0 {
		return nil
	}

	result := make([]dns.RR, length)
	for i := 0; i < length; i++ {
		result[i] = this.Addresses[i].ToRR(name)
	}
	return result
}

//endregion

//region DNS_TXT_Record
type DNS_TXT_Record struct {
	Weighted  bool
	Addresses []DNS_TXT_Address
}

func (this *DNS_TXT_Record) GetItemKind() string { return Kind_TXT }
func (this *DNS_TXT_Record) IsWeighted() bool {
	if this == nil {
		return false
	} else {
		return this.Weighted
	}
}
func (this *DNS_TXT_Record) Length() int {
	if this == nil {
		return 0
	} else {
		return len(this.Addresses)
	}
}
func (this *DNS_TXT_Record) IsEmpty() bool {
	return this == nil || len(this.Addresses) == 0
}
func (this *DNS_TXT_Record) AddressList() []IDNSAddress {
	if this.IsEmpty() {
		return nil
	} else {
		result := make([]IDNSAddress, len(this.Addresses))
		for i := 0; i < len(this.Addresses); i++ {
			result[i] = &this.Addresses[i]
		}
		return result
	}
}
func (this *DNS_TXT_Record) LimitToActive() IDNSAddressRecord {
	if this == nil {
		return &DNS_TXT_Record{Weighted: this.Weighted}
	}

	length := len(this.Addresses)
	if length == 0 {
		return &DNS_TXT_Record{Weighted: this.Weighted}
	}

	addresses := make([]DNS_TXT_Address, 0, length)
	for i := 0; i < length; i++ {
		if this.Addresses[i].Enabled && this.Addresses[i].Healthy {
			addresses = append(addresses, this.Addresses[i])
		}
	}
	return &DNS_TXT_Record{Weighted: this.Weighted, Addresses: addresses}
}
func (this *DNS_TXT_Record) ToRRList(name string) []dns.RR {
	if this == nil {
		return nil
	}

	length := len(this.Addresses)
	if length == 0 {
		return nil
	}

	result := make([]dns.RR, length)
	for i := 0; i < length; i++ {
		result[i] = this.Addresses[i].ToRR(name)
	}
	return result
}

//endregion

//region DNS_CNAME_Record
type DNS_CNAME_Record struct {
	Weighted  bool
	Addresses []DNS_CNAME_Address
}

func (this *DNS_CNAME_Record) GetItemKind() string { return Kind_CNAME }
func (this *DNS_CNAME_Record) IsWeighted() bool {
	if this == nil {
		return false
	} else {
		return this.Weighted
	}
}
func (this *DNS_CNAME_Record) Length() int {
	if this == nil {
		return 0
	} else {
		return len(this.Addresses)
	}
}
func (this *DNS_CNAME_Record) IsEmpty() bool {
	return this == nil || len(this.Addresses) == 0
}
func (this *DNS_CNAME_Record) AddressList() []IDNSAddress {
	if this.IsEmpty() {
		return nil
	} else {
		result := make([]IDNSAddress, len(this.Addresses))
		for i := 0; i < len(this.Addresses); i++ {
			result[i] = &this.Addresses[i]
		}
		return result
	}
}
func (this *DNS_CNAME_Record) LimitToActive() IDNSAddressRecord {
	if this == nil {
		return &DNS_CNAME_Record{Weighted: this.Weighted}
	}

	length := len(this.Addresses)
	if length == 0 {
		return &DNS_CNAME_Record{Weighted: this.Weighted}
	}

	addresses := make([]DNS_CNAME_Address, 0, length)
	for i := 0; i < length; i++ {
		if this.Addresses[i].Enabled && this.Addresses[i].Healthy {
			addresses = append(addresses, this.Addresses[i])
		}
	}
	return &DNS_CNAME_Record{Weighted: this.Weighted, Addresses: addresses}
}
func (this *DNS_CNAME_Record) ToRRList(name string) []dns.RR {
	if this == nil {
		return nil
	}

	length := len(this.Addresses)
	if length == 0 {
		return nil
	}

	result := make([]dns.RR, length)
	for i := 0; i < length; i++ {
		result[i] = this.Addresses[i].ToRR(name)
	}
	return result
}

//endregion

//region DNS_MX_Record
type DNS_MX_Record struct {
	Weighted  bool
	Addresses []DNS_MX_Address
}

func (this *DNS_MX_Record) GetItemKind() string { return Kind_MX }
func (this *DNS_MX_Record) IsWeighted() bool {
	if this == nil {
		return false
	} else {
		return this.Weighted
	}
}
func (this *DNS_MX_Record) Length() int {
	if this == nil {
		return 0
	} else {
		return len(this.Addresses)
	}
}
func (this *DNS_MX_Record) IsEmpty() bool {
	return this == nil || len(this.Addresses) == 0
}
func (this *DNS_MX_Record) AddressList() []IDNSAddress {
	if this.IsEmpty() {
		return nil
	} else {
		result := make([]IDNSAddress, len(this.Addresses))
		for i := 0; i < len(this.Addresses); i++ {
			result[i] = &this.Addresses[i]
		}
		return result
	}
}
func (this *DNS_MX_Record) LimitToActive() IDNSAddressRecord {
	if this == nil {
		return &DNS_MX_Record{Weighted: this.Weighted}
	}

	length := len(this.Addresses)
	if length == 0 {
		return &DNS_MX_Record{Weighted: this.Weighted}
	}

	addresses := make([]DNS_MX_Address, 0, length)
	for i := 0; i < length; i++ {
		if this.Addresses[i].Enabled && this.Addresses[i].Healthy {
			addresses = append(addresses, this.Addresses[i])
		}
	}
	return &DNS_MX_Record{Weighted: this.Weighted, Addresses: addresses}
}
func (this *DNS_MX_Record) ToRRList(name string) []dns.RR {
	if this == nil {
		return nil
	}

	length := len(this.Addresses)
	if length == 0 {
		return nil
	}

	result := make([]dns.RR, length)
	for i := 0; i < length; i++ {
		result[i] = this.Addresses[i].ToRR(name)
	}
	return result
}

//endregion

//region DNS_SRV_Record
type DNS_SRV_Record []DNS_SRV_Address

func (this DNS_SRV_Record) GetItemKind() string { return Kind_SRV }
func (this DNS_SRV_Record) IsWeighted() bool    { return false }
func (this DNS_SRV_Record) Length() int         { return len(this) }
func (this DNS_SRV_Record) IsEmpty() bool       { return len(this) == 0 }
func (this DNS_SRV_Record) AddressList() []IDNSAddress {
	if len(this) == 0 {
		return nil
	} else {
		result := make([]IDNSAddress, len(this))
		for i := 0; i < len(this); i++ {
			result[i] = &this[i]
		}
		return result
	}
}
func (this DNS_SRV_Record) LimitToActive() IDNSAddressRecord {
	length := len(this)
	if length == 0 {
		var result DNS_SRV_Record
		return result
	} else {
		result := make(DNS_SRV_Record, 0, length)
		for i := 0; i < length; i++ {
			if this[i].Enabled && this[i].Healthy {
				result = append(result, this[i])
			}
		}
		return result
	}
}
func (this DNS_SRV_Record) ToRRList(name string) []dns.RR {
	length := len(this)
	if length == 0 {
		return nil
	}

	result := make([]dns.RR, length)
	for i := 0; i < length; i++ {
		result[i] = this[i].ToRR(name)
	}
	return result
}

//endregion

type DNSRecord struct {
	// Domain of this record
	Domain string `json:"domain"`

	// If this is an A record, then this is A record's information
	ARecords *DNS_A_Record `json:"a,omitempty"`
	// If this is an AAAA record, then this is AAAA record's information
	AAAARecords *DNS_AAAA_Record `json:"aaaa,omitempty"`
	// If this is a CNAME record, then this is CNAME record's information
	CNameRecords *DNS_CNAME_Record `json:"cnames,omitempty"`
	// If this is a NS record, then this is NS record's information
	NSRecords *DNS_NS_Record `json:"ns,omitempty"`
	// If this is a MX record, then this is MX record's information
	MXRecords *DNS_MX_Record `json:"mx,omitempty"`
	// If this is a SRV record, then this is SRV record's information
	SRVRecords DNS_SRV_Record `json:"srv,omitempty"`
	// If this is a TXT record, then this is TXT record's information
	TXTRecords *DNS_TXT_Record `json:"txt,omitempty"`
}

// GetAddresses get list of all addresses in a `DNSRecord`
func (this *DNSRecord) GetAddresses() []IDNSAddress {
	if this == nil {
		return nil
	} else {
		var result []IDNSAddress
		result = append(result, this.ARecords.AddressList()...)
		result = append(result, this.AAAARecords.AddressList()...)
		result = append(result, this.NSRecords.AddressList()...)
		result = append(result, this.TXTRecords.AddressList()...)
		result = append(result, this.CNameRecords.AddressList()...)
		result = append(result, this.MXRecords.AddressList()...)
		result = append(result, this.SRVRecords.AddressList()...)
		return result
	}
}
