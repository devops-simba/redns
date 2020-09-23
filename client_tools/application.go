package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/elcuervo/redisurl"
	log "github.com/golang/glog"
	"github.com/hoisie/redis"

	"github.com/devops-simba/redns/definitions"
)

type Application struct {
	Redis    *redis.Client
	Action   Command
	Domain   string
	Title    string
	Kind     Kind
	Multi    Bool3
	TTL      int
	Enabled  Bool3
	Healthy  Bool3
	IP       IP
	Value    string
	Port     Word
	Priority Word
	Weight   Word
}

const (
	Default_Enabled        = true
	Default_Healthy        = true
	Default_TTL     uint32 = 30
	Default_Weight  uint16 = 1
)

var defaultAddress = definitions.DNS_Address{
	TTL:     Default_TTL,
	Enabled: Default_Enabled,
	Healthy: Default_Healthy,
	Weight:  Default_Weight,
}

func CreateApplicationFromArgs() *Application {
	app := &Application{
		Port:     InvalidWord,
		Priority: InvalidWord,
		Weight:   InvalidWord,
	}
	var redisUrl string
	flag.StringVar(&redisUrl, "redis", "redis://127.0.0.1:6379",
		"Address of the redis server. Expected format is `redis://[[:password]@]host:port[/db-number][?option=value]`")
	flag.Var(&app.Action, "action",
		fmt.Sprintf("Action that must executed on the server. Acceptable values are {%s}",
			strings.Join(ValidCommands, "|")))
	flag.StringVar(&app.Domain, "domain", "cloud.snapp.ir", "Domain of the record")
	flag.StringVar(&app.Title, "title", "",
		"Title of the record. This will be added to the domain to create the address")
	flag.Var(&app.Kind, "kind",
		fmt.Sprintf("Kind of the record. Acceptable values are {%s}",
			strings.Join(ValidKinds, "|")))
	flag.Var(&app.Multi, "multi", "Should we return multi result or just a single result")
	flag.IntVar(&app.TTL, "ttl", -1, "TTL of the record")
	flag.Var(&app.Enabled, "enabled", "Is this record enabled?")
	flag.Var(&app.Healthy, "healthy", "Is this record healthy?")
	flag.Var(&app.Weight, "weight", "Weight of the record")
	flag.Var(&app.IP, "ip", "IP of the record")
	flag.StringVar(&app.Value, "value", "", "Value for CNAME/NS/TXT/SRV.Target/MX.MBOX")
	flag.Var(&app.Port, "port", "Port of the server")
	flag.Var(&app.Priority, "priority", "Priority of the SRV or MX record")
	flag.Parse()

	if app.Action == EmptyCommand {
		log.Fatal("Action is required")
	}
	if app.Domain == "" && app.Action != GetCommand {
		// all commands except `get` require the domain
		log.Fatal("Domain is required")
	}
	if app.Title == "" && app.Action != RemoveCommand && app.Action != GetCommand {
		log.Fatal("Title is required")
	}
	if app.Kind == Kind_Empty && app.Action != RemoveCommand && app.Action != GetCommand {
		log.Fatal("Kind is required")
	}
	if app.TTL < -1 {
		log.Fatal("Invalid TTL")
	}
	url := redisurl.Parse(redisUrl)
	app.Redis = &redis.Client{
		Addr:     net.JoinHostPort(url.Host, strconv.Itoa(url.Port)),
		Db:       url.Database,
		Password: url.Password,
	}

	// Domain names are case insensitive
	app.Domain = strings.ToLower(app.Domain)
	app.Title = strings.ToLower(app.Title)

	return app
}

func (this *Application) GetKey() string {
	if this.Title == "@" || this.Title == "" {
		return this.Domain
	} else if this.Title == "*" {
		return "$." + this.Domain
	} else {
		return this.Title + "." + this.Domain
	}
}
func (this *Application) ReadRecord(key string, checkDomain bool) (*definitions.DNSRecord, error) {
	if key == "" {
		key = this.GetKey()
	}
	values, err := this.Redis.Mget(key)
	if err != nil {
		return nil, err
	}
	value := values[0]
	if value == nil {
		return nil, nil
	}

	result := &definitions.DNSRecord{}
	err = json.Unmarshal(value, result)
	if err != nil {
		return nil, err
	}
	if checkDomain && result.Domain != this.Domain {
		return nil, fmt.Errorf("Record defined in domain(%s), but you mentioned it with(%s)", result.Domain, this.Domain)
	}

	return result, nil
}
func (this *Application) WriteRecord(record *definitions.DNSRecord, key string) error {
	if key == "" {
		key = this.GetKey()
	}
	value, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return this.Redis.Set(key, value)
}

func (this *Application) haveBaseAddressParams() bool {
	return this.Weight != InvalidWord || this.Enabled != None || this.Healthy != None
}
func (this *Application) haveAddressParams() bool {
	return this.IP != EmptyIP || this.Value != "" || this.Port != InvalidWord || this.Priority != InvalidWord ||
		this.haveBaseAddressParams()
}

func (this *Application) create_Address() definitions.DNS_Address {
	result := definitions.DNS_Address{
		TTL:     Default_TTL,
		Weight:  Default_Weight,
		Enabled: this.Enabled != False,
		Healthy: this.Healthy != False,
	}
	if this.TTL != -1 {
		result.TTL = uint32(this.TTL)
	}
	if this.Weight != InvalidWord {
		result.Weight = uint16(this.Weight)
	}
	return result
}
func (this *Application) update_Address(addr *definitions.DNS_Address) bool {
	result := false
	if this.Enabled != None {
		addr.Enabled = this.Enabled == True
		result = true
	}
	if this.Healthy != None {
		addr.Healthy = this.Healthy == True
		result = true
	}
	if this.TTL != -1 {
		addr.TTL = uint32(this.TTL)
		result = true
	}
	if this.Weight != InvalidWord {
		addr.Weight = uint16(this.Weight)
		result = true
	}
	return result
}

func (this *Application) create_IP_Address() (*definitions.DNS_IP_Address, error) {
	result := &definitions.DNS_IP_Address{
		DNS_Address: this.create_Address(),
		IP:          string(this.IP),
	}

	if result.IP == "" {
		return nil, errors.New("Missing IP")
	}
	return result, nil
}
func (this *Application) update_IP_Address(addr *definitions.DNS_IP_Address) bool {
	result := this.update_Address(&addr.DNS_Address)
	if this.IP != EmptyIP {
		addr.IP = string(this.IP)
		result = true
	}
	return result
}

func (this *Application) create_STR_Address() (*definitions.DNS_STRING_Address, error) {
	result := &definitions.DNS_STRING_Address{
		DNS_Address: this.create_Address(),
		Value:       this.Value,
	}

	if result.Value == "" {
		return nil, errors.New("Missing value")
	}
	return result, nil
}
func (this *Application) update_STR_Address(addr *definitions.DNS_STRING_Address) bool {
	result := this.update_Address(&addr.DNS_Address)
	if this.Value != "" {
		addr.Value = this.Value
		result = true
	}
	return result
}

func (this *Application) create_MX_Address() (*definitions.DNS_MX_Address, error) {
	result := &definitions.DNS_MX_Address{
		DNS_Address: this.create_Address(),
		Server:      this.Value,
		Priority:    1,
	}
	if result.Server == "" {
		return nil, errors.New("Missing value(server address)")
	}
	return result, nil
}
func (this *Application) update_MX_Address(addr *definitions.DNS_MX_Address) bool {
	result := this.update_Address(&addr.DNS_Address)
	if this.Value != "" {
		addr.Server = this.Value
		result = true
	}
	if this.Priority != InvalidWord {
		addr.Priority = uint16(this.Priority)
		result = true
	}
	return result
}

func (this *Application) create_SRV_Address() (*definitions.DNS_SRV_Address, error) {
	result := &definitions.DNS_SRV_Address{
		DNS_Address: this.create_Address(),
		Target:      this.Value,
		Port:        uint16(this.Port),
		Priority:    1,
	}
	if result.Target == "" {
		return nil, errors.New("Missing value(target of the service)")
	}
	if this.Port == InvalidWord {
		return nil, errors.New("Missing port")
	}
	if this.Priority != InvalidWord {
		result.Priority = uint16(this.Priority)
	}
	return result, nil
}
func (this *Application) update_SRV_Address(addr *definitions.DNS_SRV_Address) bool {
	result := this.update_Address(&addr.DNS_Address)
	if this.Value != "" {
		addr.Target = this.Value
		result = true
	}
	if this.Port != InvalidWord {
		addr.Port = uint16(this.Port)
		result = true
	}
	if this.Priority != InvalidWord {
		addr.Priority = uint16(this.Priority)
		result = true
	}
	return result
}

func (this *Application) add_IP(rec *definitions.DNSRecord, addr_rec **definitions.DNS_IP_Record) error {
	if this.IP == EmptyIP {
		return errors.New("Missing ip")
	}

	if *addr_rec == nil || len((*addr_rec).Addresses) == 0 {
		return this.set_IP(rec, addr_rec)
	}

	// search for a record with this IP
	index := -1
	ip := string(this.IP)
	for i := 0; i < len((*addr_rec).Addresses); i++ {
		if (*addr_rec).Addresses[i].IP == ip {
			index = i
			break
		}
	}
	if index == -1 {
		// this is a new record
		addr, err := this.create_IP_Address()
		if err != nil {
			return err
		}

		(*addr_rec).Addresses = append((*addr_rec).Addresses, *addr)
	} else if !this.update_Address(&(*addr_rec).Addresses[index].DNS_Address) {
		return nil // nothing to update
	}

	return this.WriteRecord(rec, "")
}
func (this *Application) add_STR(rec *definitions.DNSRecord, addr_rec **definitions.DNS_STRING_Record) error {
	if this.Value == "" {
		return errors.New("Missing value")
	}

	if *addr_rec == nil || len((*addr_rec).Addresses) == 0 {
		return this.set_STR(rec, addr_rec)
	}

	// search for a record with this value
	index := -1
	for i := 0; i < len((*addr_rec).Addresses); i++ {
		if (*addr_rec).Addresses[i].Value == this.Value {
			index = i
			break
		}
	}
	if index == -1 {
		// this is a new record
		addr, err := this.create_STR_Address()
		if err != nil {
			return err
		}

		(*addr_rec).Addresses = append((*addr_rec).Addresses, *addr)
	} else if !this.update_Address(&(*addr_rec).Addresses[index].DNS_Address) {
		return nil // nothing to update
	}

	return this.WriteRecord(rec, "")
}
func (this *Application) add_MX(rec *definitions.DNSRecord) error {
	if this.Value == "" {
		return errors.New("Missing value(mbox)")
	}

	if rec.MXRecords == nil || len(rec.MXRecords.Addresses) == 0 {
		return this.set_MX(rec)
	}

	// search for a record with this value
	index := -1
	for i := 0; i < len(rec.MXRecords.Addresses); i++ {
		if rec.MXRecords.Addresses[i].Server == this.Value {
			index = i
			break
		}
	}
	if index == -1 {
		// this is a new record
		addr, err := this.create_MX_Address()
		if err != nil {
			return err
		}

		rec.MXRecords.Addresses = append(rec.MXRecords.Addresses, *addr)
	} else {
		updated := this.update_Address(&rec.MXRecords.Addresses[index].DNS_Address)
		if this.Priority != InvalidWord && uint16(this.Priority) != rec.MXRecords.Addresses[index].Priority {
			rec.MXRecords.Addresses[index].Priority = uint16(this.Priority)
			updated = true
		}
		if !updated {
			// nothing to update
			return nil
		}
	}

	return this.WriteRecord(rec, "")
}
func (this *Application) add_SRV(rec *definitions.DNSRecord) error {
	if this.Value == "" {
		return errors.New("Missing value(server)")
	}
	if this.Port == InvalidWord {
		return errors.New("Missing port")
	}

	if rec.SRVRecords == nil || len(rec.SRVRecords.Addresses) == 0 {
		return this.set_MX(rec)
	}

	// search for a record with this value
	index := -1
	port := uint16(this.Port)
	for i := 0; i < len(rec.SRVRecords.Addresses); i++ {
		if rec.SRVRecords.Addresses[i].Target == this.Value && rec.SRVRecords.Addresses[i].Port == port {
			index = i
			break
		}
	}
	if index == -1 {
		// this is a new record
		addr, err := this.create_SRV_Address()
		if err != nil {
			return err
		}

		rec.SRVRecords.Addresses = append(rec.SRVRecords.Addresses, *addr)
	} else {
		updated := this.update_Address(&rec.SRVRecords.Addresses[index].DNS_Address)
		if this.Priority != InvalidWord && uint16(this.Priority) != rec.MXRecords.Addresses[index].Priority {
			rec.MXRecords.Addresses[index].Priority = uint16(this.Priority)
			updated = true
		}
		if !updated {
			// nothing to update
			return nil
		}
	}

	return this.WriteRecord(rec, "")
}
func (this *Application) add() error {
	if this.Multi != None {
		return errors.New("You can't set multi with add command, please use `set` command instead")
	}
	if !this.haveAddressParams() {
		return errors.New("Missing new record options")
	}

	// we must first create a record and then add it to rest of the records
	record, err := this.ReadRecord("", true)
	if err != nil {
		return err
	}
	if record == nil {
		record = &definitions.DNSRecord{Domain: this.Domain}
	}

	switch this.Kind {
	case Kind_A:
		return this.add_IP(record, &record.ARecords)
	case Kind_AAAA:
		return this.add_IP(record, &record.AAAARecords)
	case Kind_CNAME:
		return this.add_STR(record, &record.CNameRecords)
	case Kind_NS:
		return this.add_STR(record, &record.NSRecords)
	case Kind_TXT:
		return this.add_STR(record, &record.TXTRecords)
	case Kind_MX:
		return this.add_MX(record)
	case Kind_SRV:
		return this.add_SRV(record)
	}

	return errors.New("Invalid kind")
}

func (this *Application) set_IP(rec *definitions.DNSRecord, addr_rec **definitions.DNS_IP_Record) error {
	if *addr_rec == nil {
		*addr_rec = &definitions.DNS_IP_Record{Weighted: this.Multi == False}
		if !this.haveAddressParams() {
			return this.WriteRecord(rec, "")
		}
	}

	if this.Multi != None {
		(*addr_rec).Weighted = (this.Multi == False)
	}

	if this.haveAddressParams() {
		addr, err := this.create_IP_Address()
		if err != nil {
			return err
		}

		(*addr_rec).Addresses = []definitions.DNS_IP_Address{*addr}
	}

	return this.WriteRecord(rec, "")
}
func (this *Application) set_STR(rec *definitions.DNSRecord, addr_rec **definitions.DNS_STRING_Record) error {
	if *addr_rec == nil {
		*addr_rec = &definitions.DNS_STRING_Record{Weighted: this.Multi == False}
		if !this.haveAddressParams() {
			return this.WriteRecord(rec, "")
		}
	}

	if this.Multi != None {
		(*addr_rec).Weighted = this.Multi == False
	}

	if this.haveAddressParams() {
		addr, err := this.create_STR_Address()
		if err != nil {
			return err
		}

		(*addr_rec).Addresses = []definitions.DNS_STRING_Address{*addr}
	}

	return this.WriteRecord(rec, "")
}
func (this *Application) set_MX(rec *definitions.DNSRecord) error {
	if rec.MXRecords == nil {
		rec.MXRecords = &definitions.DNS_MX_Record{Weighted: this.Multi == False}
		if !this.haveAddressParams() {
			return this.WriteRecord(rec, "")
		}
	}

	if this.Multi != None {
		rec.MXRecords.Weighted = this.Multi == False
	}

	if this.haveAddressParams() {
		addr, err := this.create_MX_Address()
		if err != nil {
			return err
		}

		rec.MXRecords.Addresses = []definitions.DNS_MX_Address{*addr}
	}

	return this.WriteRecord(rec, "")
}
func (this *Application) set_SRV(rec *definitions.DNSRecord) error {
	if rec.SRVRecords == nil {
		rec.SRVRecords = &definitions.DNS_SRV_Record{}
		if !this.haveAddressParams() {
			return this.WriteRecord(rec, "")
		}
	}

	if this.Multi != None {
		return errors.New("SRV records does not accept multi flag")
	}

	if this.haveAddressParams() {
		addr, err := this.create_SRV_Address()
		if err != nil {
			return err
		}

		rec.SRVRecords.Addresses = []definitions.DNS_SRV_Address{*addr}
	}

	return this.WriteRecord(rec, "")
}
func (this *Application) set() error {
	if this.Multi == None && !this.haveAddressParams() {
		return errors.New("Missing required parameters")
	}

	// we must first create a record and then set it as a value on the records
	record, err := this.ReadRecord("", true)
	if err != nil {
		return err
	}
	if record == nil {
		record = &definitions.DNSRecord{Domain: this.Domain}
	}

	switch this.Kind {
	case Kind_A:
		return this.set_IP(record, &record.ARecords)
	case Kind_AAAA:
		return this.set_IP(record, &record.AAAARecords)
	case Kind_NS:
		return this.set_STR(record, &record.NSRecords)
	case Kind_CNAME:
		return this.set_STR(record, &record.CNameRecords)
	case Kind_TXT:
		return this.set_STR(record, &record.TXTRecords)
	case Kind_MX:
		return this.set_MX(record)
	case Kind_SRV:
		return this.set_SRV(record)
	}

	return errors.New("Unknown kind")
}

func (this *Application) remove_Domain() ([]string, error) {
	if this.haveAddressParams() || this.Multi != None {
		return nil, errors.New("Invalid options")
	}

	keys, err := this.Redis.Keys("*." + this.Domain)
	if err != nil {
		return nil, err
	}

	var deleted bool
	var deleted_keys []string
	domain, err := this.ReadRecord("", true)
	if err != nil {
		return nil, err
	}
	if domain != nil {
		// delete the domain itself
		deleted, err = this.Redis.Del(this.Domain)
		if err != nil {
			return nil, err
		}
		if deleted {
			deleted_keys = append(deleted_keys, this.Domain)
		}
	}

	for _, key := range keys {
		record, err := this.ReadRecord(key, false)
		if record.Domain != this.Domain {
			continue
		}

		deleted, err = this.Redis.Del(key)
		if err != nil {
			return nil, err
		}
		if deleted {
			deleted_keys = append(deleted_keys, key)
		}
	}

	return deleted_keys, nil
}
func (this *Application) remove_Title() ([]string, error) {
	key := this.GetKey()
	deleted, err := this.Redis.Del(key)
	if err != nil {
		return nil, err
	}
	if deleted {
		return []string{key}, nil
	}
	return nil, nil
}
func (this *Application) remove_IP(rec *definitions.DNSRecord, addr_rec **definitions.DNS_IP_Record) ([]string, error) {
	if this.Value != "" || this.Priority != InvalidWord || this.Port != InvalidWord {
		return nil, errors.New("Invalid options")
	}

	if *addr_rec == nil {
		// nothing to remove
		return nil, nil
	}

	addresses := (*addr_rec).Addresses
	base_name := this.GetKey() + "::" + string(this.Kind)
	deleted_records := make([]string, 0, len(addresses)+1)
	if this.IP == EmptyIP {
		// remove all records
		deleted_records = append(deleted_records, base_name)
		for _, addr := range addresses {
			deleted_records = append(deleted_records, base_name+"::"+addr.IP)
		}
		*addr_rec = nil
	} else {
		// find out this specific IP and remove it
		ip := string(this.IP)
		for i, addr := range addresses {
			if addr.IP == ip {
				if len(addresses) == 1 {
					if (*addr_rec).Weighted {
						// if this is a weighted record, only delete the address
						(*addr_rec).Addresses = nil
					} else {
						// if this is not a weighted record, delete entire A record
						*addr_rec = nil
					}
				} else {
					addresses[i] = addresses[len(addresses)-1]
					(*addr_rec).Addresses = addresses[:len(addresses)-1]
				}
				deleted_records = append(deleted_records, base_name+"::"+ip)
				break
			}
		}
	}

	if len(deleted_records) == 0 {
		return nil, nil
	}

	return deleted_records, this.WriteRecord(rec, "")
}
func (this *Application) remove_STR(rec *definitions.DNSRecord, addr_rec **definitions.DNS_STRING_Record) ([]string, error) {
	if this.IP != EmptyIP || this.Priority != InvalidWord || this.Port != InvalidWord {
		return nil, errors.New("Invalid options")
	}

	if *addr_rec == nil {
		// nothing to remove
		return nil, nil
	}

	addresses := (*addr_rec).Addresses
	base_name := this.GetKey() + "::" + string(this.Kind)
	deleted_records := make([]string, 0, len(addresses)+1)
	if this.Value == "" {
		// remove all records
		deleted_records = append(deleted_records, base_name)
		for _, addr := range addresses {
			deleted_records = append(deleted_records, base_name+"::"+addr.Value)
		}
		*addr_rec = nil
	} else {
		// find out this specific Value and remove it
		for i, addr := range addresses {
			if addr.Value == this.Value {
				if len(addresses) == 1 {
					if (*addr_rec).Weighted {
						// if this is a weighted record, only delete the address
						(*addr_rec).Addresses = nil
					} else {
						// if this is not a weighted record, delete entire A record
						*addr_rec = nil
					}
				} else {
					addresses[i] = addresses[len(addresses)-1]
					(*addr_rec).Addresses = addresses[:len(addresses)-1]
				}
				deleted_records = append(deleted_records, base_name+"::"+this.Value)
				break
			}
		}
	}

	if len(deleted_records) == 0 {
		return nil, nil
	}

	return deleted_records, this.WriteRecord(rec, "")
}
func (this *Application) remove_MX(rec *definitions.DNSRecord) ([]string, error) {
	if this.IP != EmptyIP || this.Priority != InvalidWord || this.Port != InvalidWord {
		return nil, errors.New("Invalid options")
	}

	if rec.MXRecords == nil {
		// nothing to remove
		return nil, nil
	}

	base_name := this.GetKey() + "::MX"
	deleted_records := make([]string, 0, len(rec.MXRecords.Addresses)+1)
	if this.Value == "" {
		// remove all records
		deleted_records = append(deleted_records, base_name)
		for _, addr := range rec.MXRecords.Addresses {
			deleted_records = append(deleted_records, base_name+"::"+addr.Server)
		}
		rec.MXRecords = nil
	} else {
		// find out this specific Value and remove it
		for i, addr := range rec.MXRecords.Addresses {
			if addr.Server == this.Value {
				if len(rec.MXRecords.Addresses) == 1 {
					if rec.MXRecords.Weighted {
						// if this is a weighted record, only delete the address
						rec.MXRecords.Addresses = nil
					} else {
						// if this is not a weighted record, delete entire A record
						rec.MXRecords = nil
					}
				} else {
					rec.MXRecords.Addresses[i] = rec.MXRecords.Addresses[len(rec.MXRecords.Addresses)-1]
					rec.MXRecords.Addresses = rec.MXRecords.Addresses[:len(rec.MXRecords.Addresses)-1]
				}
				deleted_records = append(deleted_records, base_name+"::"+this.Value)
				break
			}
		}
	}

	if len(deleted_records) == 0 {
		return nil, nil
	}

	return deleted_records, this.WriteRecord(rec, "")
}
func (this *Application) remove_SRV(rec *definitions.DNSRecord) ([]string, error) {
	if this.IP != EmptyIP || this.Priority != InvalidWord {
		return nil, errors.New("Invalid options")
	}
	if (this.Value == "") != (this.Port == InvalidWord) {
		// either both server and port or none of them
		return nil, errors.New("You must specify both value(server) and port to select an SRV record for removal")
	}

	if rec.SRVRecords == nil {
		// nothing to remove
		return nil, nil
	}

	base_name := this.GetKey() + "::SRV"
	deleted_records := make([]string, 0, len(rec.SRVRecords.Addresses)+1)
	if this.Value == "" {
		// remove all records
		deleted_records = append(deleted_records, base_name)
		for _, addr := range rec.SRVRecords.Addresses {
			deleted_records = append(deleted_records, fmt.Sprintf("%s::%s:%v", base_name, addr.Target, addr.Port))
		}
		rec.SRVRecords = nil
	} else {
		// find out this specific Value and remove it
		port := uint16(this.Port)
		for i, addr := range rec.SRVRecords.Addresses {
			if addr.Target == this.Value && addr.Port == port {
				if len(rec.SRVRecords.Addresses) == 1 {
					rec.SRVRecords = nil
				} else {
					rec.SRVRecords.Addresses[i] = rec.SRVRecords.Addresses[len(rec.SRVRecords.Addresses)-1]
					rec.SRVRecords.Addresses = rec.SRVRecords.Addresses[:len(rec.SRVRecords.Addresses)-1]
				}
				deleted_records = append(deleted_records, fmt.Sprintf("%s::%s:%v", base_name, this.Value, port))
				break
			}
		}
	}

	if len(deleted_records) == 0 {
		return nil, nil
	}

	return deleted_records, this.WriteRecord(rec, "")
}
func (this *Application) remove_Kind() ([]string, error) {
	rec, err := this.ReadRecord("", true)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		// nothing to remove
		return nil, nil
	}

	switch this.Kind {
	case Kind_A:
		return this.remove_IP(rec, &rec.ARecords)
	case Kind_AAAA:
		return this.remove_IP(rec, &rec.AAAARecords)
	case Kind_CNAME:
		return this.remove_STR(rec, &rec.CNameRecords)
	case Kind_NS:
		return this.remove_STR(rec, &rec.NSRecords)
	case Kind_TXT:
		return this.remove_STR(rec, &rec.TXTRecords)
	case Kind_MX:
		return this.remove_MX(rec)
	case Kind_SRV:
		return this.remove_SRV(rec)
	}

	return nil, errors.New("Invalid kind")
}
func (this *Application) remove() ([]string, error) {
	if this.haveBaseAddressParams() || this.Multi != None {
		return nil, errors.New("Invalid options")
	}
	if this.Title == "" {
		return this.remove_Domain()
	}
	if this.Kind == Kind_Empty {
		return this.remove_Title()
	}
	return this.remove_Kind()
}

func (this *Application) get() error {
	if this.Title == "" {
	}
	keys, err := this.Redis.Keys("*")
	return nil
}

func (this *Application) Execute() error {
	switch this.Action {
	case AddCommand:
		return this.add()
	case SetCommand:
		return this.set()
	case RemoveCommand:
		deleted_keys, err := this.remove()
		if err != nil {
			return err
		}
		if len(deleted_keys) == 0 {
			fmt.Printf("NO records found")
		} else {
			fmt.Printf("%d records deleted/updated:\n", len(deleted_keys))
			for _, key := range deleted_keys {
				fmt.Printf("    %s\n", key)
			}
		}
		return nil
	default:
		panic("Unknown command")
	}
}
