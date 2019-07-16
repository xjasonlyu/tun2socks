package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/xjasonlyu/tun2socks/common/dns"
	"github.com/xjasonlyu/tun2socks/common/log"
	_ "github.com/xjasonlyu/tun2socks/common/log/simple" // Register a simple logger.
	"github.com/xjasonlyu/tun2socks/common/stats"
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/filter"
	"github.com/xjasonlyu/tun2socks/tun"
)

var version = "undefined"

var handlerCreator = make(map[string]func(), 0)

func registerHandlerCreator(name string, creator func()) {
	handlerCreator[name] = creator
}

var postFlagsInitFn = make([]func(), 0)

func addPostFlagsInitFn(fn func()) {
	postFlagsInitFn = append(postFlagsInitFn, fn)
}

type CmdArgs struct {
	Version              *bool
	TunName              *string
	TunAddr              *string
	TunGw                *string
	TunMask              *string
	TunDns               *string
	ProxyType            *string
	VConfig              *string
	SniffingType         *string
	ProxyServer          *string
	ProxyHost            *string
	ProxyPort            *uint16
	ProxyCipher          *string
	ProxyPassword        *string
	DelayICMP            *int
	UdpTimeout           *time.Duration
	DisableDnsCache      *bool
	DnsFallback          *bool
	LogLevel             *string
	EnableFakeDns        *bool
	FakeIPRange          *string
	FakeDnsAddr          *string
	FakeDnsHosts         *string
	ExceptionApps        *string
	ExceptionSendThrough *string
	Stats                *bool
	StatsAddr            *string
}

type cmdFlag uint

const (
	fProxyServer cmdFlag = iota
	fUdpTimeout
	fStats
)

var flagCreators = map[cmdFlag]func(){
	fProxyServer: func() {
		if args.ProxyServer == nil {
			args.ProxyServer = flag.String("proxyServer", "1.2.3.4:1087", "Proxy server address")
		}
	},
	fUdpTimeout: func() {
		if args.UdpTimeout == nil {
			args.UdpTimeout = flag.Duration("udpTimeout", 1*time.Minute, "UDP session timeout")
		}
	},
	fStats: func() {
		if args.Stats == nil {
			args.Stats = flag.Bool("stats", false, "Enable statistics")
		}
	},
}

func (a *CmdArgs) addFlag(f cmdFlag) {
	if fn, found := flagCreators[f]; found && fn != nil {
		fn()
	} else {
		log.Fatalf("unsupported flag")
	}
}

var args = new(CmdArgs)

var lwipWriter io.Writer

var dnsCache dns.DnsCache

var fakeDns dns.FakeDns

var sessionStater stats.SessionStater

const (
	MTU = 1500
)

func main() {
	args.Version = flag.Bool("version", false, "Print version")
	args.TunName = flag.String("tunName", "tun0", "TUN interface name")
	args.TunAddr = flag.String("tunAddr", "240.0.0.2", "TUN interface address")
	args.TunGw = flag.String("tunGw", "240.0.0.1", "TUN interface gateway")
	args.TunMask = flag.String("tunMask", "255.255.255.0", "TUN interface netmask, it should be a prefixlen (a number) for IPv6 address")
	args.TunDns = flag.String("tunDns", "8.8.8.8,8.8.4.4", "DNS resolvers for TUN interface (only need on Windows)")
	args.ProxyType = flag.String("proxyType", "socks", "Proxy handler type")
	args.DelayICMP = flag.Int("delayICMP", 10, "Delay ICMP packets for a short period of time, in milliseconds")
	args.LogLevel = flag.String("loglevel", "info", "Logging level. (debug, info, warn, error, none)")

	flag.Parse()

	if *args.Version {
		fmt.Println(version)
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
	case "warn":
		log.SetLevel(log.WARN)
	case "error":
		log.SetLevel(log.ERROR)
	case "none":
		log.SetLevel(log.NONE)
	default:
		panic("unsupported logging level")
	}

	// Open the tun device.
	dnsServers := strings.Split(*args.TunDns, ",")
	tunDev, err := tun.OpenTunDevice(*args.TunName, *args.TunAddr, *args.TunGw, *args.TunMask, dnsServers)
	if err != nil {
		log.Fatalf("failed to open tun device: %v", err)
	}

	// Setup TCP/IP stack.
	lwipWriter = core.NewLWIPStack().(io.Writer)

	// Wrap a writer to delay ICMP packets if delay time is not zero.
	if *args.DelayICMP > 0 {
		log.Infof("ICMP packets will be delayed for %dms", *args.DelayICMP)
		lwipWriter = filter.NewICMPFilter(lwipWriter, *args.DelayICMP).(io.Writer)
	}

	// Register TCP and UDP handlers to handle accepted connections.
	if creator, found := handlerCreator[*args.ProxyType]; found {
		creator()
	} else {
		log.Fatalf("unsupported proxy type")
	}

	if args.DnsFallback != nil && *args.DnsFallback {
		// Override the UDP handler with a DNS-over-TCP (fallback) UDP handler.
		if creator, found := handlerCreator["dnsfallback"]; found {
			creator()
		} else {
			log.Fatalf("DNS fallback connection handler not found, build with `dnsfallback` tag")
		}
	}

	// Register an output callback to write packets output from lwip stack to tun
	// device, output function should be set before input any packets.
	core.RegisterOutputFn(func(data []byte) (int, error) {
		return tunDev.Write(data)
	})

	// Copy packets from tun device to lwip stack, it's the main loop.
	go func() {
		_, err := io.CopyBuffer(lwipWriter, tunDev, make([]byte, MTU))
		if err != nil {
			log.Fatalf("copying data failed: %v", err)
		}
	}()

	log.Infof("Running tun2socks")

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	<-osSignals

	if sessionStater != nil {
		_ = sessionStater.Stop()
	}
}
