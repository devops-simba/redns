package main

import (
	"os"

	log "github.com/golang/glog"

	"github.com/devops-simba/redns/definitions/signals"
)

func main() {
	controller := createController()

	err := controller.Start()
	if err != nil {
		log.Fatalf("Failed to start controller: %v", err)
	}

	waitForStop()
	log.Infof("Received stop signal, stopping the controller")

	controller.Stop()
}

func createController() *Controller {
	kubeConfig := os.Getenv("KUBECONFIG_PATH")
	if kubeConfig == "" {
		kubeConfig = os.Getenv("HOME") + "/.kube/config"
	}

	nodeId := os.Getenv("NODE_ID")
	if nodeId == "" {
		log.Fatalf("Missing node ID, please set NODE_ID env")
	}

	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		log.Fatalf("Missing redis URL, please set REDIS_URL env")
	}

	controller, err := NewController(kubeConfig, nodeId, "redns-lock", "default")
	if err != nil {
		log.Fatalf("Failed to create the controller: %v", err)
	}

	return controller
}
func waitForStop() {
	stopped := make(chan struct{})
	signals.SetupSignalHandler(1)

	<-stopped
}

func parseRedisUrl(redisUrl string) {

}
