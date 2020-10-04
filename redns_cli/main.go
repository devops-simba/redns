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
		fmt.Println("	remove  Remove addresses or records")
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
	case "remove":
		command = RemoveCommand{}
	default:
		flag.Parse()
		log.Error("Unknown command.")
		os.Exit(2)
	}

	args := NewCommandArgs()
	args.BindFlags(flag.CommandLine)
	err := flag.CommandLine.Parse(os.Args[2:])
	if err != nil {
		log.Errorf("Failed to parse the arguments: %v", err)
		os.Exit(2)
	}

	context := NewDisplayContext(nil, nil)
	err = command.Normalize(context, &args)
	if err != nil {
		log.Errorf("Invalid command argument: %v", err)
		os.Exit(2)
	}

	err = command.Execute(context, args)
	if err != nil {
		log.Errorf("Error in executing the command: %v", err)
		os.Exit(1)
	}
}
