// +build session

package main

import (
	"flag"

	"github.com/xjasonlyu/tun2socks/component/session"
	"github.com/xjasonlyu/tun2socks/log"
)

func init() {
	args.EnableStats = flag.Bool("stats", false, "Enable session statistics monitor")
	args.StatsAddr = flag.String("statsAddr", "localhost:6001", "Listen address of session monitor, open in your browser to view statistics")

	registerInitFn(func() {
		if *args.EnableStats {
			monitor = session.NewServer()

			// Set stats variables
			session.ServeAddr = *args.StatsAddr

			// Start session stater
			if err := monitor.Start(); err != nil {
				log.Fatalf("Start session monitor failed: %v", err)
			}
			log.Infof("Session monitor serving at %v", session.ServeAddr)
		} else {
			monitor = nil
		}
	})
}
