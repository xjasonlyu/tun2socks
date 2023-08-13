package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"go.uber.org/automaxprocs/maxprocs"
	"gopkg.in/yaml.v3"

	_ "github.com/xjasonlyu/tun2socks/v2/dns"
	"github.com/xjasonlyu/tun2socks/v2/engine"
	"github.com/xjasonlyu/tun2socks/v2/internal/version"
	"github.com/xjasonlyu/tun2socks/v2/log"
)

var (
	key = new(engine.Key)

	configFile  string
	versionFlag bool

	autoSetup bool
)

func init() {
	flag.IntVar(&key.Mark, "fwmark", 0, "Set firewall MARK (Linux only)")
	flag.IntVar(&key.MTU, "mtu", 0, "Set device maximum transmission unit (MTU)")
	flag.DurationVar(&key.UDPTimeout, "udp-timeout", 0, "Set timeout for each UDP session")
	flag.StringVar(&configFile, "config", "", "YAML format configuration file")
	flag.StringVar(&key.Device, "device", "", "Use this device [driver://]name")
	flag.StringVar(&key.Interface, "interface", "", "Use network INTERFACE (Linux/MacOS only)")
	flag.StringVar(&key.LogLevel, "loglevel", "info", "Log level [debug|info|warning|error|silent]")
	flag.StringVar(&key.Proxy, "proxy", "", "Use this proxy [protocol://]host[:port]")
	flag.StringVar(&key.RestAPI, "restapi", "", "HTTP statistic server listen address")
	flag.StringVar(&key.TCPSendBufferSize, "tcp-sndbuf", "", "Set TCP send buffer size for netstack")
	flag.StringVar(&key.TCPReceiveBufferSize, "tcp-rcvbuf", "", "Set TCP receive buffer size for netstack")
	flag.BoolVar(&key.TCPModerateReceiveBuffer, "tcp-auto-tuning", false, "Enable TCP receive buffer auto-tuning")
	flag.StringVar(&key.TUNPreUp, "tun-pre-up", "", "Execute a command before TUN device setup")
	flag.StringVar(&key.TUNPostUp, "tun-post-up", "", "Execute a command after TUN device setup")
	flag.BoolVar(&versionFlag, "version", false, "Show version and then quit")
	flag.BoolVar(&autoSetup, "auto-setup", false, "Auto setup TUN device (Linux only)")
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

	if autoSetup {
		commands := []string{
			fmt.Sprintf("tuntap add mode tun dev %s", key.Device),
			fmt.Sprintf("addr add 10.10.10.10/24 dev %s", key.Device),
			fmt.Sprintf("link set dev %s up", key.Device),
			fmt.Sprintf("route add default dev %s metric 1", key.Device),
		}
		for _, cmd := range commands {
			if err := exec.Command("ip", strings.Split(cmd, " ")...).Run(); err != nil {
				log.Fatalf("Failed to setup TUN device: %v", err)
			}
		}

		defer func() {
			// Run shell command to delete TUN device
			commands = []string{
				fmt.Sprintf("link set dev %s down", key.Device),
				fmt.Sprintf("tuntap del mode tun dev %s", key.Device),
				fmt.Sprintf("route del default dev %s metric 1", key.Device),
			}

			for _, cmd := range commands {
				if err := exec.Command("ip", strings.Split(cmd, " ")...).Run(); err != nil {
					log.Fatalf("Failed to delete TUN device: %v", err)
				}
			}
		}()

	}

	engine.Insert(key)

	engine.Start()
	defer engine.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
