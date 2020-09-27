package definitions

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/miekg/dns"
)

const (
	Kind_A     = "A"
	Kind_AAAA  = "AAAA"
	Kind_NS    = "NS"
	Kind_TXT   = "TXT"
	Kind_CNAME = "CNAME"
	Kind_MX    = "MX"
	Kind_SRV   = "SRV"
)

type DNS_Address struct {
	TTL     uint32 `json:"ttl"`
	Enabled bool   `json:"enabled"`
	Healthy bool   `json:"healthy"`
	Weight  uint16 `json:"weight,omitempty"`
}

func (this DNS_Address) createRRHeader(name string, rrtype uint16) dns.RR_Header {
	return dns.RR_Header{Name: name, Class: dns.ClassINET, Rrtype: rrtype, Ttl: this.TTL}
}

type DNS_IP_Address struct {
	DNS_Address `json:",inline"`
	IP          string `json:"ip"`
}

func NewDNS_IP_Address(ip string, weight uint16, ttl uint32) DNS_IP_Address {
	return DNS_IP_Address{
		DNS_Address: DNS_Address{TTL: ttl, Enabled: true, Healthy: true, Weight: weight},
		IP:          ip,
	}
}

func (this DNS_IP_Address) GetValue() string { return this.IP }
func (this DNS_IP_Address) ToA(name string) dns.RR {
	return &dns.A{
		Hdr: this.createRRHeader(name, dns.TypeA),
		A:   net.ParseIP(this.IP),
	}
}
func (this DNS_IP_Address) ToAAAA(name string) dns.RR {
	return &dns.AAAA{
		Hdr:  this.createRRHeader(name, dns.TypeAAAA),
		AAAA: net.ParseIP(this.IP),
	}
}

type DNS_STRING_Address struct {
	DNS_Address `json:",inline"`
	Value       string `json:"value"`
}

func NewDNS_STRING_Address(value string, weight uint16, ttl uint32) DNS_STRING_Address {
	return DNS_STRING_Address{
		DNS_Address: DNS_Address{TTL: ttl, Enabled: true, Healthy: true, Weight: weight},
		Value:       value,
	}
}

func (this DNS_STRING_Address) GetValue() string { return this.Value }
func (this DNS_STRING_Address) ToCNAME(name string) dns.RR {
	return &dns.CNAME{
		Hdr:    this.createRRHeader(name, dns.TypeCNAME),
		Target: this.Value,
	}
}
func (this DNS_STRING_Address) ToNS(name string) dns.RR {
	return &dns.NS{
		Hdr: this.createRRHeader(name, dns.TypeNS),
		Ns:  this.Value,
	}
}
func (this DNS_STRING_Address) ToTXT(name string) dns.RR {
	return &dns.TXT{
		Hdr: this.createRRHeader(name, dns.TypeTXT),
		Txt: []string{this.Value},
	}
}

type DNS_SRV_Address struct {
	DNS_Address `json:",inline"`
	Value       string `json:"value"`
	Port        uint16 `json:"port"`
	Priority    uint16 `json:"priority"`
}

func NewDNS_SRV_Address(value string, port uint16, priority uint16, weight uint16, ttl uint32) DNS_SRV_Address {
	return DNS_SRV_Address{
		DNS_Address: DNS_Address{TTL: ttl, Enabled: true, Healthy: true, Weight: weight},
		Value:       value,
		Port:        port,
		Priority:    priority,
	}
}

func (this DNS_SRV_Address) GetValue() string {
	return fmt.Sprintf("%s:%d", this.Value, int(this.Port))
}
func (this DNS_SRV_Address) ToSRV(name string) dns.RR {
	return &dns.SRV{
		Hdr:      this.createRRHeader(name, dns.TypeSRV),
		Target:   this.Value,
		Port:     this.Port,
		Priority: this.Priority,
	}
}

type DNS_MX_Address struct {
	DNS_Address `json:",inline"`
	Value       string `json:"value"`
	Priority    uint16 `json:"priority"`
}

func NewDNS_MX_Address(value string, priority uint16, weight uint16, ttl uint32) DNS_MX_Address {
	return DNS_MX_Address{
		DNS_Address: DNS_Address{TTL: ttl, Enabled: true, Healthy: true, Weight: weight},
		Value:       value,
		Priority:    priority,
	}
}

func (this DNS_MX_Address) GetValue() string { return this.Value }
func (this DNS_MX_Address) ToMX(name string) dns.RR {
	return &dns.MX{
		Hdr:        this.createRRHeader(name, dns.TypeMX),
		Mx:         this.Value,
		Preference: this.Priority,
	}
}

type DNS_IP_Record struct {
	Weighted  bool             `json:"weighted"`
	Addresses []DNS_IP_Address `json:"ips,omitempty"`
}

func (this *DNS_IP_Record) Length() int {
	if this == nil {
		return 0
	}
	return len(this.Addresses)
}
func (this *DNS_IP_Record) IsEmpty() bool {
	return this == nil || len(this.Addresses) == 0
}
func (this *DNS_IP_Record) FindValidAddresses() DNS_IP_Record {
	if this.IsEmpty() {
		return DNS_IP_Record{}
	}

	var result []DNS_IP_Address
	for _, addr := range this.Addresses {
		if addr.Enabled && addr.Healthy {
			result = append(result, addr)
		}
	}
	return DNS_IP_Record{Weighted: this.Weighted, Addresses: result}
}
func (this *DNS_IP_Record) ComputeOverallWeight() int {
	result := 0
	for _, addr := range this.Addresses {
		result += int(addr.Weight)
	}
	return result
}
func (this *DNS_IP_Record) WeightedSelectAddress() DNS_IP_Address {
	if this.IsEmpty() {
		panic("This function should only called on non-empty records")
	}
	if len(this.Addresses) == 1 {
		return this.Addresses[0]
	}

	overallWeight := this.ComputeOverallWeight()
	n := rand.Int31n(int32(overallWeight))

	index := 0
	for i, addr := range this.Addresses {
		if n < int32(addr.Weight) {
			index = i
			break
		}

		n -= int32(addr.Weight)
	}

	return this.Addresses[index]
}
func (this *DNS_IP_Record) ToAList(name string) []dns.RR {
	var result []dns.RR
	for _, addr := range this.Addresses {
		result = append(result, addr.ToA(name))
	}
	return result
}
func (this *DNS_IP_Record) ToAAAAList(name string) []dns.RR {
	var result []dns.RR
	for _, addr := range this.Addresses {
		result = append(result, addr.ToAAAA(name))
	}
	return result
}
func (this *DNS_IP_Record) GetAddresses() []DNS_IP_Address {
	if this == nil {
		return []DNS_IP_Address{}
	} else {
		return this.Addresses
	}
}

type DNS_STRING_Record struct {
	Weighted  bool                 `json:"wighted"`
	Addresses []DNS_STRING_Address `json:"names,omitempty"`
}

func (this *DNS_STRING_Record) Length() int {
	if this == nil {
		return 0
	}
	return len(this.Addresses)
}
func (this *DNS_STRING_Record) IsEmpty() bool {
	return this == nil || len(this.Addresses) == 0
}
func (this *DNS_STRING_Record) FindValidAddresses() DNS_STRING_Record {
	if this.IsEmpty() {
		return DNS_STRING_Record{}
	}

	var result []DNS_STRING_Address
	for _, addr := range this.Addresses {
		if addr.Enabled && addr.Healthy {
			result = append(result, addr)
		}
	}
	return DNS_STRING_Record{Weighted: this.Weighted, Addresses: result}
}
func (this *DNS_STRING_Record) ComputeOverallWeight() int {
	result := 0
	for _, addr := range this.Addresses {
		result += int(addr.Weight)
	}
	return result
}
func (this *DNS_STRING_Record) WeightedSelectAddress() DNS_STRING_Address {
	if this.IsEmpty() {
		panic("This function should only called on non-empty records")
	}
	if len(this.Addresses) == 1 {
		return this.Addresses[0]
	}

	overallWeight := this.ComputeOverallWeight()
	n := rand.Int31n(int32(overallWeight))

	index := 0
	for i, addr := range this.Addresses {
		if n < int32(addr.Weight) {
			index = i
			break
		}

		n -= int32(addr.Weight)
	}

	return this.Addresses[index]
}
func (this *DNS_STRING_Record) ToCNAMEList(name string) []dns.RR {
	var result []dns.RR
	for _, addr := range this.Addresses {
		result = append(result, addr.ToCNAME(name))
	}
	return result
}
func (this *DNS_STRING_Record) ToNSList(name string) []dns.RR {
	var result []dns.RR
	for _, addr := range this.Addresses {
		result = append(result, addr.ToNS(name))
	}
	return result
}
func (this *DNS_STRING_Record) ToTXTList(name string) []dns.RR {
	var textRecords []string
	for _, addr := range this.Addresses {
		textRecords = append(textRecords, addr.Value)
	}
	if len(textRecords) == 0 {
		return nil
	}
	return []dns.RR{
		&dns.TXT{
			Hdr: this.Addresses[0].createRRHeader(name, dns.TypeTXT),
			Txt: textRecords,
		},
	}
}
func (this *DNS_STRING_Record) GetAddresses() []DNS_STRING_Address {
	if this == nil {
		return []DNS_STRING_Address{}
	} else {
		return this.Addresses
	}
}

type DNS_MX_Record struct {
	Weighted  bool             `json:"wighted"`
	Addresses []DNS_MX_Address `json:"addresses,omitempty"`
}

func (this *DNS_MX_Record) Length() int {
	if this == nil {
		return 0
	}
	return len(this.Addresses)
}
func (this *DNS_MX_Record) IsEmpty() bool {
	return this == nil || len(this.Addresses) == 0
}
func (this *DNS_MX_Record) FindValidAddresses() DNS_MX_Record {
	if this.IsEmpty() {
		return DNS_MX_Record{}
	}

	var result []DNS_MX_Address
	for _, addr := range this.Addresses {
		if addr.Enabled && addr.Healthy {
			result = append(result, addr)
		}
	}
	return DNS_MX_Record{Weighted: this.Weighted, Addresses: result}
}
func (this *DNS_MX_Record) ComputeOverallWeight() int {
	result := 0
	for _, addr := range this.Addresses {
		result += int(addr.Weight)
	}
	return result
}
func (this *DNS_MX_Record) WeightedSelectAddress() DNS_MX_Address {
	if this.IsEmpty() {
		panic("This function should only called on non-empty records")
	}
	if len(this.Addresses) == 1 {
		return this.Addresses[0]
	}

	overallWeight := this.ComputeOverallWeight()
	n := rand.Int31n(int32(overallWeight))

	index := 0
	for i, addr := range this.Addresses {
		if n < int32(addr.Weight) {
			index = i
			break
		}

		n -= int32(addr.Weight)
	}

	return this.Addresses[index]
}
func (this *DNS_MX_Record) ToMXList(name string) []dns.RR {
	var result []dns.RR
	for _, addr := range this.Addresses {
		result = append(result, addr.ToMX(name))
	}
	return result
}
func (this *DNS_MX_Record) GetAddresses() []DNS_MX_Address {
	if this == nil {
		return []DNS_MX_Address{}
	} else {
		return this.Addresses
	}
}

type DNS_SRV_Record struct {
	Addresses []DNS_SRV_Address `json:"addresses,omitempty"`
}

func (this *DNS_SRV_Record) Length() int {
	if this == nil {
		return 0
	}
	return len(this.Addresses)
}
func (this *DNS_SRV_Record) IsEmpty() bool {
	return this == nil || len(this.Addresses) == 0
}
func (this *DNS_SRV_Record) FindValidAddresses() DNS_SRV_Record {
	if this.IsEmpty() {
		return DNS_SRV_Record{}
	}

	var result []DNS_SRV_Address
	for _, addr := range this.Addresses {
		if addr.Enabled && addr.Healthy {
			result = append(result, addr)
		}
	}
	return DNS_SRV_Record{Addresses: result}
}
func (this *DNS_SRV_Record) ToSRVList(name string) []dns.RR {
	var result []dns.RR
	for _, addr := range this.Addresses {
		result = append(result, addr.ToSRV(name))
	}
	return result
}
func (this *DNS_SRV_Record) GetAddresses() []DNS_SRV_Address {
	if this == nil {
		return []DNS_SRV_Address{}
	} else {
		return this.Addresses
	}
}

type DNSRecord struct {
	// Domain of this record
	Domain string `json:"domain"`

	// If this is an A record, then this is A record's information
	ARecords *DNS_IP_Record `json:"a,omitempty"`
	// If this is an AAAA record, then this is AAAA record's information
	AAAARecords *DNS_IP_Record `json:"aaaa,omitempty"`
	// If this is a CNAME record, then this is CNAME record's information
	CNameRecords *DNS_STRING_Record `json:"cnames,omitempty"`
	// If this is a NS record, then this is NS record's information
	NSRecords *DNS_STRING_Record `json:"ns,omitempty"`
	// If this is a MX record, then this is MX record's information
	MXRecords *DNS_MX_Record `json:"mx,omitempty"`
	// If this is a SRV record, then this is SRV record's information
	SRVRecords *DNS_SRV_Record `json:"srv,omitempty"`
	// If this is a TXT record, then this is TXT record's information
	TXTRecords *DNS_STRING_Record `json:"txt,omitempty"`
}

type DNSAddressPtr struct {
	Kind string
	Ptr  *DNS_Address
}

// GetAddresses get list of all addresses in a `DNSRecord`
func (this *DNSRecord) GetAddresses() []DNSAddressPtr {
	var result []DNSAddressPtr
	for _, addr := range this.ARecords.GetAddresses() {
		result = append(result, DNSAddressPtr{Kind: Kind_A, Ptr: &addr.DNS_Address})
	}
	for _, addr := range this.AAAARecords.GetAddresses() {
		result = append(result, DNSAddressPtr{Kind: Kind_AAAA, Ptr: &addr.DNS_Address})
	}
	for _, addr := range this.NSRecords.GetAddresses() {
		result = append(result, DNSAddressPtr{Kind: Kind_NS, Ptr: &addr.DNS_Address})
	}
	for _, addr := range this.TXTRecords.GetAddresses() {
		result = append(result, DNSAddressPtr{Kind: Kind_TXT, Ptr: &addr.DNS_Address})
	}
	for _, addr := range this.CNameRecords.GetAddresses() {
		result = append(result, DNSAddressPtr{Kind: Kind_CNAME, Ptr: &addr.DNS_Address})
	}
	for _, addr := range this.MXRecords.GetAddresses() {
		result = append(result, DNSAddressPtr{Kind: Kind_MX, Ptr: &addr.DNS_Address})
	}
	for _, addr := range this.SRVRecords.GetAddresses() {
		result = append(result, DNSAddressPtr{Kind: Kind_SRV, Ptr: &addr.DNS_Address})
	}
	return result
}
