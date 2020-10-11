package main

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/adjust/redismq"
	"github.com/hoisie/redis"
)

type RedisUrl struct {
	Host     string
	Port     int
	Password string
	Db       int
	Options  url.Values
}

func ParseRedisUrl(value string, defaultQueueName string) (*RedisUrl, error) {
	if !strings.HasPrefix(value, "redis://") {
		value = "redis://" + value
	}

	u, err := url.Parse(value)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(u.Host, ":")
	host := parts[0]

	if host == "" {
		host = "127.0.0.1"
	}

	var db, port int
	if len(u.Path) > 1 {
		path := u.Path[1:]
		db, err = strconv.Atoi(path)
		if err != nil {
			return nil, err
		}
	} else {
		db = 0
	}

	if len(parts) > 1 {
		port, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
	} else {
		port = 6379
	}

	var password string
	if u.User != nil {
		password, _ = u.User.Password()
	}

	options := u.Query()
	queueNames, ok := options["queue"]
	if len(queueNames) > 1 {
		return nil, errors.New("Multiple queue names is not supported")
	} else if (!ok || len(queueNames) == 0) && len(defaultQueueName) != 0 {
		options["queue"] = defaultQueueName
	}

	return &RedisUrl{
		Host:     host,
		Port:     port,
		Password: password,
		Db:       db,
		Options:  options,
	}, nil
}

func (this *RedisUrl) PortName() string { return strconv.Itoa(this.Port) }
func (this *RedisUrl) QueueName() string {
	queueNames, ok := this.Options["queue"]
	if !ok || len(queueNames) == 0 {
		return ""
	}
	return queueNames[0]
}
func (this *RedisUrl) CreateClient() *redis.Client {
	return &redis.Client{
		Addr:     net.JoinHostPort(this.Host, this.PortName()),
		Password: this.Password,
		Db:       this.Db,
	}
}
func (this *RedisUrl) CreateQueue() *redismq.Queue {
	return redismq.CreateQueue(this.Host, this.PortName(), this.Password, int64(this.Db), this.QueueName())
}
func (this *RedisUrl) SelectQueue() (*redismq.Queue, error) {
	return redismq.SelectQueue(this.Host, this.PortName(), this.Password, int64(this.Db), this.QueueName())
}
