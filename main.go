package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/xjasonlyu/tun2socks/v2/engine"
	"github.com/xjasonlyu/tun2socks/v2/log"

	"go.uber.org/automaxprocs/maxprocs"
)

var key = new(engine.Key)

func init() {
	flag.IntVar(&key.Mark, "fwmark", 0, "Set firewall MARK (Linux only)")
	flag.IntVar(&key.MTU, "mtu", 0, "Set device maximum transmission unit (MTU)")
	flag.IntVar(&key.UDPTimeout, "udp-timeout", 0, "Set timeout for each UDP session")
	flag.BoolVar(&key.Version, "version", false, "Show version information and quit")
	flag.StringVar(&key.Config, "config", "", "YAML format configuration file")
	flag.StringVar(&key.Device, "device", "", "Use this device [driver://]name")
	flag.StringVar(&key.Interface, "interface", "", "Use network INTERFACE (Linux/MacOS only)")
	flag.StringVar(&key.LogLevel, "loglevel", "info", "Log level [debug|info|warning|error|silent]")
	flag.StringVar(&key.Proxy, "proxy", "", "Use this proxy [protocol://]host[:port]")
	flag.StringVar(&key.Stats, "stats", "", "HTTP statistic server listen address")
	flag.StringVar(&key.Token, "token", "", "HTTP statistic server auth token")
	flag.Parse()
}

func main() {
	maxprocs.Set(maxprocs.Logger(func(string, ...any) {}))

	engine.Insert(key)

	checkErr := func(msg string, f func() error) {
		if err := f(); err != nil {
			log.Fatalf("Failed to %s: %v", msg, err)
		}
	}

	checkErr("start engine", engine.Start)
	defer checkErr("stop engine", engine.Stop)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
