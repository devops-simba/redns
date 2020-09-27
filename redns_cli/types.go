package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type DomainName []string
type SubdomainName = DomainName
type CommaSeparatedValue []string
type Kind []string
type Word uint16
type Bool3 uint8

const (
	domainChars       = "a-zA-Z0-9\\*\\?"
	AnyKind           = "*"
	AnyDomain         = "*"
	InvalidWord Word  = 0xFFFF
	True        Bool3 = 0xFF
	False       Bool3 = 0x01
	None        Bool3 = 0x00
)

var (
	domainPattern = regexp.MustCompile(
		fmt.Sprintf("^(?:[%s](?:[%s-]{0,61}[%s])?\\.)+[%s][%s-]{0,61}[%s]$",
			domainChars, domainChars, domainChars, domainChars, domainChars, domainChars))

	validKinds  = []string{"A", "AAAA", "NS", "CNAME", "TXT", "MX", "SRV"}
	trueValues  = []string{"t", "true", "y", "yes", "ok", "1"}
	falseValues = []string{"f", "false", "n", "no", "0"}
)

//region DomainName
func (this *DomainName) String() string { return strings.Join(*this, ",") }
func (this *DomainName) Set(value string) error {
	if len(value) == 0 {
		*this = []string{}
		return nil
	} else {
		*this = strings.Split(value, ",")
		for i, item := range *this {
			if !domainPattern.MatchString(item) {
				return fmt.Errorf("'%s' is not a valid domain name", item)
			} else {
				(*this)[i] = strings.ToLower(item)
			}
		}

		return nil
	}
}

//endregion

//region CommaSeparatedValue
func (this *CommaSeparatedValue) String() string { return strings.Join(*this, ",") }
func (this *CommaSeparatedValue) Set(value string) error {
	if value == "" {
		*this = []string{}
		return
	}

	*this = strings.Split(value, ",")
	return nil
}
func (this *CommaSeparatedValue) Contains(value string) bool { return Contains(*this, value) }
func (this *CommaSeparatedValue) Append(value ...string)     { *this = append(*this, value...) }

//endregion

//region Kind
func (this *Kind) String() string { return strings.Join(*this, ",") }
func (this *Kind) Set(value string) error {
	if value == "" {
		*this = []string{}
		return
	}

	haveAny := false
	*this = strings.Split(value, ",")
	for i, item := range *this {
		if item == AnyKind {
			haveAny = true
			continue
		}
		n := FindString(validKinds, strings.ToUpper(item))
		if n == -1 {
			return fmt.Errorf("'%s' is not a valid Kind. Accepted values are: [%s]",
				item, strings.Join(validKinds, ","))
		}

		(*this)[i] = validKinds[n]
	}

	if haveAny {
		*this = []string{AnyKind}
	}

	return nil
}
func (this *Kind) Any() bool { return len(*this) == 1 && (*this)[0] == AnyKind }
func (this *Kind) Contains(value string) bool {
	if this.Any() {
		return true
	}

	return Contains(*this, value)
}

//endregion

//region Word
func (this *Word) String() string { return strconv.Itoa(int(*this)) }
func (this *Word) Set(value string) error {
	n, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	if n == -1 {
		*this = InvalidWord
		return nil
	}
	if n < 0 || n >= 0xFFFF {
		return OutOfRange{}
	}

	*this = Word(n)
	return nil
}

//endregion

//region Bool3
func (this *Bool3) String() string {
	switch *this {
	case None:
		return "None"
	case True:
		return "True"
	case False:
		return "False"
	default:
		return ""
	}
}
func (this *Bool3) Set(value string) error {
	if Contains(trueValues, value) {
		*this = True
	} else if Contains(falseValues, value) {
		*this = False
	} else if value == "" {
		*this = None
	} else {
		return InvalidArgs{}
	}
	return nil
}

//endregion
