package main

import (
	"fmt"
	"os"

	"github.com/openshift/kubernetes/staging/src/k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	NODE_ID         = "NODE_ID"
	REDIS_URL       = "REDIS_URL"
	KUBECONFIG_PATH = "KUBECONFIG_PATH"
	ELECTION_LOCK   = "ELECTION_LOCK"
	CHANGE_QUEUE    = "CHANGE_QUEUE"
)

var (
	DefaultKubeConfigPath = os.Getenv("HOME") + "/.kube/config"
)

type ControllerOptions struct {
	NodeId                 string
	LockName               string
	RedisDbUrl             *RedisUrl
	ChangedDnsObjectsQueue *RedisUrl
	KubeConfig             *rest.Config
}

func NewControllerOptionsFromEnv() (*ControllerOptions, error) {
	nodeId := os.Getenv(NODE_ID)
	if nodeId == "" {
		return nil, fmt.Errorf("Node id is required, please set `%s` environment variable to node ID", NODE_ID)
	}

	redisUrlValue := ReadEnv(REDIS_URL, "127.0.0.1")
	redisUrl, err := ParseRedisUrl(redisUrlValue, "")
	if err != nil {
		return nil, err
	}

	changedDnsObjectsUrl, err := ParseRedisUrl(ReadEnv(CHANGE_QUEUE, redisUrlValue), "changes")
	if err != nil {
		return nil, err
	}

	configPath := ReadEnv(KUBECONFIG_PATH, DefaultKubeConfigPath)
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", this.KubeConfigPath)
		if err != nil {
			return err
		}
	}

	return &ControllerOptions{
		NodeId:                 nodeId,
		KubeConfig:             config,
		RedisDbUrl:             redisUrl,
		ChangedDnsObjectsQueue: changedDnsObjectsUrl,
		LockName:               ReadEnv(NODE_ID, "redns-lock"),
	}, nil
}
