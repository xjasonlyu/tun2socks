package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	C "github.com/xjasonlyu/tun2socks/constant"
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/filter"
	"github.com/xjasonlyu/tun2socks/log"
	"github.com/xjasonlyu/tun2socks/proxy"
	"github.com/xjasonlyu/tun2socks/tun"

	D "github.com/xjasonlyu/tun2socks/component/fakedns"
	S "github.com/xjasonlyu/tun2socks/component/session"
)

const MTU = 1500

var (
	args = new(CmdArgs)

	// Modules init func
	registeredInitFn []func()

	fakeDNS D.FakeDNS
	monitor S.Monitor
)

type CmdArgs struct {
	// Main
	Version    *bool
	TunName    *string
	TunAddr    *string
	TunGw      *string
	TunMask    *string
	TunDNS     *string
	TunPersist *bool
	LogLevel   *string

	// Proxy
	ProxyServer *string
	UdpTimeout  *time.Duration

	// FakeDNS
	EnableFakeDNS *bool
	FakeDNSAddr   *string
	Hosts         *string
	HijackDNS     *string
	BackendDNS    *string

	// Session Monitor
	EnableMonitor *bool
	MonitorAddr   *string
}

func registerInitFn(fn func()) {
	registeredInitFn = append(registeredInitFn, fn)
}

func init() {
	// Main
	args.Version = flag.Bool("version", false, "Show current version of tun2socks")
	args.LogLevel = flag.String("loglevel", "info", "Logging level [info, warning, error, debug, silent]")
	args.TunName = flag.String("tunName", "utun0", "TUN interface name")
	args.TunAddr = flag.String("tunAddr", "240.0.0.2", "TUN interface address")
	args.TunGw = flag.String("tunGw", "240.0.0.1", "TUN interface gateway")
	args.TunMask = flag.String("tunMask", "255.255.255.0", "TUN interface netmask")
	args.TunDNS = flag.String("tunDNS", "8.8.8.8,8.8.4.4", "DNS resolvers for TUN interface (Windows Only)")
	args.TunPersist = flag.Bool("tunPersist", false, "Persist TUN interface after the program exits or the last open file descriptor is closed (Linux only)")

	// Proxy
	args.ProxyServer = flag.String("proxyServer", "", "Proxy server address")
	args.UdpTimeout = flag.Duration("udpTimeout", 30*time.Second, "UDP session timeout")
}

func showVersion() {
	version := strings.Split(C.Version[1:], "-")
	fmt.Printf("Go-tun2socks %s (%s)\n", version[0], version[1])
	fmt.Printf("%s/%s, %s, %s\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), version[2])
}

func main() {
	// Parse arguments
	flag.Parse()

	if *args.Version {
		showVersion()
		os.Exit(0)
	}

	// Set log level
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

	// Initialization modules
	for _, fn := range registeredInitFn {
		if fn != nil {
			fn()
		}
	}

	// Resolve proxy address
	proxyAddr, err := net.ResolveTCPAddr("tcp", *args.ProxyServer)
	if err != nil {
		log.Fatalf("invalid proxy server address: %v", err)
	}
	proxyHost := proxyAddr.IP.String()
	proxyPort := proxyAddr.Port

	// Open the tun device
	dnsServers := strings.Split(*args.TunDNS, ",")
	tunDev, err := tun.OpenTunDevice(*args.TunName, *args.TunAddr, *args.TunGw, *args.TunMask, dnsServers, *args.TunPersist)
	if err != nil {
		log.Fatalf("failed to open tun device: %v", err)
	}

	// Setup TCP/IP stack
	var lwipWriter = core.NewLWIPStack().(io.Writer)
	// Wrap a writer to delay ICMP packets
	lwipWriter = filter.NewICMPFilter(lwipWriter).(io.Writer)

	// Register modules to proxy
	proxy.RegisterMonitor(monitor)
	proxy.RegisterFakeDNS(fakeDNS, *args.HijackDNS)
	// Register TCP and UDP handlers to handle accepted connections.
	core.RegisterTCPConnHandler(proxy.NewTCPHandler(proxyHost, proxyPort))
	core.RegisterUDPConnHandler(proxy.NewUDPHandler(proxyHost, proxyPort, *args.UdpTimeout))

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

	// Stop fakeDNS
	if fakeDNS != nil {
		fakeDNS.Stop()
	}

	// Stop session monitor
	if monitor != nil {
		monitor.Stop()
	}
}
