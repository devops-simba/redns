package main

import (
	"regexp"

	"github.com/devops-simba/redns/definitions"
)

var (
	dot = regexp.QuoteMeta(".")
)

type DNSRecordWithKey struct {
	definitions.DNSRecord
	Key string
}

type Command interface {
	// Normalize normalize and validate command arguments
	Normalize(context DisplayContext, args *CommandArgs) error
	// Execute this command
	Execute(context DisplayContext, args CommandArgs) error
}
