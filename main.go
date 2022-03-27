package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/xjasonlyu/tun2socks/v2/engine"
	"github.com/xjasonlyu/tun2socks/v2/internal/version"
	"github.com/xjasonlyu/tun2socks/v2/log"

	"go.uber.org/automaxprocs/maxprocs"
	"gopkg.in/yaml.v3"
)

var (
	key = new(engine.Key)

	configFile  string
	versionFlag bool
)

func init() {
	flag.BoolVar(&versionFlag, "version", false, "Show version and then quit")
	flag.IntVar(&key.Mark, "fwmark", 0, "Set firewall MARK (Linux only)")
	flag.IntVar(&key.MTU, "mtu", 0, "Set device maximum transmission unit (MTU)")
	flag.IntVar(&key.UDPTimeout, "udp-timeout", 0, "Set timeout for each UDP session")
	flag.StringVar(&configFile, "config", "", "YAML format configuration file")
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

	if versionFlag {
		fmt.Println(version.String())
		fmt.Println(version.BuildString())
		os.Exit(0)
	}

	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			log.Fatalf("Failed to read config file '%s': %v", configFile, err)
		}
		if err = yaml.Unmarshal(data, key); err != nil {
			log.Fatalf("Failed to unmarshal config file '%s': %v", configFile, err)
		}
	}

	engine.Insert(key)

	engine.Start()
	defer engine.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
