package main

import (
	"strings"

	"github.com/devops-simba/redns/definitions"
)

type ListCommand struct{}
type DNSAddressWithRecord struct {
	Addr definitions.IDNSAddress
	Rec  *DNSRecordWithKey
}

func (this ListCommand) Normalize(context DisplayContext, args *CommandArgs) error {
	if len(args.Domain) == 0 {
		args.Domain = []string{AnyDomain}
	}
	if len(args.Name) == 0 {
		args.Name = []string{AnyName}
	}
	if len(args.Kind) == 0 {
		args.Kind = Kind{AnyKind}
	}

	return nil
}

func (this ListCommand) Execute(context DisplayContext, args CommandArgs) error {
	records, err := args.FindRecords(context)
	if err != nil {
		return err
	}

	var addresses []DNSAddressWithRecord
	for i := 0; i < len(records); i++ {
		rec := &records[i]
		for _, address := range rec.GetAddresses() {
			if !args.Kind.Contains(address.GetKind()) {
				continue
			}
			dnsAddress := address.BaseAddress()
			if args.TTL != InvalidDWord && dnsAddress.TTL != uint32(args.TTL) {
				continue
			}
			if args.Enabled != None && dnsAddress.Enabled != args.Enabled.Bool() {
				continue
			}
			if args.Healthy != None && dnsAddress.Healthy != args.Healthy.Bool() {
				continue
			}
			if args.Weight != InvalidWord && dnsAddress.Weight != uint16(args.Weight) {
				continue
			}
			if args.Priority != InvalidWord && address.GetPriority() != uint16(args.Priority) {
				continue
			}
			if len(args.Value) != 0 && !args.Value.Contains(address.GetValue()) {
				continue
			}

			addresses = append(addresses, DNSAddressWithRecord{
				Addr: address,
				Rec:  rec,
			})
		}
	}

	// now display this addresses
	var last_rec *DNSRecordWithKey
	for _, addr := range addresses {
		// if this is a new record, print its header
		if addr.Rec != last_rec {
			last_rec = addr.Rec
			key := last_rec.Key
			if strings.HasPrefix(key, "$.") {
				key = "*." + key[2:]
			}
			context.Printf("%s(%s):\n", key, last_rec.Domain)
		}

		// now print the address
		context.Printf("  - ")
		context.PrintAddress(addr.Addr, 4)
	}

	return nil
}
