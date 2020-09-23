package main

import (
	"log"
	"strings"

	"github.com/miekg/dns"

	"github.com/devops-simba/redns/definitions"
)

type DNSRecordType string

type DNSDatabase interface {
	GetSerialNumber() (uint32, error)
	FindRecord(name string, qType uint16) (*definitions.DNSRecord, error)
}

type DNSServer struct {
	server   *dns.Server
	database DNSDatabase
}

func NewDNSServer(database DNSDatabase, port, net string) *DNSServer {
	server := &DNSServer{
		server:   &dns.Server{Addr: "0.0.0.0:" + port, Net: net},
		database: database,
	}
	server.server.Handler = server
	return server
}

func (this *DNSServer) getSerialNumber() uint32 {
	sn, err := this.database.GetSerialNumber()
	if err != nil {
		log.Printf("[ERR] Error in reading serial number from database: %v", err)
		return uint32(0)
	}
	return sn
}
func (this *DNSServer) ServeDNS(w dns.ResponseWriter, msg *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(msg)
	m.Authoritative = true
	m.RecursionAvailable = false
	for _, question := range msg.Question {
		qtype := dns.TypeToString[question.Qtype]
		//log.Printf("[INF] %v %v", qtype, question.Name)

		qName := question.Name
		if strings.HasSuffix(qName, ".") {
			qName = qName[:len(qName)-1]
		}
		record, err := this.database.FindRecord(qName, question.Qtype)
		if err != nil {
			log.Printf("[ERR] Error in finding record %s(%s): %v", qtype, qName, err)
			continue
		}
		if record == nil {
			log.Printf("[WRN] No record found for %s(%s)", qtype, qName)
			continue
		}

		switch question.Qtype {
		case dns.TypeA:
			m.Answer = append(m.Answer, A(question.Name, record)...)
		case dns.TypeAAAA:
			m.Answer = append(m.Answer, AAAA(question.Name, record)...)
		case dns.TypeCNAME:
			m.Answer = append(m.Answer, CNAME(question.Name, record)...)
		case dns.TypeNS:
			m.Answer = append(m.Answer, NS(question.Name, record)...)
		case dns.TypeTXT:
			m.Answer = append(m.Answer, TXT(question.Name, record)...)
		case dns.TypeMX:
			m.Answer = append(m.Answer, MX(question.Name, record)...)
		case dns.TypeSRV:
			m.Answer = append(m.Answer, SRV(question.Name, record)...)
		case dns.TypeSOA:
			m.Answer = append(m.Answer, SOA(question.Name, record, this.getSerialNumber())...)
		default:
			log.Printf("[WRN] Invalid question type: %v %v", qtype, qName)
		}
	}

	if len(m.Answer) == 0 {
		m.Rcode = dns.RcodeNameError
		m.Answer = append(m.Answer, SOA(".", &definitions.DNSRecord{
			NSRecords: &definitions.DNS_STRING_Record{
				Addresses: []definitions.DNS_STRING_Address{
					DNS_STRING_Address{
						Value: "dns.cloud.snapp.ir",
					},
				},
			},
		}, this.getSerialNumber())...)
	}

	err := w.WriteMsg(m)
	if err != nil {
		log.Printf("[ERR] failed to write message: %v", err)
	}
}
func (this *DNSServer) Start() error    { return this.server.ListenAndServe() }
func (this *DNSServer) Shutdown() error { return this.server.Shutdown() }

func A(name string, record *definitions.DNSRecord) []dns.RR {
	validAddresses := record.ARecords.FindValidAddresses()
	if validAddresses.IsEmpty() {
		if !record.CNameRecords.IsEmpty() {
			return CNAME(name, record)
		}
		return nil
	}

	if validAddresses.Weighted && validAddresses.Length() > 1 {
		// find the record that we should return to the caller
		return []dns.RR{
			validAddresses.WeightedSelectAddress().ToA(name),
		}
	}

	return validAddresses.ToAList(name)
}
func AAAA(name string, record *definitions.DNSRecord) []dns.RR {
	validAddresses := record.AAAARecords.FindValidAddresses()
	if validAddresses.IsEmpty() {
		if !record.CNameRecords.IsEmpty() {
			return CNAME(name, record)
		}
		return nil
	}

	if validAddresses.Weighted && validAddresses.Length() > 1 {
		// find the record that we should return to the caller
		return []dns.RR{
			validAddresses.WeightedSelectAddress().ToAAAA(name),
		}
	}

	return validAddresses.ToAAAAList(name)
}
func CNAME(name string, record *definitions.DNSRecord) []dns.RR {
	validAddresses := record.CNameRecords.FindValidAddresses()
	if validAddresses.IsEmpty() {
		return nil
	}

	if validAddresses.Weighted && validAddresses.Length() > 1 {
		// find the record that we should return to the caller
		return []dns.RR{
			validAddresses.WeightedSelectAddress().ToCNAME(name),
		}
	}

	return validAddresses.ToCNAMEList(name)
}
func NS(name string, record *definitions.DNSRecord) []dns.RR {
	validAddresses := record.NSRecords.FindValidAddresses()
	if validAddresses.IsEmpty() {
		return nil
	}

	if validAddresses.Weighted && validAddresses.Length() > 1 {
		// find the record that we should return to the caller
		return []dns.RR{
			validAddresses.WeightedSelectAddress().ToNS(name),
		}
	}

	return validAddresses.ToNSList(name)
}
func TXT(name string, record *definitions.DNSRecord) []dns.RR {
	if record.TXTRecords.IsEmpty() {
		return nil
	}

	return record.TXTRecords.ToTXTList(name)
}
func MX(name string, record *definitions.DNSRecord) []dns.RR {
	validAddresses := record.MXRecords.FindValidAddresses()
	if validAddresses.IsEmpty() {
		return nil
	}
	if validAddresses.Weighted && validAddresses.Length() > 1 {
		// find the record that we should return to the caller
		return []dns.RR{
			validAddresses.WeightedSelectAddress().ToMX(name),
		}
	}

	return validAddresses.ToMXList(name)
}
func SRV(name string, record *definitions.DNSRecord) []dns.RR {
	validAddresses := record.SRVRecords.FindValidAddresses()
	if validAddresses.IsEmpty() {
		return nil
	}
	return validAddresses.ToSRVList(name)
}

func SOA(name string, record *definitions.DNSRecord, serialNumber uint32) []dns.RR {
	validAddresses := record.NSRecords.FindValidAddresses()
	if validAddresses.IsEmpty() {
		return nil
	}

	mbox := ""
	validMailAddresses := record.MXRecords.FindValidAddresses()
	if !validMailAddresses.IsEmpty() {
		mbox = validMailAddresses.Addresses[0].Server
	}

	return []dns.RR{
		&dns.SOA{
			Hdr:     dns.RR_Header{Name: name, Class: dns.ClassINET, Rrtype: dns.TypeSOA, Ttl: 60},
			Ns:      dns.Fqdn(validAddresses.Addresses[0].Value),
			Mbox:    mbox,
			Serial:  serialNumber,
			Refresh: 86400,
			Retry:   7200,
			Expire:  3600, // RFC1912 suggests 2-4 weeks 1209600-2419200
			Minttl:  60,
		},
	}
}
