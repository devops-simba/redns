package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/devops-simba/redns/definitions"
)

type GetCommand struct{}

func (this *GetCommand) Validate(args CommandArgs) error {
	if args.HaveBaseAddressParams() || args.Multi != None || args.Priority != InvalidWord {
		return InvalidOptions{}
	}

	if args.Domain == "" {
		if args.Name != "" || args.Kind != Kind_Empty {
			return errors.New("Options are not allowed without a domain")
		}
		if args.HaveAddressParams() {
			return InvalidOptions{}
		}
	} else if args.Name == "" {
		if args.Kind != Kind_Empty {
			return errors.New("Record name is required")
		}
		if args.HaveAddressParams() {
			return InvalidOptions{}
		}
	} else if args.Kind == Kind_Empty {
		if args.HaveAddressParams() {
			return InvalidOptions{}
		}
	} else {
		switch args.Kind {
		case Kind_A, Kind_AAAA:
			if args.Value != "" || args.Port != InvalidWord {
				return InvalidOptions{}
			}

		case Kind_CNAME, Kind_NS:
			if args.IP != EmptyIP || args.Port != InvalidWord {
				return InvalidOptions{}
			}

		case Kind_TXT:
			if args.HaveAddressParams() {
				return InvalidOptions{}
			}

		case Kind_SRV:
			if args.IP != EmptyIP {
				return InvalidOptions{}
			}

		case Kind_MX:
			if args.IP != EmptyIP || args.Port != InvalidWord {
				return InvalidOptions{}
			}
		}
	}

	return nil
}

func (this *GetCommand) Execute(args CommandArgs) error {
	if args.Domain == "" {
		// return list of all records
		keys, err := args.Redis.Keys("*")
	} else if args.Name == "" {
		// return list of all records from a certain domain
	} else {
		// return only records from certain record
	}

	return nil
}

func join(sep string, items ...string) string {
	insertPosition := 0
	for _, item := range items {
		if item != "" {
			items[insertPosition] = item
			insertPosition++
		}
	}
	items = items[:insertPosition]
	return strings.Join(items, sep)
}
func join_append(sep, src string, items ...string) string {
	for _, item := range items {
		if src != "" {
			src += sep
		}

		src += item
	}
	return src
}

func getIPAddressTitle(name string, rec *definitions.DNS_IP_Record) string {
	if rec == nil {
		return ""
	}
	result := name
	if rec.Weighted {
		result += "(LB)"
	}
	addrCount := len(rec.Addresses)
	if addrCount != 1 {
		result += "=" + strconv.Itoa(addrCount)
	}
	return result
}
func getSTRAddressTitle(name string, rec *definitions.DNS_STRING_Record) string {
	if rec == nil {
		return ""
	}
	result := name
	if rec.Weighted {
		result += "(LB)"
	}
	addrCount := len(rec.Addresses)
	if addrCount != 1 {
		result += "=" + strconv.Itoa(addrCount)
	}
	return result
}
func getMXAddressTitle(rec *definitions.DNS_MX_Record) string {
	if rec == nil {
		return ""
	}

	result := "MX"
	if rec.Weighted {
		result += "(LB)"
	}
	addrCount := len(rec.Addresses)
	if addrCount != 1 {
		result += "=" + strconv.Itoa(addrCount)
	}
	return result
}
func getSRVAddressTitle(rec *definitions.DNS_SRV_Record) string {
	if rec == nil {
		return ""
	}

	result := "SRV"
	addrCount := len(rec.Addresses)
	if addrCount != 1 {
		result += "=" + strconv.Itoa(addrCount)
	}
	return result
}

func printRecordHeader() {
	fmt.Printf("%20s %20s %20s\n", "Domain", "Name", "AvailableAddresses")
}
func printRecord(key string, record definitions.DNSRecord) {
	var name string
	if len(key) > (len(record.Domain) + 1) {
		name = name[:len(name)-len(record.Domain)-1]
		if name == "$" {
			name = "*"
		}
	}

	addresses := join(",",
		getIPAddressTitle("A", record.ARecords),
		getIPAddressTitle("AAAA", record.AAAARecords),
		getSTRAddressTitle("CNAME", record.CNameRecords),
		getSTRAddressTitle("NS", record.NSRecords),
		getSTRAddressTitle("TXT", record.TXTRecords),
		getMXAddressTitle(record.MXRecords),
		getSRVAddressTitle(record.SRVRecords))
	if addresses == "" {
		addresses = "<NO ADDRESSES>"
	}
	fmt.Printf("%20s %20s %s\n", record.Domain, name, addresses)
}

func getAddressFlags(addr definitions.DNS_Address) string {
	enabled_status := ""
	health_status := ""
	if !addr.Enabled {
		enabled_status = definitions.Console.Write(definitions.DarkRed, "DISABLED")
	} else {
		enabled_status = definitions.Console.Write(definitions.DarkGreen, "enabled")
	}

	if !addr.Healthy {
		health_status = definitions.Console.Write(definitions.DarkRed, "UNHEALTHY")
	} else {
		health_status = definitions.Console.Write(definitions.DarkGreen, "healthy")
	}

	return enabled_status + "," + health_status
}

func printIPRecords(name string, rec *definitions.DNS_IP_Record) {
	if rec == nil {
		return
	}
	for _, address := range rec.Addresses {
		fmt.Printf("    %5s %20s %6s %4ds %3d %s\n", name, address.IP, "", address.TTL,
			address.Weight, getAddressFlags(address.DNS_Address))
	}
}
func printSTRRecords(name string, rec *definitions.DNS_STRING_Record) {
	if rec == nil {
		return
	}
	for _, address := range rec.Addresses {
		fmt.Printf("    %5s %20s %6s %4ds %3d %s\n", name, address.Value, "", address.TTL,
			address.Weight, getAddressFlags(address.DNS_Address))
	}
}
func printMXRecords(rec *definitions.DNS_MX_Record) {
	if rec == nil {
		return
	}
	for _, address := range rec.Addresses {
		fmt.Printf("    %5s %20s (%4d) %4ds %3d %s\n", "MX",
			address.Value, address.Priority, address.TTL,
			address.Weight, getAddressFlags(address.DNS_Address))

	}
}
func printSRVRecords(rec *definitions.DNS_SRV_Record) {
	if rec == nil {
		return
	}
	for _, address := range rec.Addresses {
		srvAddress := address.Value + ":" + strconv.Itoa(int(address.Port))
		fmt.Printf("    %5s %20s (%4d) %4ds %3d %s\n", "SRV",
			srvAddress, address.Priority, address.TTL,
			address.Weight, getAddressFlags(address.DNS_Address))
	}
}

func getAllRecords(args CommandArgs) error {
	items, err := args.ReadRecordsByPattern("*")
	if err != nil {
		return err
	}
	if len(items) == 0 {
		definitions.Console.Print(definitions.DarkRed, "<No records found>\n")
		return nil
	}

	failedRecords := 0
	groupByDomain := make(map[string][]ReadRecordResponse)

	for _, item := range items {
		if item.Error != nil {
			failedRecords++
		} else {
			records, _ := groupByDomain[item.Record.Domain]
			records = append(records, item)
			groupByDomain[item.Record.Domain] = records
		}
	}

	if len(groupByDomain) == 0 {
		return fmt.Errorf("Failed to get %d records", failedRecords)
	}

	printRecordHeader()
	for _, items := range groupByDomain {
		for _, item := range items {
			printRecord(item.Key, item.Record)
			printIPRecords("A", item.Record.ARecords)
			printIPRecords("AAAA", item.Record.AAAARecords)
			printSTRRecords("CNAME", item.Record.CNameRecords)
			printSTRRecords("NS", item.Record.NSRecords)
			printSTRRecords("TXT", item.Record.TXTRecords)
			printMXRecords(item.Record.MXRecords)
			printSRVRecords(item.Record.SRVRecords)
		}
	}

	if failedRecords != 0 {
		definitions.Console.Printf("Failed to read %d records from REDIS", failedRecords)
	}

	return nil
}

func getDomainRecords(args CommandArgs) error {
	items, err := args.ReadRecordsByPattern("*")
	if err != nil {
		return err
	}
	if len(items) == 0 {
		definitions.Console.Print(definitions.DarkRed, "<No records found>\n")
		return nil
	}

	failedRecords := 0
	domainRecords := make([]ReadRecordResponse, 0, len(items))
	for _, item := range items {
		if item.Error != nil {
			failedRecords++
		} else {

		}
	}
}
