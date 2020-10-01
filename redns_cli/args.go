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

func (this CommandArgs) BindFlags(flagset *flag.FlagSet) {
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
		return nil, err
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

// FindRecords find all the records that match the input domain and name
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
