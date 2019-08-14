package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/xjasonlyu/tun2socks/common/dns"
	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/stats"
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/filter"
	"github.com/xjasonlyu/tun2socks/proxy"
	"github.com/xjasonlyu/tun2socks/tun"

	// init logger
	_ "github.com/xjasonlyu/tun2socks/common/log/simple"
)

const MTU = 1500

var (
	version     = "unknown version"
	description = "A tun2socks implementation written in Go."

	args            = new(CmdArgs)
	postFlagsInitFn []func()

	fakeDNS       dns.FakeDNS
	sessionStater stats.SessionStater
)

type CmdArgs struct {
	// Main
	Version  *bool
	TunName  *string
	TunAddr  *string
	TunGw    *string
	TunMask  *string
	TunDNS   *string
	LogLevel *string

	// Proxy
	ProxyServer *string
	UdpTimeout  *time.Duration

	// FakeDNS
	EnableFakeDNS *bool
	DNSCacheSize  *int
	FakeIPRange   *string
	FakeDNSAddr   *string
	FakeDNSHosts  *string

	// Stats
	Stats     *bool
	StatsAddr *string
}

func addPostFlagsInitFn(fn func()) {
	postFlagsInitFn = append(postFlagsInitFn, fn)
}

func init() {
	// Main
	args.Version = flag.Bool("version", false, "Print version")
	args.LogLevel = flag.String("loglevel", "info", "Logging level. (info, warning, error, debug, silent)")
	args.TunName = flag.String("tunName", "tun0", "TUN interface name")
	args.TunAddr = flag.String("tunAddr", "240.0.0.2", "TUN interface address")
	args.TunGw = flag.String("tunGw", "240.0.0.1", "TUN interface gateway")
	args.TunMask = flag.String("tunMask", "255.255.255.0", "TUN interface netmask, it should be a prefix length (a number) for IPv6 address")
	args.TunDNS = flag.String("tunDNS", "1.1.1.1", "DNS resolvers for TUN interface (Windows Only)")

	// Proxy
	args.ProxyServer = flag.String("proxyServer", "", "Proxy server address")
	args.UdpTimeout = flag.Duration("udpTimeout", 30*time.Second, "UDP session timeout")
}

func main() {
	// Parse args
	flag.Parse()

	if *args.Version {
		fmt.Println("Go-tun2socks", version)
		fmt.Println(description)
		os.Exit(0)
	}

	// Initialization ops after parsing flags.
	for _, fn := range postFlagsInitFn {
		if fn != nil {
			fn()
		}
	}

	// Set log level.
	switch strings.ToLower(*args.LogLevel) {
	case "debug":
		log.SetLevel(log.DEBUG)
	case "info":
		log.SetLevel(log.INFO)
	case "warning":
		log.SetLevel(log.WARNING)
	case "error":
		log.SetLevel(log.ERROR)
	case "silent":
		log.SetLevel(log.SILENT)
	default:
		panic("unsupported logging level")
	}

	// Open the tun device.
	dnsServers := strings.Split(*args.TunDNS, ",")
	tunDev, err := tun.OpenTunDevice(*args.TunName, *args.TunAddr, *args.TunGw, *args.TunMask, dnsServers)
	if err != nil {
		log.Fatalf("failed to open tun device: %v", err)
	}

	// Setup TCP/IP stack.
	var lwipWriter = core.NewLWIPStack().(io.Writer)
	// Wrap a writer to delay ICMP packets.
	lwipWriter = filter.NewICMPFilter(lwipWriter).(io.Writer)

	// Register TCP and UDP handlers to handle accepted connections.
	proxyAddr, err := net.ResolveTCPAddr("tcp", *args.ProxyServer)
	if err != nil {
		log.Fatalf("invalid proxy server address: %v", err)
	}
	proxyHost := proxyAddr.IP.String()
	proxyPort := proxyAddr.Port
	core.RegisterTCPConnHandler(proxy.NewTCPHandler(proxyHost, proxyPort, fakeDNS, sessionStater))
	core.RegisterUDPConnHandler(proxy.NewUDPHandler(proxyHost, proxyPort, *args.UdpTimeout, fakeDNS, sessionStater))

	// Register an output callback to write packets output from lwip stack to tun
	// device, output function should be set before input any packets.
	core.RegisterOutputFn(func(data []byte) (int, error) {
		return tunDev.Write(data)
	})

	// Copy packets from tun device to lwip stack, it's the main loop.
	go func() {
		if _, err := io.CopyBuffer(lwipWriter, tunDev, make([]byte, MTU)); err != nil {
			log.Fatalf("copying data failed: %v", err)
		}
	}()

	log.Infof("Running Go-tun2socks")

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	<-osSignals

	// Stop fakeDNS server
	if fakeDNS != nil {
		fakeDNS.Stop()
	}

	// Stop session stater
	if sessionStater != nil {
		sessionStater.Stop()
	}
}
