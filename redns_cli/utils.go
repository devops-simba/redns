package main

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/devops-simba/redns/definitions"
)

const domainChars = "a-zA-Z0-9"

var (
	baseDnsAddress = definitions.DNS_Address{
		TTL:     30,
		Enabled: true,
		Healthy: true,
		Weight:  1,
	}
	domainNamePattern = regexp.MustCompile(
		fmt.Sprintf("^(?:[%s](?:[%s-]{0,61}[%s])?\\.)+[%s][%s-]{0,61}[%s]$",
			domainChars, domainChars, domainChars, domainChars, domainChars, domainChars))
)

// Merge a name and domain to create a key in the redis
func GetKey(domain string, name string) string {
	if name == "*" {
		return "&." + domain
	} else if name == "@" {
		return domain
	} else {
		return name + "." + domain
	}
}

// Check if an string contains wildcard syntax characters
func IsWildcard(s string) bool { return strings.IndexAny(s, "*?") != -1 }

// WildcardToRegexp Compile wildcard to a regexp pattern
func WildcardToRegexp(s string) string {
	pattern := ""
	for {
		i := strings.IndexAny(s, "?*")
		if i == -1 {
			pattern += regexp.QuoteMeta(s)
			break
		} else {
			pattern += regexp.QuoteMeta(s[:i])
			if s[i] == '?' {
				pattern += "."
			} else {
				pattern += ".*"
			}
			s = s[i+1:]
		}
	}
	return pattern
}

// CreateRegexp create a Regexp from a series of parts
func CreateRegexp(parts []string) (*regexp.Regexp, error) {
	if len(parts) == 0 {
		return nil, InvalidArgs{}
	} else if len(parts) == 1 {
		return regexp.Compile(parts[0])
	} else {
		return regexp.Compile("(" + strings.Join(parts, ")|(") + ")")
	}
}

// FindString return index of a value in a collection or -1 if value is not in the collection
func FindString(collection []string, value string) int {
	for i, item := range collection {
		if item == value {
			return i
		}
	}

	return -1
}

// FindStringIf return index of a value in a collection or -1 if value is not in the collection
func FindStringIf(collection []string, pred func(string) bool) int {
	for i, item := range collection {
		if pred(item) {
			return i
		}
	}

	return -1
}

// Contains return `true` if value is in the collection and `false` otherwise
func Contains(collection []string, value string) bool { return FindString(collection, value) != -1 }

// GetRedisKey return the key that we should use in REDIS to hold a record for a domain and a name
func GetRedisKey(domain string, name string) string {
	if name == "@" {
		return domain
	} else if name == "*" {
		return "$." + name
	} else {
		return name + "." + domain
	}
}

// IsDomainName check if a value is a domain name
func IsDomainName(value string) bool {
	return domainNamePattern.MatchString(value)
}

// IsIP check if a value is an IP or not
func IsIP(value string) bool {
	return net.ParseIP(value) != nil
}

// IsIPv4 check if a value is an IP v4
func IsIPv4(value string) bool {
	ip := net.ParseIP(value)
	if ip == nil {
		return false
	}
	return ip.To4() != nil
}

// IsIPv6 check if a value is an IP v4
func IsIPv6(value string) bool {
	ip := net.ParseIP(value)
	if ip == nil {
		return false
	}
	return ip.To16() != nil
}

// ParseSRV parse a string as a SRV record value
func ParseSRV(value string) (string, Word, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 || !IsDomainName(parts[0]) {
		return "", 0, errors.New("Invalid SRV value(must be <domain_name>:<port>")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil || port <= 0 || port > 0xFFFF {
		return "", 0, errors.New("Invalid port number")
	}

	return parts[0], Word(port), nil
}

// AddIPAddress add a new IP address to an IP record, return added IP address and a flag that
// indicate whether IP added as a new address(`true`) or it was an existing one(`false`)
func AddIPAddress(prec **definitions.DNS_IP_Record, ip string) (*definitions.DNS_IP_Address, bool) {
	if *prec == nil {
		*prec = &definitions.DNS_IP_Record{}
	}
	rec := *prec
	for i := 0; i < len(rec.Addresses); i++ {
		if rec.Addresses[i].IP == ip {
			return &rec.Addresses[i], false
		}
	}

	rec.Addresses = append(rec.Addresses, definitions.DNS_IP_Address{
		DNS_Address: baseDnsAddress,
		IP:          ip,
	})

	return &rec.Addresses[len(rec.Addresses)-1], true
}

// AddSTRAddress add a new STRING to a DNS string record. return added address and a flag that
// indicate whether this is a new address(`true`) or it was an existing one(`false`)
func AddSTRAddress(prec **definitions.DNS_STRING_Record, value string, forceAdd bool) (*definitions.DNS_STRING_Address, bool) {
	if *prec == nil {
		*prec = &definitions.DNS_STRING_Record{}
	}
	rec := *prec
	if !forceAdd {
		for i := 0; i < len(rec.Addresses); i++ {
			if rec.Addresses[i].Value == value {
				return &rec.Addresses[i], false
			}
		}
	}

	rec.Addresses = append(rec.Addresses, definitions.DNS_STRING_Address{
		DNS_Address: baseDnsAddress,
		Value:       value,
	})

	return &rec.Addresses[len(rec.Addresses)-1], true
}

// AddSTRAddress add a new MX address to a DNS MX record. return added address and a flag that
// indicate whether this is a new address(`true`) or it was an existing one(`false`)
func AddMXAddress(prec **definitions.DNS_MX_Record, value string) (*definitions.DNS_MX_Address, bool) {
	if *prec == nil {
		*prec = &definitions.DNS_MX_Record{}
	}
	rec := *prec

	for i := 0; i < len(rec.Addresses); i++ {
		if rec.Addresses[i].Value == value {
			return &rec.Addresses[i], false
		}
	}

	rec.Addresses = append(rec.Addresses, definitions.DNS_MX_Address{
		DNS_Address: baseDnsAddress,
		Value:       value,
	})

	return &rec.Addresses[len(rec.Addresses)-1], true
}

// AddSRVAddress add a new SRV address to a DNS SRV record. return added address and a flag that
// indicate whether this is a new address(`true`) or it was an existing one(`false`)
func AddSRVAddress(prec **definitions.DNS_SRV_Record, server string, port uint16) (*definitions.DNS_SRV_Address, bool) {
	if *prec == nil {
		*prec = &definitions.DNS_SRV_Record{}
	}
	rec := *prec

	for i := 0; i < len(rec.Addresses); i++ {
		if rec.Addresses[i].Value == server && rec.Addresses[i].Port == port {
			return &rec.Addresses[i], false
		}
	}

	rec.Addresses = append(rec.Addresses, definitions.DNS_SRV_Address{
		DNS_Address: baseDnsAddress,
		Value:       server,
		Port:        port,
	})

	return &rec.Addresses[len(rec.Addresses)-1], true
}
