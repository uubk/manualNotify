package main

import (
	"github.com/coreos/go-systemd/sdjournal"
	"fmt"
	"time"
	"github.com/uubk/manualNotify"
	"math"
	"github.com/sirupsen/logrus"
	"os"
	"flag"
	"sync"
	"github.com/coreos/go-systemd/daemon"
)

func main() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	verbose := flag.Bool("verbose", false, "Enable verbose output")
	cfgLoc := flag.String("config", "config.yml", "Config file location")
	flag.Parse()

	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	cfg, err := manualNotify.LoadConfigFromFile(*cfgLoc)
	if err != nil {
		logrus.WithError(err).WithField("cfgLoc", *cfgLoc).Fatal("Couldn't load config!")
		os.Exit(-1)
	}
	logrus.WithFields(logrus.Fields{
		"cfg": cfg,
	}).Debug("Config loaded")

	wg := sync.WaitGroup{}
	for _, zone := range cfg.Zones {
		journalCfg := sdjournal.JournalReaderConfig{
			Since: -time.Duration(5)*time.Second,
			Matches: []sdjournal.Match{
				{"SYSLOG_IDENTIFIER", cfg.Unitname},
			},
			Formatter: func(entry *sdjournal.JournalEntry) (string, error) {
				if *verbose {
					logrus.WithFields(logrus.Fields{
						"entry": *entry,
						"message": entry.Fields["MESSAGE"],
					}).Debug("Message from journal received")
				}
				return fmt.Sprintln(entry.Fields["MESSAGE"]), nil
			},
		}
		msgs := manualNotify.NewChannelWriter()
		logrus.WithFields(logrus.Fields{
			"name": zone.Name,
			"destination": zone.Destination,
			"isSigned": zone.IsSigned,
		}).Info("Handling zone")

		if journal, err := sdjournal.NewJournalReader(journalCfg) ; err == nil {
			wg.Add(1)
			go func () {
				defer wg.Done()
				journal.Follow(time.After(time.Duration(math.MaxInt64)), msgs)
			}()

			dnsHandlerChan := make(chan int, 1)
			journalHandler := manualNotify.NewJournalHandler(msgs.Get(), zone.Name, zone.IsSigned, dnsHandlerChan)
			wg.Add(1)
			go func () {
				defer wg.Done()
				journalHandler.Spin()
			}()
			dnsHandler := manualNotify.NewDNSHandler(dnsHandlerChan, zone.Name, cfg.Resolvconf, cfg.Hostname, zone.Destination)
			wg.Add(1)
			go func () {
				defer wg.Done()
				dnsHandler.Spin()
			} ()
		} else {
			logrus.WithError(err).WithFields(logrus.Fields{
				"name": zone.Name,
				"destination": zone.Destination,
				"isSigned": zone.IsSigned,
			}).Fatal("Couldn't follow journal!")
			os.Exit(-1)
		}
	}
	daemon.SdNotify(true, daemon.SdNotifyReady)
	wg.Wait()
}