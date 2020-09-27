package main

import (
	"github.com/devops-simba/redns/definitions"
)

type ListCommand struct{}
type addressPtrWithRecord struct {
	Kind string
	Ptr  *definitions.DNS_Address
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
		for _, addressPtr := rec.GetAddresses() {
			if !args.Kind.Contains(addressPtr.Kind) {
				continue
			}

			addresses = append(addresses, addressPtrWithRecord{
				Kind: addressPtr.Kind,
				Ptr: addressPtr.Ptr,
				Rec: rec,
			})
		}
	}

	return nil
}
