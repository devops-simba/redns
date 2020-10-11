package main

import (
	"k8s.io/client-go/kubernetes"
	"sync"
	"time"

	"github.com/hoisie/redis"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	rednsclientset "github.com/devops-simba/redns/definitions/client/clientset/versioned"
)

const (
	maxRetries = 5

	stoppedStatus int32 = 0
	runningStatus int32 = 1
	leaderStatus  int32 = 2

	startingStatus int32 = 10
	stoppingStatus int32 = 11
)

type eventType struct {
	Key          string
	EventType    string
	Namespace    string
	ResourceType string
}

type Controller struct {
	// this will be used to update REDNS objects in the REDIS
	redisClient *redis.Client
	// this will be used to update REDNS objects in the REDIS
	rednsClient rednsclientset.Interface
	// a flag that indicate we are leader
	leader int32
	// this function will be called to stop leader elector
	stopFunc func()
	// this will be used to stop informers
	stopChannel chan struct{}
	// this will be used to watch for changes in DNSRecord objects
	recordInformer cache.SharedIndexInformer
	// this will be used to watch for changes in DNSLoadBalancer objects
	loadbalancerInformer cache.SharedIndexInformer
	// this will be used to wait for completion of different parts
	stopped sync.WaitGroup
}

func NewController(options *ControllerOptions) (*Controller, error) {
	controller := &Controller{}

	var err error
	controller.rednsClient, err = rednsclientset.NewForConfig(options.KubeConfig)
	if err != nil {
		return nil, err
	}

	return controller, nil
}

//region Leader Election
func (this *Controller) onBecomeLeader() {
	//
}
func (this *Controller) onStoppedLeading() {
	//
}
func (this *Controller) startLeaderElector(options *ControllerOptions) error {
	restConfig, err := kubernetes.NewForConfig(options.)
	config := leaderelection.LeaderElectionConfig{
		Lock: &resourcelock.LeaseLock{
			Client:     result.Client.CoordinationV1(),
			LeaseMeta:  metav1.ObjectMeta{Name: options.LockName, Namespace: rest.NamespaceNone},
			LockConfig: resourcelock.ResourceLockConfig{Identity: options.NodeId},
		},
		ReleaseOnCancel: true,
		LeaseDuration:   15 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: onBecomeLeader,
			OnStoppedLeading: onStoppedLeading,
			OnNewLeader: func(id string) {
				if id == lock.LockConfig.Identity {
					onBecomeLeader(nil)
				} else if this.IsLeader {
					onStoppedLeading()
				} else {
					this.Callbacks.OnLeaderChanged(id)
				}
			},
		},
	}
}

//endregion
