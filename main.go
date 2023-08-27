package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
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
var privateIP string

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
	flag.StringVar(&privateIP, "private-ip", "10.10.10.10", "Private IP address for TUN device (Linux only)")
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

	if autoSetup && runtime.GOOS == "linux" {
		commands := []string{
			fmt.Sprintf("tuntap add mode tun dev %s", key.Device),
			fmt.Sprintf("addr add %s/24 dev %s", privateIP, key.Device),
			fmt.Sprintf("link set dev %s up", key.Device),
			fmt.Sprintf("route add default dev %s metric 1", key.Device),
		}
		for _, cmd := range commands {
			if err := exec.Command("ip", strings.Split(cmd, " ")...).Run(); err != nil {
				log.Fatalf("Failed to setup TUN device: %v on command: %s", err, cmd)
			}
		}

		defer func() {
			// Run shell command to delete TUN device
			commands = []string{
				fmt.Sprintf("link set dev %s down", key.Device),
				fmt.Sprintf("tuntap del mode tun dev %s", key.Device),
			}

			for _, cmd := range commands {
				if err := exec.Command("ip", strings.Split(cmd, " ")...).Run(); err != nil {
					log.Fatalf("Failed to delete TUN device: %v on command: %s", err, cmd)
				}
			}
		}()

	}

	engine.Insert(key)

	engine.Start()
	defer engine.Stop()

	// Use ifconfig to bring the TUN interface up and assign addresses for it.
	if autoSetup && runtime.GOOS == "darwin" {

		if err := exec.Command("ifconfig", key.Device, privateIP, privateIP, "up").Run(); err != nil {
			log.Fatalf("Failed to setup TUN device: %v", err)
		}
		routes := []string{
			"1.0.0.0/8",
			"2.0.0.0/7",
			"4.0.0.0/6",
			"8.0.0.0/5",
			"16.0.0.0/4",
			"32.0.0.0/3",
			"64.0.0.0/2",
			"128.0.0.0/1",
			"198.18.0.0/15",
		}
		for _, route := range routes {
			if err := exec.Command("route", "add", "-net", route, privateIP).Run(); err != nil {
				log.Fatalf("Failed to add route: %v", err)
			}
		}
	}
	if autoSetup && runtime.GOOS == "windows" {
		cmd := fmt.Sprintf(`interface ip set address name="%s" source=static addr=%s mask=255.255.255.0 gateway=none`, key.Device, privateIP)
		if err := exec.Command("netsh", strings.Split(cmd, " ")...).Run(); err != nil {
			log.Fatalf("Failed to setup TUN device: %v", err)
		}
		interfaceIndex, err := getInterfaceIndex(key.Device)
		if err != nil {
			log.Fatalf("Could not find interface index: %v", err)
		}
		// route add 0.0.0.0 mask 0.0.0.0 192.168.123.1 if <IF NUM> metric 5
		cmd = fmt.Sprintf(`add 0.0.0.0 mask 0.0.0.0 %s if %s metric 5`, privateIP, interfaceIndex)
		if err := exec.Command("route", strings.Split(cmd, " ")...).Run(); err != nil {
			log.Fatalf("Failed to setup TUN device: %v", err)
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

// Windows only
func getInterfaceIndex(deviceName string) (string, error) {
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("not windows")
	}
	cmd := exec.Command("netsh", "interface", "ipv4", "show", "interfaces")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running netsh command: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, deviceName) {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return fields[2], nil
			}
		}
	}

	return "", fmt.Errorf("interface not found")
}
