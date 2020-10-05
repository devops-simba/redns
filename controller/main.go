package main

import (
	"os"

	log "github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	rednsclientset "github.com/devops-simba/redns/controller/pkg/apis/redns/client/clientset/versioned"
)

func createClients() (kubernetes.Interface, rednsclientset.Interface) {
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

	// create the config from path
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Fatalf("Failed to load kube config: %v", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Faild to create kubernetes client: %v", err)
	}

	rednsClient, err := rednsclientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create redns client: %v", err)
	}

	return client, rednsClient
}

func main() {
	var rednsClient rednsclientset.Interface
}
