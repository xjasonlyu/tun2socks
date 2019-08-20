// +build session

package main

import (
	"flag"

	"github.com/xjasonlyu/tun2socks/component/session"
	"github.com/xjasonlyu/tun2socks/log"
)

func init() {
	args.EnableMonitor = flag.Bool("monitor", false, "Enable session statistics monitor")
	args.MonitorAddr = flag.String("monitorAddr", "localhost:6001", "Listen address of session monitor, open in your browser to view statistics")

	registerInitFn(func() {
		if *args.EnableMonitor {
			monitor = session.New(*args.MonitorAddr)

			// Start session monitor
			if err := monitor.Start(); err != nil {
				log.Fatalf("Start session monitor failed: %v", err)
			}
		} else {
			monitor = nil
		}
	})
}
