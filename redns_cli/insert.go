package main

import (
	"errors"
	"net"
	"strings"

	"github.com/devops-simba/redns/definitions"
)

type InsertCommand bool

const (
	AddCommand = InsertCommand(true)
	SetCommand = InsertCommand(false)
)

func (this InsertCommand) normalizeDomain(context DisplayContext, args *CommandArgs) error {
	// RULES:
	//	1- Domain is required
	//  2- All records must belong to same domain
	//  3- Domain name can't be a wildcard
	// TASKS:
	//	1- Domain name must be lower case
	if len(args.Domain) == 0 {
		return errors.New("Missing domain for new address to add")
	}
	if len(args.Domain) != 1 || IsWildcard(args.Domain[0]) {
		return errors.New("Invalid domain name")
	}

	// domain name is not case sensitive
	args.Domain[0] = strings.ToLower(args.Domain[0])

	return nil
}
func (this InsertCommand) normalizeName(context DisplayContext, args *CommandArgs) error {
	// RULES:
	//	1- Exactly one name is required
	//  2- '*', '@' are acceptable
	//  3- Name can't be wildcard
	// TASKS:
	//  1- Names will be converted to lower case
	if len(args.Name) == 0 {
		return errors.New("Missing name of new address")
	}
	if len(args.Name) != 1 {
		return errors.New("Bad number of names")
	}
	if args.Name[0] != "*" && args.Name[0] != "@" && IsWildcard(args.Name[0]) {
		return errors.New("Invalid name")
	}
	args.Name[0] = strings.ToLower(args.Name[0])

	return nil
}
func (this InsertCommand) normalizeKind(context DisplayContext, args *CommandArgs) error {
	if len(args.Kind) == 0 {
		return errors.New("Missing value kind")
	}
	if len(args.Value) == 0 {
		return errors.New("Missing value")
	}
	if len(args.Kind) != 1 && len(args.Kind) != len(args.Value) {
		return errors.New("Invalid number of kind")
	}
	for _, kind := range args.Kind {
		if kind == AnyKind {
			return errors.New("Invalid kind")
		}
	}
	if len(args.Kind) != len(args.Value) {
		kind := args.Kind[0]
		args.Kind = make(Kind, len(args.Name))
		for i := 0; i < len(args.Name); i++ {
			args.Kind[i] = kind
		}
	}

	return nil
}
func (this InsertCommand) normalizeValue(context DisplayContext, args *CommandArgs) error {
	if len(args.Value) == 0 {
		return errors.New("Missing value")
	}
	for i := 0; i < len(args.Value); i++ {
		kind := args.Kind[i]
		value := args.Value[i]
		var ip net.IP
		switch kind {
		case definitions.Kind_A:
			if !IsIPv4(value) {
				return errors.New("Invalid IPv4")
			}
		case definitions.Kind_AAAA:
			if !IsIPv6(value) {
				return errors.New("Invalid IPv6")
			}
		case definitions.Kind_NS:
			if !IsDomainName(value) {
				return errors.New("Invalid NS(must be a domain name)")
			}
		case definitions.Kind_CNAME:
			if !IsDomainName(value) {
				return errors.New("Invalid CNAME(must be a domain name)")
			}
		case definitions.Kind_TXT:
		case definitions.Kind_MX:
			if !IsDomainName(value) {
				return errors.New("Invalid MX(must be a domain name)")
			}
		case definitions.Kind_SRV:
			_, _, err := ParseSRV(value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
func (this InsertCommand) normalizePriority(context DisplayContext, args *CommandArgs) error {
	if args.Priority != InvalidWord {
		ok := false
		for _, kind := range args.Kind {
			if kind == definitions.Kind_SRV || kind == definitions.Kind_MX {
				ok = true
				break
			}
		}
		if !ok {
			context.Warn("Priority is not supported by any of the provided values")
			args.Priority = InvalidWord
		}
	}

	return nil
}
func (this InsertCommand) Normalize(context DisplayContext, args *CommandArgs) error {
	err := this.normalizeDomain(context, args)
	if err != nil {
		return err
	}

	err = this.normalizeName(context, args)
	if err != nil {
		return err
	}

	err = this.normalizeKind(context, args)
	if err != nil {
		return err
	}

	err = this.normalizeValue(context, args)
	if err != nil {
		return err
	}

	err = this.normalizePriority(context, args)
	if err != nil {
		return err
	}

	return nil
}
func (this InsertCommand) Execute(context DisplayContext, args CommandArgs) error {
	var rec definitions.DNSRecord
	if this {
		// we should add all information to current record
		prec, err := args.ReadRecord(args.Domain[0], args.Name[0])
		if err != nil {
			return err
		}
		rec = prec.DNSRecord
	} else {
		rec.Domain = args.Domain[0]
	}

	for i, kind := range args.Kind {
		found := false
		value := args.Value[i]
		var addr *definitions.DNS_Address
		var ipAddr *definitions.DNS_IP_Address
		var strAddr *definitions.DNS_STRING_Address
		var mxAddr *definitions.DNS_MX_Address
		var srvAddr *definitions.DNS_SRV_Address
		switch kind {
		case definitions.Kind_A:
			ipAddr, _ = AddIPAddress(&rec.ARecords, value)
			addr = &ipAddr.DNS_Address
		case definitions.Kind_AAAA:
			AddIPAddress(&rec.AAAARecords, value)
			addr = &ipAddr.DNS_Address
		case definitions.Kind_NS:
			strAddr, _ = AddSTRAddress(&rec.NSRecords, value, false)
			addr = &strAddr.DNS_Address
		case definitions.Kind_TXT:
			strAddr, _ = AddSTRAddress(&rec.TXTRecords, value, true)
			addr = &strAddr.DNS_Address
		case definitions.Kind_CNAME:
			strAddr, _ = AddSTRAddress(&rec.CNameRecords, value, false)
			addr = &strAddr.DNS_Address
		case definitions.Kind_MX:
			mxAddr, _ = AddMXAddress(&rec.MXRecords, value)
			if args.Priority != InvalidWord {
				mxAddr.Priority = uint16(args.Priority)
			}
			addr = &mxAddr.DNS_Address
		case definitions.Kind_SRV:
			srv, port, err := ParseSRV(value)
			if err != nil {
				return err
			}
			srvAddr, _ = AddSRVAddress(&rec.SRVRecords, srv, uint16(port))
			if args.Priority != InvalidWord {
				srvAddr.Priority = uint16(args.Priority)
			}
			addr = &srvAddr.DNS_Address
		}

		if args.Enabled != None {
			addr.Enabled = args.Enabled.Bool(true)
		}
		if args.Healthy != None {
			addr.Healthy = args.Healthy.Bool(true)
		}
		if args.Weight != InvalidWord {
			addr.Weight = uint16(args.Weight)
		}
		if args.TTL != InvalidDWord {
			addr.TTL = uint32(args.TTL)
		}
	}

	args.WriteRecord(&rec, args.Domain[0], args.Name[0])

	return nil
}
