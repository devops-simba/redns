package main

import (
	"flag"

	log "github.com/golang/glog"
)

func main() {
	flag.Parse()

	options, err := NewControllerOptionsFromEnv()
	if err != nil {
		log.Fatalf("Failed to read the options: %v", err)
	}

	controller, err := NewController(options)
	if err != nil {
		log.Fatalf("Failed to start the controller: %v", err)
	}

}
