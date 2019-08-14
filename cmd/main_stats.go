// +build stats

package main

import (
	"flag"

	"github.com/xjasonlyu/tun2socks/common/stats/session"
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
			sessionStater.Start()
		} else {
			sessionStater = nil
		}
	})
}
