package manualNotify

import (
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"net"
	"time"
	"strings"
)

type dnsClient interface {
	Exchange(m *dns.Msg, address string) (r *dns.Msg, rtt time.Duration, err error)
}

type dnsClientFactory interface {
	Create() dnsClient
}

type defaultDNSClient struct{}

func (*defaultDNSClient) Create() dnsClient {
	return new(dns.Client)
}

type dnsHandler struct {
	channel        chan int
	zone           string
	resolv_conf    string
	my_name        string
	destination    string
	client_factory dnsClientFactory
}

func NewDNSHandler(channel chan int, zone string, resolv_conf string, my_name string, destination string) *dnsHandler {
	return &dnsHandler{
		channel:        channel,
		zone:           zone,
		resolv_conf:    resolv_conf,
		my_name:        my_name,
		destination:    destination,
		client_factory: &defaultDNSClient{},
	}
}

func (handler *dnsHandler) getSoa() (string, error) {
	dns_config, err := dns.ClientConfigFromFile(handler.resolv_conf)
	if err != nil {
		logrus.WithError(err).WithField("resolv", handler.resolv_conf).Warn("Couldn't load dns client config, skipping this round!")
		return "", err
	}
	dns_address := net.JoinHostPort(dns_config.Servers[0], dns_config.Port)
	dns_client := handler.client_factory.Create()

	// Step 1: Figure out if this name server is responsible (look at SOA)
	soa_query_msg := new(dns.Msg)
	if !strings.HasSuffix(handler.zone, ".") {
		handler.zone = handler.zone + "."
	}
	soa_query_msg.SetQuestion(handler.zone, dns.TypeSOA)
	soa_query_msg.RecursionDesired = true
	soa_query, _, err := dns_client.Exchange(soa_query_msg, dns_address)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"server": dns_address,
		}).WithError(err).Warn("Couldn't read SOA record, skipping this round!")
		return "", err
	}
	if soa_query.Rcode != dns.RcodeSuccess {
		logrus.WithFields(logrus.Fields{
			"rcode":  soa_query.Rcode,
			"server": dns_address,
		}).Warn("Couldn't read SOA record, skipping this round!")
		return "", err
	}
	for _, record := range soa_query.Answer {
		logrus.WithFields(logrus.Fields{
			"answer": record.String(),
			"header": record.Header(),
		}).Debug("Got SOA")
		if soa, ok := record.(*dns.SOA); ok {
			return soa.Ns, nil
		} else {
			logrus.WithFields(logrus.Fields{
				"answer": record.String(),
				"header": record.Header(),
			}).Warn("Got SOA answer which is not a SOA record!")
		}
	}
	return "", nil
}

func (handler *dnsHandler) sendNotify() (bool, error) {
	soa, err := handler.getSoa()
	if err != nil {
		return false, err
	}

	if !strings.HasSuffix(handler.my_name, ".") {
		handler.my_name = handler.my_name + "."
	}

	if soa == handler.my_name {
		msg := new(dns.Msg)
		msg.SetNotify(handler.zone)
		dns_client := handler.client_factory.Create()
		msg, _, err := dns_client.Exchange(msg, handler.destination)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"my_name": handler.my_name,
				"soa":     soa,
				"zone":    handler.zone,
				"answer":  msg,
			}).WithError(err).Error("Couldn't send notify!")
			return false, err
		}
		logrus.WithFields(logrus.Fields{
			"my_name": handler.my_name,
			"soa":     soa,
			"zone":    handler.zone,
			"answer":  msg,
		}).Info("NOTIFY sent successfully")
		return true, nil
	} else {
		logrus.WithFields(logrus.Fields{
			"my_name": handler.my_name,
			"soa":     soa,
		}).Info("SOA doesn't match hostname, skipping!")
	}
	return false, nil
}

func (handler *dnsHandler) Spin() {
	for {
		<-handler.channel
		handler.sendNotify()
	}
}
