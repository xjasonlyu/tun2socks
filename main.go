package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/xjasonlyu/tun2socks/engine"
	"github.com/xjasonlyu/tun2socks/log"
)

var key = new(engine.Key)

func init() {
	flag.IntVar(&key.Mark, "fwmark", 0, "Set firewall MARK (Linux only)")
	flag.IntVar(&key.MTU, "mtu", 0, "Set device maximum transmission unit (MTU)")
	flag.IntVar(&key.UDPTimeout, "udp-timeout", 0, "Set timeout for each UDP session")
	flag.BoolVar(&key.Version, "version", false, "Show version information and quit")
	flag.StringVar(&key.Device, "device", "", "Use this device [driver://]name")
	flag.StringVar(&key.Interface, "interface", "", "Use network INTERFACE (Linux/MacOS only)")
	flag.StringVar(&key.LogLevel, "loglevel", "info", "Log level [debug|info|warn|error|silent]")
	flag.StringVar(&key.Proxy, "proxy", "", "Use this proxy [protocol://]host[:port]")
	flag.StringVar(&key.Stats, "stats", "", "HTTP statistic server listen address")
	flag.StringVar(&key.Token, "token", "", "HTTP statistic server auth token")
	flag.BoolVar(&key.RemoteDNS, "remote-dns", false, "Enable remote DNS (SOCKS5 and HTTP)")
	flag.StringVar(&key.RemoteDNSNetIPv4, "remote-dns-net-ipv4", "169.254.0.0/16",
		"IPv4 network for remote DNS A records")
	flag.StringVar(&key.RemoteDNSNetIPv6, "remote-dns-net-ipv6", "fd80:dead:beef:badc:0ded:c0de:ba5e::/112",
		"IPv6 network for remote DNS AAAA records")
	flag.Parse()
}

func main() {
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
