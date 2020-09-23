package main

import (
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/golang/glog"
)

func main() {
	port := flag.Uint("port", 53, "Port that we should listen on it")
	redisServerUrl := flag.String("redis", "", "Address of the redis server in format `redis://[:password]@]host:port[/db-number][?option=value]`")
	flag.Parse()

	if *port == 0 || *port > 65535 {
		log.Fatalf("%v is not a valid port number", *port)
	}
	if len(*redisServerUrl) == 0 {
		log.Fatal("Missing redis db address")
	}

	db, err := NewRedisDNSDatabase(*redisServerUrl)
	if err != nil {
		log.Fatalf("Error in opening REDIS db: %v", err)
	}

	stopRequestedChan := make(chan os.Signal, 1)
	signal.Notify(stopRequestedChan, syscall.SIGINT, syscall.SIGTERM)

	server := NewDNSServer(db, strconv.Itoa(int(*port)), "udp")
	serverStopped := runServer(server)
	select {
	case <-stopRequestedChan:
		log.Info("Got OS shutdown signal, shutting down DNS server gracefully...")
		server.Shutdown()

	case err = <-serverStopped:
		log.Warningf("Server stopped unexpectedly: %v", err)
	}
}

func runServer(server *DNSServer) chan error {
	stopped := make(chan error, 1)
	go func() {
		err := server.Start()
		if err != nil {
			log.Errorf("Error in starting server: %v", err)
		}
		stopped <- err
	}()

	return stopped
}
