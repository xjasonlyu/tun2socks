// +build stats

package main

import (
	"flag"

	"github.com/xjasonlyu/tun2socks/component/stats/session"
	"github.com/xjasonlyu/tun2socks/log"
)

func init() {
	args.Stats = flag.Bool("stats", false, "Enable session statistics")
	args.StatsAddr = flag.String("statsAddr", "localhost:6001", "Listen address of stats, open in your browser to view statistics")

	addPostFlagsInitFn(func() {
		if *args.Stats {
			sessionStater = session.NewSimpleSessionStater()

			// Set stats variables
			session.ServeAddr = *args.StatsAddr

			// Start session stater
			if err := sessionStater.Start(); err != nil {
				log.Fatalf("Start session stater failed: %v", err)
			}
			log.Infof("Session stater serving at %v", session.ServeAddr)
		} else {
			sessionStater = nil
		}
	})
}
