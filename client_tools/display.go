package main

import (
	"fmt"
	"strings"

	"github.com/devops-simba/redns/definitions"
)

// GetRecordNameFromKey get name of the record from its key and domain
func GetRecordNameFromKey(key string, record *definitions.DNSRecord) string {
	if key == record.Domain {
		return "@"
	}
	if len(key) <= len(record.Domain) {
		return "<error>"
	}

	if !strings.HasSuffix(key, record.Domain) || key[len(key)-len(record.Domain)-1] != '.' {
		return "<error>"
	}

	key = key[:len(key)-len(record.Domain)-1]
	if key == "$" {
		return "*"
	}
	return key
}

// PrintAddress print an standalone address
func PrintAddress(kind Kind, address interface{}) {
	switch kind {
	case Kind_A:
	case Kind_AAAA:
	case Kind_CNAME:
	case Kind_NS:
	case Kind_TXT:
	case Kind_MX:
	case Kind_SRV:
	}
}

// PrintAddressInRecord print an address as part of printing a record
func PrintAddressInRecord(kind Kind, address interface{}) {

}

var (
	RecordsHeader = fmt.Sprintf("%20s %20s %s", "Domain", "Name", "AvailableAddresses")
)

// PrintRecord print a record
func PrintRecord(rec *definitions.DNSRecord) {

}
