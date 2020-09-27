package main

import (
	"github.com/devops-simba/redns/definitions"
)

type ListCommand struct{}
type addressPtrWithRecord struct {
	Kind string
	Addr definitions.IDNSAddress
	Rec  *definitions.DNSRecord
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
			if args.TTL != InvalidWord && dnsAddress.TTL != args.TTL {
				continue
			}
			if args.Enabled != None && dnsAddress.Enabled != args.Enabled {
				continue
			}
			if args.Healthy != None && dnsAddress.Healthy != args.Healthy {
				continue
			}
			if args.Weight != InvalidWord && dnsAddress.Weight != args.Weight {
				continue
			}
			switch addressPtr.Kind {
			case Kind_A, Kind_AAAA, Kind_NS, Kind_TXT, Kind_CNAME:
				if args.Priority != InvalidWord {
					continue
				}
			case Kind_MX:
				if args.Priority != InvalidWord {
					mxAddr, _ := addressPtr.Address.(definitions.DNS_MX_Address)
					if mxAddr.Priority != mxAddr.Priority {
						continue
					}
				}
			case Kind_SRV:
				if args.Priority != InvalidWord {
					srvAddr, _ := addressPtr.Address.(definitions.DNS_SRV_Address)
					if srvAddr.Priority != srvAddr.Priority {
						continue
					}
				}
			}
			if len(args.Value) != 0 && !args.Value.Contains(addressPtr.Addr.GetValue()) {
				continue
			}

			addresses = append(addresses, addressPtrWithRecord{
				Kind: addressPtr.Kind,
				Addr: addressPtr.Addr,
				Rec:  rec,
			})
		}
	}

	return nil
}
