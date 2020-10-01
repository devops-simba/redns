package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/golang/glog"
)

func main() {
	flag.Usage = func() {
		fmt.Println("action Action that should executed by this tool. Available actions are:")
		fmt.Println("	list    List records and addresses based on enetered criteria")
		fmt.Println("	add     Add one or more addresses to a record")
		fmt.Println("	set     Replace content of a record with addresses that specified in this command")
		flag.PrintDefaults()
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(2)
	}

	var command Command
	action := os.Args[1]
	switch action {
	case "list":
		command = ListCommand{}
	case "add":
		command = AddCommand
	case "set":
		command = SetCommand
	}

	args := NewCommandArgs()
	flagset := flag.NewFlagSet("CommandArgs", flag.ExitOnError)
	args.BindFlags(flagset)
	err := flagset.Parse(os.Args[2:])
	if err != nil {
		log.Errorf("Failed to parse the arguments: %v", err)
		os.Exit(2)
	}

}
