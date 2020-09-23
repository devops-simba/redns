package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Bool3 string

const (
	True  = Bool3("true")
	False = Bool3("false")
	None  = Bool3("")
)

var (
	trueValues  = []string{"t", "true", "y", "yes", "ok", "1"}
	falseValues = []string{"f", "false", "n", "no", "0"}
)

func (this *Bool3) String() string { return string(*this) }
func (this *Bool3) Set(value string) error {
	if Contains(trueValues, value) {
		*this = "true"
	} else if Contains(falseValues, value) {
		*this = "false"
	} else if value == "" {
		*this = ""
	} else {
		return errors.New("Invalid boolean value")
	}
	return nil
}

type Word uint16

const (
	ZeroWord    = Word(0)
	InvalidWord = Word(0xFFFF)
)

func (this *Word) String() string { return strconv.Itoa(int(*this)) }
func (this *Word) Set(value string) error {
	n, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	if n < 0 || n > 0xFFFF {
		return fmt.Errorf("Value(%v) is out of range [0, %v]", n, 0xFFFF)
	}

	*this = Word(uint16(n))
	return nil
}

type Command string

const (
	EmptyCommand  = Command("")
	GetCommand    = Command("get")
	AddCommand    = Command("add")
	SetCommand    = Command("set")
	RemoveCommand = Command("remove")
)

var (
	ValidCommands = []string{
		string(GetCommand),
		string(AddCommand),
		string(SetCommand),
		string(RemoveCommand),
	}
)

func (this *Command) String() string { return string(*this) }
func (this *Command) Set(value string) error {
	if Contains(ValidCommands, value) {
		*this = Command(value)
		return nil
	}

	return fmt.Errorf("%v is not a valid command. Accepted values are {%s}", value,
		strings.Join(ValidCommands, "|"))
}

type Kind string

const (
	Kind_Empty = Kind("")
	Kind_A     = Kind("A")
	Kind_AAAA  = Kind("AAAA")
	Kind_CNAME = Kind("CNAME")
	Kind_NS    = Kind("NS")
	Kind_TXT   = Kind("TXT")
	Kind_MX    = Kind("MX")
	Kind_SRV   = Kind("SRV")
)

var (
	ValidKinds = []string{
		string(Kind_A),
		string(Kind_AAAA),
		string(Kind_CNAME),
		string(Kind_NS),
		string(Kind_TXT),
		string(Kind_MX),
		string(Kind_SRV),
	}
)

func (this *Kind) String() string { return string(*this) }
func (this *Kind) Set(value string) error {
	if Contains(ValidKinds, value) {
		*this = Kind(value)
		return nil
	}

	return fmt.Errorf("%v is not a valid record kind. Acceptable values are {%s}",
		value, strings.Join(ValidKinds, "|"))
}

type IP string

const EmptyIP = IP("")

func (this *IP) String() string { return string(*this) }
func (this *IP) Set(value string) error {
	if value == "" {
		*this = IP(value)
		return nil
	} else {
		if net.ParseIP(value) != nil {
			*this = IP(value)
			return nil
		}

		return fmt.Errorf("%v is not a valid IP", value)
	}
}
