package main

import (
	"encoding/json"
	"regexp"

	"github.com/devops-simba/redns/definitions"
	"github.com/hoisie/redis"
)

var (
	dot = regexp.QuoteMeta(".")
)

type CommandArgs struct {
	// Required
	Redis redis.Client

	Domain DomainName
	Name   SubdomainName
	Kind   Kind
	Value  CommaSeparatedValue

	TTL      Word
	Weight   Word
	Priority Word
	Enabled  Bool3
	Healthy  Bool3
}

func NewCommandArgs() CommandArgs {
	return CommandArgs{
		TTL:      InvalidWord,
		Weight:   InvalidWord,
		Priority: InvalidWord,
		Enabled:  None,
		Healthy:  None,
	}
}

type DisplayContext interface {
	Info(format string, a ...interface{})
	Warn(format string, a ...interface{})
	Error(format string, a ...interface{})

	PrintRecord(rec *definitions.DNSRecord)
	PrintAddressRecord(kind Kind, rec interface{})
	PrintAddress(kind Kind, address interface{})
}

type Command interface {
	// Normalize normalize and validate command arguments
	Normalize(context DisplayContext, args *CommandArgs) error
	// Execute this command
	Execute(context DisplayContext, args CommandArgs) error
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
		keyPattern = this.Name[0] + "." + this.Domain[0]
		keySelector = nil
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
func (this CommandArgs) FindRecords(context DisplayContext) ([]definitions.DNSRecord, error) {
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

	records := make([]definitions.DNSRecord, 0, len(keys))
	for i, item := range data {
		var rec definitions.DNSRecord
		if item == nil {
			context.Warn("Failed to read content of key '%s'", keys[i])
			continue
		}

		err = json.Unmarshal(item, &rec)
		if err != nil {
			context.Warn("Failed to deserialize content of key '%s'", keys[i])
			continue
		}
		if domainSelector != nil && !domainSelector.MatchString(rec.Domain) {
			context.Warn(
				"Found record '%s' that match the descriptor but it belong to another domain: '%s'",
				keys[i], rec.Domain)
			continue
		}

		records = append(records, rec)
	}

	return records, nil
}
