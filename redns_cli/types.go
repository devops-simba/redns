package main

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/hoisie/redis"
)

type DomainName []string
type SubdomainName []string
type CommaSeparatedValue []string
type Kind []string
type Word uint16
type DWord uint32
type Bool3 uint8

const (
	domainCharsWithWC       = "a-zA-Z0-9\\*\\?"
	AnyKind                 = "*"
	AnyDomain               = "*"
	AnyName                 = "*"
	InvalidWord       Word  = 0xFFFF
	InvalidDWord      DWord = 0xFFFFFFFF
	True              Bool3 = 0xFF
	False             Bool3 = 0x01
	None              Bool3 = 0x00
)

var (
	domainPattern = regexp.MustCompile(
		fmt.Sprintf("^(?:\\*|(?:[%s](?:[%s-]{0,61}[%s])?\\.)+[%s][%s-]{0,61}[%s])$",
			domainCharsWithWC, domainCharsWithWC, domainCharsWithWC,
			domainCharsWithWC, domainCharsWithWC, domainCharsWithWC))

	subDomainPattern = regexp.MustCompile(
		fmt.Sprintf("^(?:\\*|@|(?:(?:[%s](?:[%s-]{0,61}[%s])?)(?:\\.[%s][%s-]{0,61}[%s]))*)$",
			domainCharsWithWC, domainCharsWithWC, domainCharsWithWC,
			domainCharsWithWC, domainCharsWithWC, domainCharsWithWC))

	validKinds  = []string{"A", "AAAA", "NS", "CNAME", "TXT", "MX", "SRV"}
	trueValues  = []string{"t", "true", "y", "yes", "ok", "1"}
	falseValues = []string{"f", "false", "n", "no", "0"}
)

//region RedisValue
// RedisValue this is a simple helper class to read redis configuration from command line
type RedisValue redis.Client

func NewRedisValue(client *redis.Client) *RedisValue {
	client.Addr = "127.0.0.1:6379"
	client.Password = ""
	client.Db = 0
	return (*RedisValue)(client)
}

func (this *RedisValue) String() string {
	result := "redis://"
	if this.Password != "" {
		result += ":" + this.Password + "@"
	}
	result += this.Addr
	if this.Db != 0 {
		result += "/" + strconv.Itoa(this.Db)
	}
	return result
}
func (this *RedisValue) Set(value string) error {
	var password, host string
	var db, port int

	if !strings.HasPrefix(value, "redis://") {
		value = "redis://" + value
	}

	u, err := url.Parse(value)
	if err != nil {
		return err
	}

	parts := strings.Split(u.Host, ":")
	host = parts[0]

	if host == "" {
		host = "127.0.0.1"
	}

	if len(u.Path) > 1 {
		path := u.Path[1:]
		db, err = strconv.Atoi(path)
		if err != nil {
			return err
		}
	} else {
		db = 0
	}

	if len(parts) > 1 {
		port, err = strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
	} else {
		port = 6379
	}

	if u.User != nil {
		password, _ = u.User.Password()
	}

	this.Addr = net.JoinHostPort(host, strconv.Itoa(port))
	this.Password = password
	this.Db = db
	return nil
}

//endregion

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

//region SubdomainName
func (this *SubdomainName) String() string { return strings.Join(*this, ",") }
func (this *SubdomainName) Set(value string) error {
	if len(value) == 0 {
		*this = []string{}
		return nil
	} else {
		*this = strings.Split(value, ",")
		for i, item := range *this {
			if !subDomainPattern.MatchString(item) {
				return fmt.Errorf("'%s' is not a valid subdomain name", item)
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
		return nil
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
		return nil
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

func (this Word) Value() uint16 { return uint16(this) }
func (this Word) ValueOr(defaultValue uint16) uint16 {
	if this == InvalidWord {
		return defaultValue
	} else {
		return uint16(this)
	}
}

//endregion

//region DWord
func (this *DWord) String() string { return strconv.FormatUint(uint64(*this), 10) }
func (this *DWord) Set(value string) error {
	n, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	if n == -1 {
		*this = InvalidDWord
		return nil
	}
	if n < 0 {
		return OutOfRange{}
	}

	*this = DWord(n)
	return nil
}

func (this DWord) Value() uint32 { return uint32(this) }
func (this DWord) ValueOr(defaultValue uint32) uint32 {
	if this == InvalidDWord {
		return defaultValue
	} else {
		return uint32(this)
	}
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
func (this Bool3) Bool() bool { return this == True }
func (this Bool3) BoolOr(defaultValue bool) bool {
	if this == None {
		return defaultValue
	} else {
		return this == True
	}
}

//endregion
