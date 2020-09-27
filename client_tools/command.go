package main

import (
	"encoding/json"
	"fmt"

	"github.com/devops-simba/redns/definitions"
	"github.com/hoisie/redis"
)

type ReadRecordResponse struct {
	Key    string
	Record definitions.DNSRecord
	Error  error
}

type CommandArgs struct {
	Redis    *redis.Client
	Domain   string
	Name     string
	Kind     Kind
	Multi    Bool3
	TTL      int
	Enabled  Bool3
	Healthy  Bool3
	Weight   Word
	IP       IP
	Value    string
	Port     Word
	Priority Word
}

func (this CommandArgs) HaveBaseAddressParams() bool {
	return this.TTL != -1 || this.Enabled != None || this.Healthy != None || this.Weight != InvalidWord
}
func (this CommandArgs) HaveAddressParams() bool {
	return this.IP != EmptyIP || this.Value != "" || this.Port != InvalidWord ||
		this.Priority != InvalidWord ||
		this.HaveBaseAddressParams()
}
func (this CommandArgs) GetKey() string {
	if this.Name == "@" || this.Name == "" {
		return this.Domain
	} else if this.Name == "*" {
		return "$." + this.Domain
	} else {
		return this.Name + "." + this.Domain
	}
}
func (this CommandArgs) ReadRecord(key string, checkDomain bool) (definitions.DNSRecord, error) {
	if key == "" {
		key = this.GetKey()
	}
	values, err := this.Redis.Mget(key)

	if err != nil {
		return definitions.DNSRecord{}, err
	}
	value := values[0]
	if value == nil {
		return definitions.DNSRecord{}, nil
	}

	result := definitions.DNSRecord{}
	err = json.Unmarshal(value, result)
	if err != nil {
		return definitions.DNSRecord{}, err
	}
	if checkDomain && result.Domain != this.Domain {
		return definitions.DNSRecord{},
			fmt.Errorf("Record defined in another domain(%s instead of $s)", result.Domain, this.Domain)
	}

	return result, nil
}
func (this CommandArgs) ReadRecords(keys ...string) ([]ReadRecordResponse, error) {
	values, err := this.Redis.Mget(keys...)
	if err != nil {
		return nil, err
	}

	result := make([]ReadRecordResponse, 0, len(values))
	for i, value := range values {
		item := ReadRecordResponse{Key: keys[i]}
		item.Error = json.Unmarshal(value, &item.Record)
		result = append(result, item)
	}

	return result, nil
}
func (this CommandArgs) ReadRecordsByPattern(key string) ([]ReadRecordResponse, error) {
	keys, err := this.Redis.Keys("*")
	if err != nil {
		return nil, err
	}

	return this.ReadRecords(keys...)
}
func (this CommandArgs) WriteRecord(record *definitions.DNSRecord, key string) error {
	if key == "" {
		key = this.GetKey()
	}
	value, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return this.Redis.Set(key, value)
}

type Command interface {
	Validate(args CommandArgs) error
	Execute(args CommandArgs) error
}
