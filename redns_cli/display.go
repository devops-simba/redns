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
	PrintAddressRecord(rec definitions.IDNSAddressRecord, indent int)
	PrintAddress(addr definitions.IDNSAddress, indent int)
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
		pallette = ColorPallette{
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

func (this DefaultDisplayContext) PrintAddress(addr definitions.IDNSAddress, indent int) {
	baseAddr := addr.BaseAddress()

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
	nPriority := addr.GetPriority()
	if nPriority != definitions.PriorityIsNotSupported {
		priority = "PRI:" + strconv.Itoa(int(nPriority))
	}

	value := addr.GetValue()
	if len(value) < 30 {
		value += strings.Repeat(" ", 30-len(value))
	}

	kind := addr.GetKind()
	this.Printf("%s%s%s %s E:%s H:%s W:%d TTL:%d%s\n",
		strings.Repeat(" ", indent),
		kind,
		strings.Repeat(" ", 5-len(kind)),
		value,
		enabled,
		healthy,
		int(baseAddr.Weight),
		baseAddr.TTL,
		priority)
}
func (this DefaultDisplayContext) PrintAddressRecord(rec definitions.IDNSAddressRecord, indent int) {
	kind := rec.GetItemKind()
	this.Printf("%s%s%s",
		strings.Repeat(" ", indent),
		kind,
		strings.Repeat(" ", 5-len(kind)),
	)

	if rec.IsWeighted() {
		this.Print(" [LOAD BALANCED]")
	}

	addresses := rec.AddressList()
	if len(addresses) == 0 {
		this.Print(" (EMPTY)\n")
		return
	}

	this.Print("\n")
	for i := 0; i < len(addresses); i++ {
		this.PrintAddress(addresses[i], 0)
	}
}
func (this DefaultDisplayContext) PrintRecord(rec *definitions.DNSRecord, indent int) {
	sIndent := strings.Repeat(" ", indent)
	this.Printf("%sDomain: %s\n", sIndent, rec.Domain)
	addresses := rec.GetAddresses()
	for i := 0; i < len(addresses); i++ {
		this.Printf("%s- ", sIndent)
		this.PrintAddress(addresses[i], 0)
	}
}
