package main

import (
	"github.com/devops-simba/redns/definitions"
)

type ListCommand struct{}
type addressPtrWithRecord struct {
	definitions.DNSAddressPtr
	Rec *DNSRecordWithKey
}

func (this ListCommand) Normalize(context DisplayContext, args *CommandArgs) error {
	if len(args.Domain) == 0 {
		args.Domain = []string{AnyDomain}
	}
	if len(args.Name) == 0 {
		args.Name = []string{AnyDomain}
	}
	if len(args.Kind) == 0 {
		args.Kind = []string{AnyKind}
	}

	return nil
}

func (this ListCommand) Execute(context DisplayContext, args CommandArgs) error {
	records, err := args.FindRecords(context)
	if err != nil {
		return err
	}

	var addresses []addressPtrWithRecord
	for _, rec := range records {
		for _, addressPtr := range rec.GetAddresses() {
			if !args.Kind.Contains(addressPtr.Kind) {
				continue
			}
			dnsAddress := addressPtr.Addr.BaseAddress()
			if args.TTL != InvalidDWord && dnsAddress.TTL != uint32(args.TTL) {
				continue
			}
			if args.Enabled != None && dnsAddress.Enabled != args.Enabled.Bool(true) {
				continue
			}
			if args.Healthy != None && dnsAddress.Healthy != args.Healthy.Bool(true) {
				continue
			}
			if args.Weight != InvalidWord && dnsAddress.Weight != uint16(args.Weight) {
				continue
			}
			if args.Priority != InvalidWord {
				switch addressPtr.Kind {
				case definitions.Kind_MX:
					mxAddr, _ := addressPtr.Addr.(definitions.DNS_MX_Address)
					if mxAddr.Priority != mxAddr.Priority {
						continue
					}
				case definitions.Kind_SRV:
					srvAddr, _ := addressPtr.Addr.(definitions.DNS_SRV_Address)
					if srvAddr.Priority != srvAddr.Priority {
						continue
					}
				default:
					// rest of the kinds does not support priority
					continue
				}
			}
			if len(args.Value) != 0 && !args.Value.Contains(addressPtr.Addr.GetValue()) {
				continue
			}

			addresses = append(addresses, addressPtrWithRecord{
				DNSAddressPtr: addressPtr,
				Rec:           &rec,
			})
		}
	}

	// now display this addresses
	var last_rec *DNSRecordWithKey
	for _, addr := range addresses {
		if addr.Rec != last_rec {
			last_rec = addr.Rec
			context.Printf("%s(%s):\n", last_rec.Key, last_rec.Domain)
		}

		// not print the address
		context.Printf("  - ")
		context.PrintAddress(addr.DNSAddressPtr, 4)
	}

	return nil
}
