// +build stats

package main

import (
	"flag"

	"github.com/xjasonlyu/tun2socks/common/stats/session"
)

func init() {
	args.Stats = flag.Bool("stats", false, "Enable statistics")
	args.StatsAddr = flag.String("statsAddr", "localhost:6001", "Listen address of stats, open in your browser to view statistics")

	addPostFlagsInitFn(func() {
		if *args.Stats {
			// stats variables
			session.StatsAddr = *args.StatsAddr
			session.StatsVersion = version

			sessionStater = session.NewSimpleSessionStater()
			sessionStater.Start()
		} else {
			sessionStater = nil
		}
	})
}
