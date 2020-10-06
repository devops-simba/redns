package main

import (
	"context"
	"sync/atomic"
	"time"

	log "github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	rednsclientset "github.com/devops-simba/redns/definitions/client/clientset/versioned"
)

type Controller struct {
	NodeId      string
	Client      kubernetes.Interface
	RednsClient rednsclientset.Interface

	isLeader int32
	cancel   func()
	lock     *resourcelock.LeaseLock
}

func NewController(configPath string, nodeId string, lockName string, redisUrl string) (*Controller, error) {
	result := &Controller{NodeId: nodeId}

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}

	result.Client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	result.RednsClient, err = rednsclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	result.lock = &resourcelock.LeaseLock{
		LeaseMeta:  metav1.ObjectMeta{Name: lockName, Namespace: ns},
		Client:     result.Client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{Identity: nodeId},
	}

	return result, nil
}

func (this *Controller) onBecomeLeader() {
	if atomic.SwapInt32(&this.isLeader, 1) == 0 {
		log.Infof("%s: Start working as the leader", this.NodeId)
		// start looking for changed resources
	}
}
func (this *Controller) onStoppedLeading() bool {
	if atomic.SwapInt32(&this.isLeader, 0) == 1 {
		log.Infof("%s: We are not leader anymore", this.NodeId)
		// stop looking for changed resources
		return true
	} else {
		return false
	}
}
func (this *Controller) onLeaderChanged(leaderId string) {
	if leaderId == this.NodeId {
		this.onBecomeLeader()
	} else if !this.onStoppedLeading() {
		log.Infof("Leader changed to `%s`", leaderId)
	}
}

func (this *Controller) Start() error {
	lec := leaderelection.LeaderElectionConfig{
		Lock:            this.lock,
		ReleaseOnCancel: true,
		LeaseDuration:   15 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(c context.Context) { this.onBecomeLeader() },
			OnStoppedLeading: func() { this.onStoppedLeading() },
			OnNewLeader:      func(id string) { this.onLeaderChanged(id) },
		},
	}

	le, err := leaderelection.NewLeaderElector(lec)
	if err != nil {
		return err
	}

	var ctx context.Context
	ctx, this.cancel = context.WithCancel(context.Background())
	go le.Run(ctx)

	return nil
}
func (this *Controller) Stop() {
	this.cancel()
	for {
		// wait until we are not leader anymore
		if atomic.LoadInt32(&this.isLeader) == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}
