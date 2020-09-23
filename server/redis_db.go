package main

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/elcuervo/redisurl"
	"github.com/hoisie/redis"

	"github.com/devops-simba/redns/definitions"
)

const serialNumberKey = "dns-server-serial-no"

type RedisDNSDatabase struct {
	redis.Client
}

func NewRedisDNSDatabase(url string) (*RedisDNSDatabase, error) {
	redisUrl := redisurl.Parse(url)

	db := &RedisDNSDatabase{}
	db.Addr = net.JoinHostPort(redisUrl.Host, strconv.Itoa(redisUrl.Port))
	db.Db = redisUrl.Database
	db.Password = redisUrl.Password
	_, err := db.GetSerialNumber() // open connection
	return db, err
}
func (this *RedisDNSDatabase) lookup(key string) (*definitions.DNSRecord, error) {
	lowerKey := strings.ToLower(key)
	bytearr, err := this.Get(lowerKey)
	if err != nil {
		return nil, err
	}
	if bytearr != nil {
		r := &definitions.DNSRecord{}
		err = json.Unmarshal(bytearr, r)
		if err != nil {
			log.Printf("[ERR] Error in unmarshaling data that received from redis: %s: %v", key, err)
			return nil, err
		}
		return r, nil
	}
	return nil, nil
}
func (this *RedisDNSDatabase) GetSerialNumber() (uint32, error) {
	sn, err := this.Get(serialNumberKey)
	if err != nil {
		return 0, nil
	}
	if len(sn) != 0 {
		return 0, nil
	}
	return binary.LittleEndian.Uint32(sn), nil
}
func (this *RedisDNSDatabase) FindRecord(key string, qType uint16) (*definitions.DNSRecord, error) {
	record, err := this.lookup(key)
	if err != nil {
		return nil, err
	}
	if record != nil {
		return record, nil
	}

	parts := strings.Split(key, ".")
	if len(parts) > 2 {
		parts[0] = "$" // replace '*' with '$' so we does not mess with REDIS escape chars
		record, err = this.lookup(strings.Join(parts, "."))
		if err != nil {
			return nil, err
		}
		if record != nil {
			return record, nil
		}
	}

	return nil, nil
}
