package main

import (
	"encoding/json"
	"fmt"
	"hash/adler32"
	"hash/crc32"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	rednsclientset "github.com/devops-simba/redns/definitions/client/clientset/versioned"
)

const (
	lastAppliedVersionAnnotation = "devops.snapp.ir/last-applied-version"
)

func ReadEnv(envName, defaultValue string) string {
	value, ok := os.LookupEnv(envName)
	if !ok {
		value = defaultValue
	}
	return value
}

func getKubeClients(configPath string) (kubernetes.Interface, rednsclientset.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		if configPath == "" {
			configPath = os.Getenv("HOME") + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			return nil, nil, err
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	rednsClient, err := rednsclientset.NewForConfig(config)
	return client, rednsClient, err
}

func getLastAppliedVersion(obj *metav1.ObjectMeta) string {
	result, _ := obj.Annotations[lastAppliedVersionAnnotation]
	return result
}
func setLastAppliedVersion(obj *metav1.ObjectMeta, value string) {
	obj.Annotations[lastAppliedVersionAnnotation] = value
}

func computeObjectVersion(obj interface{}) string {
	s, _ := json.Marshal(obj)

	adler_hash := adler32.New()
	adler_hash.Write(s)

	crc_hash := crc32.New(crc32.MakeTable(crc32.IEEE))
	crc_hash.Write(s)

	return fmt.Sprintf("%s/%s", adler_hash.Sum32(), crc_hash.Sum32())
}
