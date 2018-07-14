package manualNotify

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type journalHandler struct {
	channel         chan byte
	dnsUpdateNotify chan int
	zone            string
	is_signed       bool
}

func NewJournalHandler(src chan byte, zone string, is_signed bool, dnsUpdateNotify chan int) *journalHandler {
	return &journalHandler{
		channel:         src,
		zone:            zone,
		is_signed:       is_signed,
		dnsUpdateNotify: dnsUpdateNotify,
	}
}

func (handler *journalHandler) readJournalMsg() {
	buf := make([]byte, 512)
	idx := 0
	for ; idx <= 511; idx++ {
		buf[idx] = <-handler.channel
		if buf[idx] == '\n' {
			break
		}
	}
	if idx <= 511 {
		strbuf := string(buf[:idx])
		extra_str := ""
		if handler.is_signed {
			extra_str = " (signed)"
		}
		sval := fmt.Sprintf("zone %v/IN/external%v: sending notifies (serial", handler.zone, extra_str)
		if strings.Contains(strbuf, sval) {
			strbuf = strbuf[len(sval):]
			strbuf = strings.Trim(strbuf, " ()")
			i, err := strconv.Atoi(strbuf)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"strbuf": strbuf,
					"sval":   sval,
				}).Warn("Couldn't extract zone id from string")
			} else {
				// Send update
				logrus.WithFields(logrus.Fields{
					"zone": handler.zone,
					"serial": i,
				}).Info("Detected notify statement")
				handler.dnsUpdateNotify <- i
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"strbuf": strbuf,
				"sval":   sval,
			}).Debug("Didn't detect notify statement")
		}
	} else {
		fmt.Errorf("Line longer than 512 bytes, discarded!")
	}
}

func (handler *journalHandler) Spin() {
	for {
		handler.readJournalMsg()
	}
}
