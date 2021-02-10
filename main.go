package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/xjasonlyu/tun2socks/engine"
	"github.com/xjasonlyu/tun2socks/log"

	flag "github.com/spf13/pflag"
)

var key = new(engine.Key)

func init() {
	flag.StringVarP(&key.Device, "device", "d", "", "use this device [driver://]name")
	flag.IntVar(&key.Mark, "fwmark", 0, "set firewall MARK (Linux only)")
	flag.StringVarP(&key.Interface, "interface", "i", "", "use network INTERFACE (Linux/MacOS only)")
	flag.StringVarP(&key.LogLevel, "loglevel", "l", "info", "log level [debug|info|warn|error|silent]")
	flag.Uint32VarP(&key.MTU, "mtu", "m", 0, "set device maximum transmission unit (MTU)")
	flag.StringVarP(&key.Proxy, "proxy", "p", "", "use this proxy [protocol://]host[:port]")
	flag.StringVar(&key.Stats, "stats", "", "HTTP statistic server listen address")
	flag.StringVar(&key.Token, "token", "", "HTTP statistic server auth token")
	flag.BoolVarP(&key.Version, "version", "v", false, "show version information and quit")
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
