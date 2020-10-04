package main

import (
	"errors"

	"github.com/devops-simba/redns/definitions"
)

type RemoveCommand struct{}

func (this RemoveCommand) Normalize(context DisplayContext, args *CommandArgs) error {
	if len(args.Domain) == 0 {
		return errors.New("Domain is required")
	}
	if len(args.Name) == 0 {
		args.Name = []string{AnyName}
	}
	if len(args.Kind) == 0 {
		args.Kind = Kind{AnyKind}
	}
	return nil
}
func (this RemoveCommand) Execute(context DisplayContext, args CommandArgs) error {
	records, err := args.FindRecords(context)
	if err != nil {
		return err
	}

	if !args.Kind.Any() || args.Enabled != None ||
		args.Healthy != None || len(args.Value) != 0 ||
		args.TTL != InvalidDWord || args.Priority != InvalidWord ||
		args.Weight != InvalidWord {
		return removeAddresses(context, args, records)
	} else {
		return removeRecords(context, args, records)
	}
}

func shouldRemove(args CommandArgs, addr definitions.IDNSAddress) bool {
	if !args.Kind.Contains(addr.GetKind()) {
		return false
	}

	baseAddr := addr.BaseAddress()
	if args.Enabled != None && args.Enabled.Bool() != baseAddr.Enabled {
		return false
	}
	if args.Healthy != None && args.Healthy.Bool() != baseAddr.Healthy {
		return false
	}
	if args.TTL != InvalidDWord && uint32(args.TTL) != baseAddr.TTL {
		return false
	}
	if args.Weight != InvalidWord && uint16(args.Weight) != baseAddr.Weight {
		return false
	}
	if args.Priority != InvalidWord && uint16(args.Priority) != addr.GetPriority() {
		return false
	}
	if len(args.Value) != 0 && !Contains(args.Value, addr.GetValue()) {
		return false
	}

	return true
}
func removeAddresses(context DisplayContext, args CommandArgs, records []DNSRecordWithKey) error {
	for i := 0; i < len(records); i++ {
		changed := false
		rec := records[i]
		for j := 0; j < rec.ARecords.Length(); j++ {
			if shouldRemove(args, &rec.ARecords.Addresses[j]) {
				rec.ARecords = RemoveRecord_A(rec.ARecords, j)
				changed = true
				j--
			}
		}
		for j := 0; j < rec.AAAARecords.Length(); j++ {
			if shouldRemove(args, &rec.AAAARecords.Addresses[j]) {
				rec.AAAARecords = RemoveRecord_AAAA(rec.AAAARecords, j)
				changed = true
				j--
			}
		}
		for j := 0; j < rec.NSRecords.Length(); j++ {
			if shouldRemove(args, &rec.NSRecords.Addresses[j]) {
				rec.NSRecords = RemoveRecord_NS(rec.NSRecords, j)
				changed = true
				j--
			}
		}
		for j := 0; j < rec.TXTRecords.Length(); j++ {
			if shouldRemove(args, &rec.TXTRecords.Addresses[j]) {
				rec.TXTRecords = RemoveRecord_TXT(rec.TXTRecords, j)
				changed = true
				j--
			}
		}
		for j := 0; j < rec.CNameRecords.Length(); j++ {
			if shouldRemove(args, &rec.CNameRecords.Addresses[j]) {
				rec.CNameRecords = RemoveRecord_CNAME(rec.CNameRecords, j)
				changed = true
				j--
			}
		}
		for j := 0; j < rec.MXRecords.Length(); j++ {
			if shouldRemove(args, &rec.MXRecords.Addresses[j]) {
				rec.MXRecords = RemoveRecord_MX(rec.MXRecords, j)
				changed = true
				j--
			}
		}
		for j := 0; j < rec.SRVRecords.Length(); j++ {
			if shouldRemove(args, &rec.SRVRecords[j]) {
				rec.SRVRecords = RemoveRecord_SRV(rec.SRVRecords, j)
				changed = true
				j--
			}
		}

		if !changed {
			continue
		}
		if rec.ARecords == nil && rec.AAAARecords == nil &&
			rec.NSRecords == nil && rec.TXTRecords == nil && rec.CNameRecords == nil &&
			rec.MXRecords == nil && rec.SRVRecords == nil {
			ok, err := args.Redis.Del(rec.Key)
			if err != nil {
				context.Errorf("Failed to remove `%s`: %v\n", rec.Key, err)
			} else if ok {
				context.Infof("Removed `%s`\n", rec.Key)
			}
		} else {
			err := args.WriteRecordByKey(&rec.DNSRecord, rec.Key)
			if err == nil {
				context.Infof("Updated `%s`\n", rec.Key)
			} else {
				context.Errorf("Failed to update `%s`: %v\n", rec.Key, err)
			}
		}
	}

	return nil
}
func removeRecords(context DisplayContext, args CommandArgs, records []DNSRecordWithKey) error {
	for i := 0; i < len(records); i++ {
		rec := records[i]
		ok, err := args.Redis.Del(rec.Key)
		if err != nil {
			context.Errorf("Failed to remove key `%s`: %v\n", rec.Key, err)
		} else if ok {
			context.Infof("Successfully removed `%s`", rec.Key)
		}
	}

	return nil
}
