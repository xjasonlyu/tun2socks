package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	flag "github.com/spf13/pflag"
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
)

func init() {
	flag.IntVar(&key.Mark, "fwmark", 0, "Set firewall MARK (Linux/BSD)")
	flag.IntVar(&key.MTU, "mtu", 0, "Set device maximum transmission unit (MTU)")
	flag.DurationVar(&key.UDPTimeout, "udp-timeout", 0, "Set timeout for each UDP session")
	flag.StringVarP(&configFile, "config", "c", "", "YAML format configuration file")
	flag.StringVarP(&key.Device, "device", "d", "", "Use this device [driver://]name")
	flag.StringVarP(&key.Proxy, "proxy", "p", "", "Use this proxy [protocol://]host[:port]")
	flag.StringVar(&key.Interface, "interface", "", "Use network INTERFACE (Linux/MacOS/Windows)")
	flag.StringVar(&key.LogLevel, "loglevel", "info", "Log level [debug|info|warn|error|silent]")
	flag.StringVar(&key.RestAPI, "restapi", "", "HTTP statistic server listen address")
	flag.StringVar(&key.TCPSendBufferSize, "tcp-sndbuf", "", "Set TCP send buffer size for netstack")
	flag.StringVar(&key.TCPReceiveBufferSize, "tcp-rcvbuf", "", "Set TCP receive buffer size for netstack")
	flag.BoolVar(&key.TCPModerateReceiveBuffer, "tcp-auto-tuning", false, "Enable TCP receive buffer auto-tuning")
	flag.BoolVar(&key.TCPDeferredHandshake, "tcp-deferred-handshake", false, "Complete local TCP handshake only after outbound connect succeeds")
	flag.DurationVar(&key.TCPProxyHandshakeTimeout, "tcp-proxy-handshake-timeout", 0, "HTTP CONNECT/SOCKS4/SOCKS5 handshake timeout after proxy connection in deferred mode (0 uses 21s)")
	flag.BoolVar(&key.RejectIPv6, "reject-ipv6", false, "Locally reject IPv6 TCP with RST and UDP with ICMPv6 unreachable")
	flag.DurationVar(&key.TCPFailureCacheTTL, "tcp-failure-cache-ttl", 0, "Failed SYN cache lifetime in deferred mode (0 uses 30s; hits do not refresh)")
	flag.IntVar(&key.TCPFailureCacheCapacity, "tcp-failure-cache-capacity", 0, "Failed SYN cache capacity in deferred mode (0 uses 4096; shared by IPv4/IPv6)")
	flag.StringSliceVar(&key.MulticastGroups, "multicast-groups", nil, "Set multicast groups, separated by commas")
	flag.StringVar(&key.TUNPreUp, "tun-pre-up", "", "Execute a command before TUN device setup")
	flag.StringVar(&key.TUNPostUp, "tun-post-up", "", "Execute a command after TUN device setup")
	flag.BoolVarP(&versionFlag, "version", "v", false, "Show version and then quit")
	flag.Parse()
}

func main() {
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

	engine.StartOrFatal()
	defer engine.StopOrFatal()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
