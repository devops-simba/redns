package definitions

import (
	"net"
	"strconv"

	"github.com/miekg/dns"
)

const (
	PriorityIsNotSupported uint16 = 0xFFFF
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

// IDNSAddress general representation of a DNS Address
type IDNSAddress interface {
	// GetKind Get kind of this address
	GetKind() string
	// GetValue get value of this address, in case of SRV this is server:port
	GetValue() string
	// GetPriority get priority of this address, if priority is not supported, return `PriorityIsNotSupported`
	GetPriority() uint16
	// BaseAddress return this address as a `DNS_Address`
	BaseAddress() *DNS_Address
	// ToRR return a RR(Resource Record) for this address
	ToRR(name string) dns.RR
}

//region DNS_IP_Address
type DNS_IP_Address struct {
	DNS_Address `json:",inline"`
	IP          string `json:"ip"`
}

func (this *DNS_IP_Address) BaseAddress() *DNS_Address { return &this.DNS_Address }
func (this *DNS_IP_Address) GetValue() string          { return this.IP }
func (this *DNS_IP_Address) GetPriority() uint16       { return PriorityIsNotSupported }

type DNS_A_Address struct {
	DNS_IP_Address
}

func (this *DNS_A_Address) GetKind() string { return Kind_A }

func (this *DNS_A_Address) ToRR(name string) dns.RR {
	return &dns.A{
		Hdr: this.createRRHeader(name, dns.TypeA),
		A:   net.ParseIP(this.IP),
	}
}

type DNS_AAAA_Address struct {
	DNS_IP_Address
}

func (this *DNS_AAAA_Address) GetKind() string { return Kind_AAAA }

func (this *DNS_AAAA_Address) ToRR(name string) dns.RR {
	return &dns.AAAA{
		Hdr:  this.createRRHeader(name, dns.TypeAAAA),
		AAAA: net.ParseIP(this.IP),
	}
}

//endregion

//region DNS_STR_Address
type DNS_STR_Address struct {
	DNS_Address `json:",inline"`
	Value       string `json:"value"`
}

func (this *DNS_STR_Address) GetValue() string          { return this.Value }
func (this *DNS_STR_Address) BaseAddress() *DNS_Address { return &this.DNS_Address }
func (this *DNS_STR_Address) GetPriority() uint16       { return PriorityIsNotSupported }

type DNS_NS_Address struct {
	DNS_STR_Address
}

func (this *DNS_NS_Address) GetKind() string { return Kind_NS }
func (this *DNS_NS_Address) ToRR(name string) dns.RR {
	return &dns.NS{
		Hdr: this.createRRHeader(name, dns.TypeNS),
		Ns:  this.Value,
	}
}

type DNS_TXT_Address struct {
	DNS_STR_Address
}

func (this *DNS_TXT_Address) GetKind() string { return Kind_TXT }
func (this *DNS_TXT_Address) ToRR(name string) dns.RR {
	return &dns.TXT{
		Hdr: this.createRRHeader(name, dns.TypeTXT),
		Txt: []string{this.Value},
	}
}

type DNS_CNAME_Address struct {
	DNS_STR_Address
}

func (this *DNS_CNAME_Address) GetKind() string { return Kind_CNAME }
func (this *DNS_CNAME_Address) ToRR(name string) dns.RR {
	return &dns.CNAME{
		Hdr:    this.createRRHeader(name, dns.TypeCNAME),
		Target: this.Value,
	}
}

//endregion

//region DNS_MX_Address
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

func (this *DNS_MX_Address) GetKind() string           { return Kind_MX }
func (this *DNS_MX_Address) GetValue() string          { return this.Value }
func (this *DNS_MX_Address) GetPriority() uint16       { return this.Priority }
func (this *DNS_MX_Address) BaseAddress() *DNS_Address { return &this.DNS_Address }
func (this *DNS_MX_Address) ToRR(name string) dns.RR {
	return &dns.MX{
		Hdr:        this.createRRHeader(name, dns.TypeMX),
		Mx:         this.Value,
		Preference: this.Priority,
	}
}

//endregion

//region DNS_SRV_Address
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

func (this *DNS_SRV_Address) GetKind() string           { return Kind_SRV }
func (this *DNS_SRV_Address) GetPriority() uint16       { return this.Priority }
func (this *DNS_SRV_Address) BaseAddress() *DNS_Address { return &this.DNS_Address }
func (this *DNS_SRV_Address) GetValue() string {
	return this.Value + ":" + strconv.FormatUint(this.Port, 10)
}
func (this *DNS_SRV_Address) ToRR(name string) dns.RR {
	return &dns.SRV{
		Hdr:      this.createRRHeader(name, dns.TypeSRV),
		Target:   this.Value,
		Port:     this.Port,
		Priority: this.Priority,
	}
}

//endregion
