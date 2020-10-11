package main

import (
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	log "github.com/golang/glog"
)

type HealthChecker interface {
	Check(value string) bool
}

type HealthCheckRecord struct {
	Checker HealthChecker
	Value   string
}

//region TCP HealthCheck
type tcpHealthCheck struct{}

func (this tcpHealthCheck) Check(value string) bool {
	c, err := net.Dial("tcp", value)
	if err == nil {
		c.Close()
		return true
	}
	return false
}

//endregion

//region ICMP HealthCheck
const (
	icmpProtocol    = 1
	icmpMaxRetries  = 4
	icmpMessageData = "HealthCheck"
)

type icmpHealthCheck struct{}

func (this icmpHealthCheck) Check(value string) bool {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Errorf("Error in listening for ICMP packages: %v", err)
		return false
	}
	defer c.Close()

	dst, err := net.ResolveIPAddr("ip4", value)
	if err != nil {
		log.Errorf("Can't resolve `%s` as ICMP address: %v", value, err)
		return false
	}

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte(icmpMessageData),
		},
	}

	buffer, err := msg.Marshal(nil)
	if err != nil {
		log.Errorf("Error in marshaling ICMP message: %v", err)
		return false
	}

	replyBuffer := make([]byte, 1500)
	for i := 0; i < icmpMaxRetries; i++ {
		n, err := c.WriteTo(buffer, dst)
		if err != nil {
			log.Errorf("%d) Failed to send ICMP data to `%s`: %v", i+1, value, err)
			continue
		} else if n != len(buffer) {
			log.Errorf("%d) Falied to send all of the buffer to `%s`: got %v, expected %v", i+1, value, n, len(buffer))
			continue
		}

		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			err = c.SetReadDeadline(deadline)
			if err != nil {
				log.Warningf("Failed to set read timeout on ICMP connection: %v", err)
				continue
			}

			n, peer, err := c.ReadFrom(replyBuffer)
			if err != nil {
				log.Warningf("%d) Failed to read ICMP response from `%s`: %v", i+1, value, err)
				continue
			}

			if peer.Network() != dst.Network() || peer.String() != dst.String() {
				log.Warningf("%d) Received a message from an invalid peer: %s", i+1, peer.String())
				continue
			}

			reply, err := icmp.ParseMessage(icmpProtocol, replyBuffer[:n])
			if err != nil {
				log.Warningf("%d) Received an invalid message from the peer: %v", i+1, err)
				continue
			}

			switch reply.Type {
			case ipv4.ICMPTypeEchoReply:
				log.Infof("%d) Received %v from the server", i+1, reply.Body)

			default:
				log.Warningf("%d) Received invalid kind of message. got %+v, expected echo reply", i+1, reply)
			}

			return true
		}

		time.Sleep(100 * time.Millisecond)
	}

	return false
}

//endregion

//region HTTP HealthCheck
type httpHealthCheck struct{}

func (this httpHealthCheck) Check(value string) bool {
	resp, err := http.DefaultClient.Get(value)
	if err == nil {
		if resp.StatusCode == 200 {
			return true
		}
		if resp.StatusCode > 200 && resp.StatusCode < 300 {
			log.Infof("HealthCheck of `%s` returned non 200 success code: %d(%s)", value, resp.StatusCode, resp.Status)
			return true
		}
	}
	return false
}

//endregion

var (
	TCP  HealthChecker = tcpHealthCheck{}
	ICMP HealthChecker = icmpHealthCheck{}
	HTTP HealthChecker = httpHealthCheck{}
)
