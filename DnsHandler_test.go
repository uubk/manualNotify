package manualNotify

import (
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
	"time"
)

type mockDNSClient struct {
	numCalled   int
	lastMessage dns.Msg
}

func (obj *mockDNSClient) Exchange(m *dns.Msg, address string) (r *dns.Msg, rtt time.Duration, err error) {
	obj.numCalled++
	obj.lastMessage = *m
	if obj.numCalled != 1 {
		return m, 0, nil
	} else {
		msg := new(dns.Msg)
		msg.Answer = []dns.RR{
			&dns.SOA{
				Ns: "ns1.google.com.",
			},
		}
		return msg, 0, nil
	}
}

type debugDNSClient struct {
	obj *mockDNSClient
}

func (this *debugDNSClient) Create() dnsClient {
	if this.obj == nil {
		this.obj = &mockDNSClient{
			numCalled: 0,
		}
	}
	return this.obj
}

func TestSOARead(t *testing.T) {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.ErrorLevel)

	my_chan := make(chan int, 5)
	uut := NewDNSHandler(my_chan, "google.ch.", "/etc/resolv.conf", "ns1.google.com.", "127.0.0.1")
	result, err := uut.getSoa()
	if err != nil {
		t.Error("Unexpected error return", err)
	}
	if result != "ns1.google.com." {
		t.Error("Unexpected return value", result)
	}
}

func TestSendNotify(t *testing.T) {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.ErrorLevel)

	my_chan := make(chan int, 5)
	uut := NewDNSHandler(my_chan, "google.ch.", "/etc/resolv.conf", "ns1.google.com.", "127.0.0.1:53")
	factory := &debugDNSClient{
		obj: nil,
	}
	uut.client_factory = factory
	rc, err := uut.sendNotify()
	if err != nil {
		t.Error("Unexpected error return", err)
	}
	if factory.obj.numCalled != 2 {
		t.Error("Unexpected number of calls to client: ", factory.obj.numCalled)
	}
	if !rc {
		t.Error("Notify unsuccessful!")
	}
}
