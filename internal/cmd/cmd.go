package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/xjasonlyu/clash/component/dialer"
	"github.com/xjasonlyu/tun2socks/internal/api"
	"github.com/xjasonlyu/tun2socks/internal/core"
	"github.com/xjasonlyu/tun2socks/internal/dns"
	"github.com/xjasonlyu/tun2socks/internal/proxy"
	"github.com/xjasonlyu/tun2socks/internal/tunnel"
	"github.com/xjasonlyu/tun2socks/pkg/log"
	"github.com/xjasonlyu/tun2socks/pkg/tun"
)

func bindToInterface(name string) {
	dialer.DialHook = dialer.DialerWithInterface(name)
	dialer.ListenPacketHook = dialer.ListenPacketWithInterface(name)
}

func printVersion(app *cli.App) {
	fmt.Printf("%s %s\n%s/%s, %s, %s\n",
		app.Name,
		app.Version,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
		app.Compiled.Format(time.RFC3339),
	)
}

func Main(c *cli.Context) error {
	if c.Bool("version") {
		printVersion(c.App)
		return nil
	}

	level, err := log.ParseLevel(c.String("loglevel"))
	if err != nil {
		return err
	}
	log.SetLevel(level)

	if c.IsSet("interface") {
		name := c.String("interface")
		bindToInterface(name)
		log.Infof("[DIALER] bind to interface: %s", name)
	}

	if c.IsSet("api") { /* initiate API */
		raw := c.String("api")
		if err := api.Start(raw, c.App); err != nil {
			return fmt.Errorf("start API server %s: %w", raw, err)
		}
		log.Infof("[API] listen and serve at: %s", raw)
	}

	if c.IsSet("dns") { /* initiate DNS */
		raw := c.String("dns")
		if err := dns.Start(raw, c.StringSlice("hosts")); err != nil {
			return fmt.Errorf("start DNS server %s: %w", raw, err)
		}
		log.Infof("[DNS] listen and serve at: %s", raw)
	}

	deviceURL := c.String("device")
	device, err := tun.Open(deviceURL)
	if err != nil {
		return fmt.Errorf("open device %s: %w", deviceURL, err)
	}
	defer device.Close()

	proxyURL := c.String("proxy")
	if err := proxy.Register(proxyURL); err != nil {
		return fmt.Errorf("register proxy %s: %w", proxyURL, err)
	}

	if _, err := core.NewDefaultStack(device, tunnel.Add, tunnel.AddPacket); err != nil {
		return fmt.Errorf("initiate stack: %w", err)
	}
	log.Infof("[STACK] %s --> %s", deviceURL, proxyURL)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	return nil
}
