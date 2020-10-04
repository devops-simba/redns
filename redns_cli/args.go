package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"regexp"

	"github.com/devops-simba/redns/definitions"
	"github.com/hoisie/redis"
)

type CommandArgs struct {
	// Required
	Redis redis.Client

	Domain DomainName
	Name   SubdomainName
	Kind   Kind
	Value  CommaSeparatedValue

	TTL      DWord
	Weight   Word
	Priority Word
	Enabled  Bool3
	Healthy  Bool3
}

func NewCommandArgs() CommandArgs {
	return CommandArgs{
		TTL:      InvalidDWord,
		Weight:   InvalidWord,
		Priority: InvalidWord,
		Enabled:  None,
		Healthy:  None,
	}
}

func (this *CommandArgs) BindFlags(flagset *flag.FlagSet) {
	if flagset == nil {
		flagset = flag.CommandLine
	}

	flagset.Var(NewRedisValue(&this.Redis), "redis",
		"Redis server that we should connect to it. Its format is [redis://][:password@]host[:port][/DatabaseID]. Default value is 127.0.0.1")
	flagset.Var(&this.Domain, "domain", "Domain or list of domains")
	flagset.Var(&this.Name, "name", "Name(s) of the record(s)")
	flagset.Var(&this.Kind, "kind", "Kind(s) of value(s)")
	flagset.Var(&this.Value, "value", "Value(s) of the record(s)(Address(es))")
	flagset.Var(&this.TTL, "ttl", "TTL of the record")
	flagset.Var(&this.Weight, "weight",
		"Weight of the record, this will be used in load balancing mode")
	flagset.Var(&this.Enabled, "enabled", "Is this address enabled?")
	flagset.Var(&this.Healthy, "healthy", "Is this address healthy?")
	flagset.Var(&this.Priority, "priority",
		"For addresses that support this, it is priority of the address")
}

// ReadRecordByKey Read a record using its key
func (this CommandArgs) ReadRecordByKey(key string) (*definitions.DNSRecord, error) {
	value, err := this.Redis.Get(key)
	if err != nil {
		return nil, nil
	}

	var rec definitions.DNSRecord
	err = json.Unmarshal(value, &rec)
	if err != nil {
		return nil, err
	}

	return &rec, nil
}

// ReadRecord Read a record from this redis
func (this CommandArgs) ReadRecord(domain string, name string) (*DNSRecordWithKey, error) {
	key := GetRedisKey(domain, name)

	rec, err := this.ReadRecordByKey(key)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, nil
	}

	if rec.Domain != domain {
		return nil, fmt.Errorf("Record belong to another domain(Expected: %s, Found: %s)", domain, rec.Domain)
	}

	return &DNSRecordWithKey{Key: key, DNSRecord: *rec}, nil
}

// WriteRecordByKey Write content of the record to Redis server
func (this CommandArgs) WriteRecordByKey(rec *definitions.DNSRecord, key string) error {
	content, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return this.Redis.Set(key, content)
}

//
func (this CommandArgs) WriteRecord(rec *definitions.DNSRecord, domain string, name string) error {
	key := GetRedisKey(domain, name)
	return this.WriteRecordByKey(rec, key)
}

// GetRecordKeyAndSelector get the key that we must pass to redis to select keys that match the parameters
// and 2 regexp to filter selected keys or validate read records
func (this CommandArgs) GetRecordKeyAndSelector() (
	keyPattern string,
	keySelector *regexp.Regexp,
	domainSelector *regexp.Regexp) {
	simpleDomain := len(this.Domain) == 1 && !IsWildcard(this.Domain[0])
	simpleName := len(this.Name) == 1 && !IsWildcard(this.Name[0])
	if simpleDomain && simpleName {
		keySelector = nil
		keyPattern = GetRedisKey(this.Domain[0], this.Name[0])
		domainSelector = regexp.MustCompile(regexp.QuoteMeta(this.Domain[0]))
	} else {
		keyPattern = "*"

		domain_patterns := make([]string, 0, len(this.Domain))
		for _, domain := range this.Domain {
			domain_patterns = append(domain_patterns, WildcardToRegexp(domain))
		}

		name_patterns := make([]string, 0, len(this.Name))
		for _, name := range this.Name {
			name_patterns = append(name_patterns, WildcardToRegexp(name))
		}

		key_selectors := make([]string, 0, len(this.Domain)*len(this.Name))
		for _, domain := range domain_patterns {
			for _, name := range name_patterns {
				var key_pattern string
				if name == "@" {
					key_pattern = domain
				} else {
					key_pattern = name + dot + domain
				}
				key_selectors = append(key_selectors, key_pattern)
			}
		}

		// because we handcrafted this patterns, there is no way that compilation return an error,
		// so we ignore the errors
		keySelector, _ = CreateRegexp(key_selectors)
		domainSelector, _ = CreateRegexp(domain_patterns)
	}

	return keyPattern, keySelector, domainSelector
}

func (this CommandArgs) FindRecords(context DisplayContext) ([]DNSRecordWithKey, error) {
	if len(this.Domain) == 0 || len(this.Name) == 0 {
		return nil, InvalidArgs{}
	}

	var err error
	var keys []string
	pattern, keySelector, domainSelector := this.GetRecordKeyAndSelector()

	// first find all keys that match domain and name
	if IsWildcard(pattern) {
		keys, err = this.Redis.Keys(pattern)
		if err != nil {
			return nil, err
		}
		if keySelector != nil {
			selected_keys := make([]string, 0, len(keys))
			for _, key := range keys {
				if keySelector.MatchString(key) {
					selected_keys = append(selected_keys, key)
				}
			}
			keys = selected_keys
		}
	} else {
		keys = []string{pattern}
	}

	if len(keys) == 0 {
		return nil, nil
	}

	// now read data of those keys
	data, err := this.Redis.Mget(keys...)
	if err != nil {
		return nil, err
	}

	records := make([]DNSRecordWithKey, 0, len(keys))
	for i, item := range data {
		var rec definitions.DNSRecord
		if item == nil {
			context.Warnf("Failed to read content of key '%s'", keys[i])
			continue
		}

		err = json.Unmarshal(item, &rec)
		if err != nil {
			context.Warnf("Failed to deserialize content of key '%s'", keys[i])
			continue
		}
		if domainSelector != nil && !domainSelector.MatchString(rec.Domain) {
			context.Warnf(
				"Found record '%s' that match the descriptor but it belong to another domain: '%s'",
				keys[i], rec.Domain)
			continue
		}

		records = append(records, DNSRecordWithKey{
			DNSRecord: rec,
			Key:       keys[i],
		})
	}

	return records, nil
}

func (this CommandArgs) NewDnsAddress() definitions.DNS_Address {
	return this.UpdatedDnsAddress(&baseDnsAddress)
}
func (this CommandArgs) UpdatedDnsAddress(src *definitions.DNS_Address) definitions.DNS_Address {
	return definitions.DNS_Address{
		TTL:     this.TTL.ValueOr(src.TTL),
		Weight:  this.Weight.ValueOr(src.Weight),
		Enabled: this.Enabled.BoolOr(src.Enabled),
		Healthy: this.Enabled.BoolOr(src.Healthy),
	}
}

func (this CommandArgs) NewIPAddress(value string) definitions.DNS_IP_Address {
	return definitions.DNS_IP_Address{
		DNS_Address: this.NewDnsAddress(),
		IP:          value,
	}
}

func (this CommandArgs) NewSTRAddress(value string) definitions.DNS_STR_Address {
	return definitions.DNS_STR_Address{
		DNS_Address: this.NewDnsAddress(),
		Value:       value,
	}
}

func (this CommandArgs) AddRecord_A(rec *definitions.DNS_A_Record, value string) (
	*definitions.DNS_A_Record, *definitions.DNS_A_Address, bool) {
	var result *definitions.DNS_A_Record
	if rec == nil {
		result = &definitions.DNS_A_Record{
			Addresses: []definitions.DNS_A_Address{
				definitions.DNS_A_Address{
					DNS_IP_Address: this.NewIPAddress(value),
				},
			},
		}
		return result, &result.Addresses[0], true
	} else {
		// search for this IP
		result = &definitions.DNS_A_Record{
			Weighted:  rec.Weighted,
			Addresses: append([]definitions.DNS_A_Address{}, rec.Addresses...),
		}

		for i := 0; i < len(rec.Addresses); i++ {
			if rec.Addresses[i].IP == value {
				// already exists
				result.Addresses[i].DNS_Address = this.UpdatedDnsAddress(&result.Addresses[i].DNS_Address)
				return result, &result.Addresses[i], false
			}
		}

		result.Addresses = append(result.Addresses, definitions.DNS_A_Address{
			DNS_IP_Address: this.NewIPAddress(value),
		})

		return result, &result.Addresses[len(result.Addresses)-1], true
	}
}
func RemoveRecord_A(rec *definitions.DNS_A_Record, index int) *definitions.DNS_A_Record {
	if index < 0 || index >= rec.Length() {
		panic("Invalid index")
	}
	if rec.Length() == 1 {
		if rec.Weighted {
			return &definitions.DNS_A_Record{
				Weighted: true,
			}
		} else {
			return nil
		}
	}

	addresses := make([]definitions.DNS_A_Address, 0, rec.Length()-1)
	addresses = append(addresses, rec.Addresses[:index]...)
	addresses = append(addresses, rec.Addresses[index+1:]...)
	return &definitions.DNS_A_Record{Weighted: rec.Weighted, Addresses: addresses}
}

func (this CommandArgs) AddRecord_AAAA(rec *definitions.DNS_AAAA_Record, value string) (
	*definitions.DNS_AAAA_Record, *definitions.DNS_AAAA_Address, bool) {
	var result *definitions.DNS_AAAA_Record
	if rec == nil {
		result = &definitions.DNS_AAAA_Record{
			Addresses: []definitions.DNS_AAAA_Address{
				definitions.DNS_AAAA_Address{
					DNS_IP_Address: this.NewIPAddress(value),
				},
			},
		}
		return result, &result.Addresses[0], true
	} else {
		// search for this IP
		result = &definitions.DNS_AAAA_Record{
			Weighted:  rec.Weighted,
			Addresses: append([]definitions.DNS_AAAA_Address{}, rec.Addresses...),
		}
		for i := 0; i < len(rec.Addresses); i++ {
			if rec.Addresses[i].IP == value {
				// already exists
				result.Addresses[i].DNS_Address = this.UpdatedDnsAddress(&rec.Addresses[i].DNS_Address)
				return result, &result.Addresses[i], false
			}
		}

		result.Addresses = append(result.Addresses, definitions.DNS_AAAA_Address{
			DNS_IP_Address: this.NewIPAddress(value),
		})
		return result, &result.Addresses[len(result.Addresses)-1], true
	}
}
func RemoveRecord_AAAA(rec *definitions.DNS_AAAA_Record, index int) *definitions.DNS_AAAA_Record {
	if index < 0 || index >= rec.Length() {
		panic("Invalid index")
	}
	if rec.Length() == 1 {
		if rec.Weighted {
			return &definitions.DNS_AAAA_Record{
				Weighted: true,
			}
		} else {
			return nil
		}
	}

	addresses := make([]definitions.DNS_AAAA_Address, 0, rec.Length()-1)
	addresses = append(addresses, rec.Addresses[:index]...)
	addresses = append(addresses, rec.Addresses[index+1:]...)
	return &definitions.DNS_AAAA_Record{Weighted: rec.Weighted, Addresses: addresses}
}

func (this CommandArgs) AddRecord_NS(rec *definitions.DNS_NS_Record, value string) (
	*definitions.DNS_NS_Record, *definitions.DNS_NS_Address, bool) {
	var result *definitions.DNS_NS_Record
	if rec == nil {
		result = &definitions.DNS_NS_Record{
			Addresses: []definitions.DNS_NS_Address{
				definitions.DNS_NS_Address{
					DNS_STR_Address: this.NewSTRAddress(value),
				},
			},
		}
		return result, &result.Addresses[0], true
	} else {
		// search for this IP
		result = &definitions.DNS_NS_Record{
			Weighted:  rec.Weighted,
			Addresses: append([]definitions.DNS_NS_Address{}, rec.Addresses...),
		}
		for i := 0; i < len(rec.Addresses); i++ {
			if rec.Addresses[i].Value == value {
				// already exists
				result.Addresses[i].DNS_Address = this.UpdatedDnsAddress(&result.Addresses[i].DNS_Address)
				return result, &result.Addresses[i], false
			}
		}

		result.Addresses = append(result.Addresses, definitions.DNS_NS_Address{
			DNS_STR_Address: this.NewSTRAddress(value),
		})
		return result, &result.Addresses[len(result.Addresses)-1], true
	}
}
func RemoveRecord_NS(rec *definitions.DNS_NS_Record, index int) *definitions.DNS_NS_Record {
	if index < 0 || index >= rec.Length() {
		panic("Invalid index")
	}
	if rec.Length() == 1 {
		if rec.Weighted {
			return &definitions.DNS_NS_Record{
				Weighted: true,
			}
		} else {
			return nil
		}
	}

	addresses := make([]definitions.DNS_NS_Address, 0, rec.Length()-1)
	addresses = append(addresses, rec.Addresses[:index]...)
	addresses = append(addresses, rec.Addresses[index+1:]...)
	return &definitions.DNS_NS_Record{Weighted: rec.Weighted, Addresses: addresses}
}

func (this CommandArgs) AddRecord_TXT(rec *definitions.DNS_TXT_Record, value string) (
	*definitions.DNS_TXT_Record, *definitions.DNS_TXT_Address, bool) {
	var result *definitions.DNS_TXT_Record
	if rec == nil {
		result = &definitions.DNS_TXT_Record{
			Addresses: []definitions.DNS_TXT_Address{
				definitions.DNS_TXT_Address{
					DNS_STR_Address: this.NewSTRAddress(value),
				},
			},
		}
		return result, &result.Addresses[0], true
	} else {
		// search for this IP
		result = &definitions.DNS_TXT_Record{
			Weighted:  rec.Weighted,
			Addresses: append([]definitions.DNS_TXT_Address{}, rec.Addresses...),
		}
		for i := 0; i < len(rec.Addresses); i++ {
			if rec.Addresses[i].Value == value {
				// already exists
				result.Addresses[i].DNS_Address = this.UpdatedDnsAddress(&result.Addresses[i].DNS_Address)
				return result, &result.Addresses[i], false
			}
		}

		result.Addresses = append(result.Addresses, definitions.DNS_TXT_Address{
			DNS_STR_Address: this.NewSTRAddress(value),
		})
		return result, &result.Addresses[len(result.Addresses)-1], true
	}
}
func RemoveRecord_TXT(rec *definitions.DNS_TXT_Record, index int) *definitions.DNS_TXT_Record {
	if index < 0 || index >= rec.Length() {
		panic("Invalid index")
	}
	if rec.Length() == 1 {
		if rec.Weighted {
			return &definitions.DNS_TXT_Record{
				Weighted: true,
			}
		} else {
			return nil
		}
	}

	addresses := make([]definitions.DNS_TXT_Address, 0, rec.Length()-1)
	addresses = append(addresses, rec.Addresses[:index]...)
	addresses = append(addresses, rec.Addresses[index+1:]...)
	return &definitions.DNS_TXT_Record{Weighted: rec.Weighted, Addresses: addresses}
}

func (this CommandArgs) AddRecord_CNAME(rec *definitions.DNS_CNAME_Record, value string) (
	*definitions.DNS_CNAME_Record, *definitions.DNS_CNAME_Address, bool) {
	var result *definitions.DNS_CNAME_Record
	if rec == nil {
		result = &definitions.DNS_CNAME_Record{
			Addresses: []definitions.DNS_CNAME_Address{
				definitions.DNS_CNAME_Address{
					DNS_STR_Address: this.NewSTRAddress(value),
				},
			},
		}
		return result, &result.Addresses[0], true
	} else {
		// search for this IP
		result = &definitions.DNS_CNAME_Record{
			Weighted:  rec.Weighted,
			Addresses: append([]definitions.DNS_CNAME_Address{}, rec.Addresses...),
		}
		for i := 0; i < len(rec.Addresses); i++ {
			if rec.Addresses[i].Value == value {
				// already exists
				result.Addresses[i].DNS_Address = this.UpdatedDnsAddress(&result.Addresses[i].DNS_Address)
				return result, &result.Addresses[i], false
			}
		}

		result.Addresses = append(result.Addresses, definitions.DNS_CNAME_Address{
			DNS_STR_Address: this.NewSTRAddress(value),
		})
		return result, &result.Addresses[len(result.Addresses)-1], true
	}
}
func RemoveRecord_CNAME(rec *definitions.DNS_CNAME_Record, index int) *definitions.DNS_CNAME_Record {
	if index < 0 || index >= rec.Length() {
		panic("Invalid index")
	}
	if rec.Length() == 1 {
		if rec.Weighted {
			return &definitions.DNS_CNAME_Record{
				Weighted: true,
			}
		} else {
			return nil
		}
	}

	addresses := make([]definitions.DNS_CNAME_Address, 0, rec.Length()-1)
	addresses = append(addresses, rec.Addresses[:index]...)
	addresses = append(addresses, rec.Addresses[index+1:]...)
	return &definitions.DNS_CNAME_Record{Weighted: rec.Weighted, Addresses: addresses}
}

func (this CommandArgs) AddRecord_MX(rec *definitions.DNS_MX_Record, value string) (
	*definitions.DNS_MX_Record, *definitions.DNS_MX_Address, bool) {
	var result *definitions.DNS_MX_Record
	if rec == nil {
		result = &definitions.DNS_MX_Record{
			Addresses: []definitions.DNS_MX_Address{
				definitions.DNS_MX_Address{
					DNS_Address: this.NewDnsAddress(),
					Value:       value,
					Priority:    this.Priority.ValueOr(1),
				},
			},
		}
		return result, &result.Addresses[0], true
	} else {
		// search for this IP
		result = &definitions.DNS_MX_Record{
			Weighted:  rec.Weighted,
			Addresses: append([]definitions.DNS_MX_Address{}, rec.Addresses...),
		}
		for i := 0; i < len(rec.Addresses); i++ {
			if rec.Addresses[i].Value == value {
				// already exists
				result.Addresses[i].DNS_Address = this.UpdatedDnsAddress(&result.Addresses[i].DNS_Address)
				result.Addresses[i].Priority = this.Priority.ValueOr(result.Addresses[i].Priority)
				return result, &result.Addresses[i], false
			}
		}

		result.Addresses = append(result.Addresses, definitions.DNS_MX_Address{
			DNS_Address: this.NewDnsAddress(),
			Value:       value,
			Priority:    this.Priority.ValueOr(1),
		})
		return result, &result.Addresses[len(result.Addresses)-1], true
	}
}
func RemoveRecord_MX(rec *definitions.DNS_MX_Record, index int) *definitions.DNS_MX_Record {
	if index < 0 || index >= rec.Length() {
		panic("Invalid index")
	}
	if rec.Length() == 1 {
		if rec.Weighted {
			return &definitions.DNS_MX_Record{
				Weighted: true,
			}
		} else {
			return nil
		}
	}

	addresses := make([]definitions.DNS_MX_Address, 0, rec.Length()-1)
	addresses = append(addresses, rec.Addresses[:index]...)
	addresses = append(addresses, rec.Addresses[index+1:]...)
	return &definitions.DNS_MX_Record{Weighted: rec.Weighted, Addresses: addresses}
}

func (this CommandArgs) AddRecord_SRV(rec definitions.DNS_SRV_Record, server string, port uint16) (
	definitions.DNS_SRV_Record, *definitions.DNS_SRV_Address, bool) {
	result := make(definitions.DNS_SRV_Record, 0, rec.Length()+1)
	result = append(result, rec...)
	for i := 0; i < rec.Length(); i++ {
		if rec[i].Value == server && rec[i].Port == port {
			// already exists
			result[i].DNS_Address = this.UpdatedDnsAddress(&result[i].DNS_Address) // override possibly changed address
			result[i].Priority = this.Priority.ValueOr(result[i].Priority)
			return result, &result[i], false
		}
	}

	result = append(result, definitions.DNS_SRV_Address{
		DNS_Address: this.NewDnsAddress(),
		Value:       server,
		Port:        port,
	})
	return result, &result[len(result)-1], true
}
func RemoveRecord_SRV(rec definitions.DNS_SRV_Record, index int) definitions.DNS_SRV_Record {
	if index < 0 || index >= rec.Length() {
		panic("Invalid index")
	}

	if rec.Length() == 1 {
		return nil
	}

	result := make(definitions.DNS_SRV_Record, 0, rec.Length()-1)
	result = append(result, rec[:index]...)
	result = append(result, rec[index+1:]...)
	return result
}
