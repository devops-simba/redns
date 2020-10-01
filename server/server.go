package main

import (
	"log"
	"math/rand"
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
			NSRecords: &definitions.DNS_NS_Record{
				Addresses: []definitions.DNS_NS_Address{
					definitions.DNS_NS_Address{
						DNS_STR_Address: definitions.DNS_STR_Address{
							Value: "dns.cloud.snapp.ir",
						},
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

func WeightedSelect(rec definitions.IDNSAddressRecord) definitions.IDNSAddress {
	if rec.IsEmpty() {
		panic("This function should only called on non-empty records")
	}

	addresses := rec.AddressList()
	if len(addresses) == 1 {
		return addresses[0]
	}

	overallWeight := 0
	for i := 0; i < len(addresses); i++ {
		overallWeight += int(addresses[i].BaseAddress().Weight)
	}

	n := rand.Int31n(int32(overallWeight))

	index := 0
	for i := 0; i < len(addresses); i++ {
		weight := int32(addresses[i].BaseAddress().Weight)
		if n < weight {
			index = i
			break
		}

		n -= weight
	}

	return addresses[index]
}
func ToRR(name string, rec definitions.IDNSAddressRecord, rec2 definitions.IDNSAddressRecord) []dns.RR {
	activeRec := rec.LimitToActive()
	if activeRec.IsEmpty() {
		if rec2 != nil {
			activeRec := rec2.LimitToActive()
			if activeRec.IsEmpty() {
				return nil
			}
		} else {
			return nil
		}
	}

	if activeRec.IsWeighted() {
		return []dns.RR{WeightedSelect(activeRec).ToRR(name)}
	}

	return activeRec.ToRRList(name)
}

func A(name string, record *definitions.DNSRecord) []dns.RR {
	return ToRR(name, record.ARecords, record.CNameRecords)
}
func AAAA(name string, record *definitions.DNSRecord) []dns.RR {
	return ToRR(name, record.AAAARecords, record.CNameRecords)
}
func CNAME(name string, record *definitions.DNSRecord) []dns.RR {
	return ToRR(name, record.CNameRecords, nil)
}
func NS(name string, record *definitions.DNSRecord) []dns.RR {
	return ToRR(name, record.NSRecords, nil)
}
func TXT(name string, record *definitions.DNSRecord) []dns.RR {
	return ToRR(name, record.TXTRecords, nil)
}
func MX(name string, record *definitions.DNSRecord) []dns.RR {
	return ToRR(name, record.MXRecords, nil)
}
func SRV(name string, record *definitions.DNSRecord) []dns.RR {
	return ToRR(name, record.SRVRecords, nil)
}

func SOA(name string, record *definitions.DNSRecord, serialNumber uint32) []dns.RR {
	nsRecord := record.NSRecords.LimitToActive()
	if nsRecord.IsEmpty() {
		return nil
	}

	mbox := ""
	mxRecord := record.MXRecords.LimitToActive()
	if !mxRecord.IsEmpty() {
		mbox = mxRecord.AddressList()[0].GetValue()
	}

	return []dns.RR{
		&dns.SOA{
			Hdr:     dns.RR_Header{Name: name, Class: dns.ClassINET, Rrtype: dns.TypeSOA, Ttl: 60},
			Ns:      dns.Fqdn(nsRecord.AddressList()[0].GetValue()),
			Mbox:    mbox,
			Serial:  serialNumber,
			Refresh: 86400,
			Retry:   7200,
			Expire:  3600, // RFC1912 suggests 2-4 weeks 1209600-2419200
			Minttl:  60,
		},
	}
}
