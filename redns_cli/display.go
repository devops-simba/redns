package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/devops-simba/redns/definitions"
	log "github.com/golang/glog"
)

var (
	c = definitions.Console
)

const (
	color_OK      = definitions.DarkGreen
	color_BAD     = definitions.DarkOrange
	color_SUCCESS = definitions.DarkGreen
	color_ERR     = definitions.DarkRed
)

type DisplayContext interface {
	Info(msg string)
	Infof(format string, a ...interface{})
	Warn(msg string)
	Warnf(format string, a ...interface{})
	Error(msg string)
	Errorf(format string, a ...interface{})

	Print(msg string)
	Printf(format string, a ...interface{})

	PrintRecord(rec *definitions.DNSRecord, indent int)
	PrintAddressRecord(kind string, rec interface{}, indent int)
	PrintAddress(addr definitions.DNSAddressPtr, indent int)
}

type ColorPallette map[string]definitions.Color

func (this *ColorPallette) GetColor(name string, othernames ...string) (definitions.Color, bool) {
	color, ok := (*this)[name]
	if ok {
		return color, ok
	}

	for _, n := range othernames {
		color, ok = (*this)[n]
		if ok {
			return color, ok
		}
	}

	return color, false
}
func (this *ColorPallette) GetOr(name string, defaultColor definitions.Color, othernames ...string) definitions.Color {
	color, ok := (*this)[name]
	if ok {
		return color
	} else {
		for _, n := range othernames {
			color, ok = (*this)[n]
			if ok {
				return color
			}
		}
		return defaultColor
	}
}

type DefaultDisplayContext struct {
	Pallette     ColorPallette
	ColorContext definitions.ColorContext
}

func NewDisplayContext(pallette ColorPallette, colorContext definitions.ColorContext) DisplayContext {
	if pallette == nil {
		pallette := ColorPallette{
			"bad":     color_BAD,
			"error":   color_ERR,
			"ok":      color_OK,
			"success": color_SUCCESS,
		}
	}
	if colorContext == nil {
		colorContext = definitions.Console
	}

	return DefaultDisplayContext{
		Pallette:     pallette,
		ColorContext: colorContext,
	}
}

func (this DefaultDisplayContext) Info(msg string)                       { log.Info(msg) }
func (this DefaultDisplayContext) Infof(format string, a ...interface{}) { log.Infof(format, a...) }

func (this DefaultDisplayContext) Warn(msg string)                       { log.Warning(msg) }
func (this DefaultDisplayContext) Warnf(format string, a ...interface{}) { log.Warningf(format, a...) }

func (this DefaultDisplayContext) Error(msg string)                       { log.Error(msg) }
func (this DefaultDisplayContext) Errorf(format string, a ...interface{}) { log.Errorf(format, a...) }

func (this DefaultDisplayContext) Print(msg string)                       { fmt.Print(msg) }
func (this DefaultDisplayContext) Printf(format string, a ...interface{}) { fmt.Printf(format, a...) }

func (this DefaultDisplayContext) PrintAddress(addr definitions.DNSAddressPtr, indent int) {
	baseAddr := addr.Addr.BaseAddress()

	okColor := this.Pallette.GetOr("ok", color_OK, "success")
	badColor := this.Pallette.GetOr("bad", color_BAD, "error")

	enabled := ""
	if baseAddr.Enabled {
		enabled = this.ColorContext.Write(okColor, "Y")
	} else {
		enabled = this.ColorContext.Write(badColor, "N")
	}

	healthy := ""
	if baseAddr.Healthy {
		healthy = this.ColorContext.Write(okColor, "Y")
	} else {
		healthy = this.ColorContext.Write(badColor, "N")
	}

	priority := ""
	if addr.Kind == definitions.Kind_MX {
		mxAddr, _ := addr.Addr.(definitions.DNS_MX_Address)
		priority = strconv.Itoa(int(mxAddr.Priority))
	} else if addr.Kind == definitions.Kind_SRV {
		srvAddr, _ := addr.Addr.(definitions.DNS_SRV_Address)
		priority = strconv.Itoa(int(srvAddr.Priority))
	}

	this.Printf("%s%s%s %20s E:%s H:%s W:%d TTL:%d PRI:%s\n",
		strings.Repeat(" ", indent),
		addr.Kind,
		strings.Repeat(" ", 5-len(addr.Kind)),
		addr.Addr.GetValue(),
		enabled,
		healthy,
		int(baseAddr.Weight),
		baseAddr.TTL,
		priority)
}
func (this DefaultDisplayContext) PrintAddressRecord(kind string, rec interface{}, indent int) {
	this.Printf("%s%s%s ",
		strings.Repeat(" ", indent),
		kind,
		strings.Repeat(" ", 5-len(kind)),
	)

	weighted := false
	switch kind {
	case definitions.Kind_A, definitions.Kind_AAAA:
		rec, _ := rec.(*definitions.DNS_IP_Record)
		if rec.Weighted {
			this.Print("[LOAD BALANCED] ")
		}
		if len(rec.Addresses) == 0 {
			this.Printf("(NO ADDRESSES)\n")
			return
		}
		for _, addr := range rec.Addresses {
			this.PrintAddress(definitions.DNSAddressPtr{Kind: kind, Addr: addr}, indent+2)
		}
	case definitions.Kind_NS, definitions.Kind_TXT, definitions.Kind_CNAME:
		rec, _ := rec.(*definitions.DNS_STRING_Record)
		if rec.Weighted {
			this.Print("[LOAD BALANCED] ")
		}
		if len(rec.Addresses) == 0 {
			this.Printf("(NO ADDRESSES)\n")
			return
		}
		this.Print("\n")
		for _, addr := range rec.Addresses {
			this.PrintAddress(definitions.DNSAddressPtr{Kind: kind, Addr: addr}, indent+2)
		}
	case definitions.Kind_MX:
		rec, _ := rec.(*definitions.DNS_MX_Record)
		if rec.Weighted {
			this.Print("[LOAD BALANCED] ")
		}
		if len(rec.Addresses) == 0 {
			this.Printf("(NO ADDRESSES)\n")
			return
		}
		this.Print("\n")
		for _, addr := range rec.Addresses {
			this.PrintAddress(definitions.DNSAddressPtr{Kind: kind, Addr: addr}, indent+2)
		}
	case definitions.Kind_SRV:
		rec, _ := rec.(*definitions.DNS_SRV_Record)
		if len(rec.Addresses) == 0 {
			this.Printf("(NO ADDRESSES)\n")
			return
		}
		this.Print("\n")
		for _, addr := range rec.Addresses {
			this.PrintAddress(definitions.DNSAddressPtr{Kind: kind, Addr: addr}, indent+2)
		}
	}
}
func (this DefaultDisplayContext) PrintRecord(rec *definitions.DNSRecord, indent int) {

}
