package engine

import (
	"errors"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/docker/go-units"
	"github.com/google/shlex"
	"github.com/vishvananda/netlink"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/option"
	"github.com/xjasonlyu/tun2socks/v2/dialer"
	"github.com/xjasonlyu/tun2socks/v2/log"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/restapi"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"
)

var (
	_engineMu sync.Mutex

	// _defaultKey holds default key for engine.
	_defaultKey *Key

	// _defaultProxy holds default proxy for engine.
	_defaultProxy proxy.Proxy

	// _defaultDevice holds default device for engine.
	_defaultDevice device.Device

	// _defaultStack holds default stack for engine.
	_defaultStack *stack.Stack
)

// Start starts default engine up.
func Start() {
	if err := start(); err != nil {
		log.Fatalf("[ENGINE] failed to start: %v", err)
	}
}

// Stop shuts down default engine.
func Stop() {
	if err := stop(); err != nil {
		log.Fatalf("[ENGINE] failed to stop: %v", err)
	}
}

// Insert loads *Key to default engine.
func Insert(k *Key) {
	_engineMu.Lock()
	_defaultKey = k
	_engineMu.Unlock()
}

// UpdateProxy updates the proxy configuration dynamically.
func UpdateProxy(proxyStr, user, pass string) error {
	_engineMu.Lock()
	defer _engineMu.Unlock()

	p, err := parseProxy(proxyStr)
	if err != nil {
		return err
	}

	// Update tunnel proxy
	tunnel.T().SetProxy(p)
	_defaultProxy = p

	if _defaultKey != nil {
		_defaultKey.Proxy = proxyStr
	}

	log.Infof("[PROXY] Updated proxy to: %s", proxyStr)
	return nil
}

func start() error {
	_engineMu.Lock()
	defer _engineMu.Unlock()

	if _defaultKey == nil {
		return errors.New("empty key")
	}

	for _, f := range []func(*Key) error{
		general,
		restAPI,
		netstack,
	} {
		if err := f(_defaultKey); err != nil {
			return err
		}
	}

	// Set engine running flag for Web API
	restapi.SetEngineRunning(true)

	return nil
}

func stop() error {
	_engineMu.Lock()
	defer _engineMu.Unlock()

	// Set engine stopped flag for Web API
	restapi.SetEngineRunning(false)

	if _defaultDevice != nil {
		_defaultDevice.Close()
	}
	if _defaultStack != nil {
		_defaultStack.Close()
		_defaultStack.Wait()
	}
	return nil
}

func execCommand(cmd string) error {
	parts, err := shlex.Split(cmd)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return errors.New("empty command")
	}
	_, err = exec.Command(parts[0], parts[1:]...).Output()
	return err
}

func general(k *Key) error {
	level, err := log.ParseLevel(k.LogLevel)
	if err != nil {
		return err
	}
	log.SetLogger(log.Must(log.NewLeveled(level)))

	if k.Interface != "" {
		iface, err := net.InterfaceByName(k.Interface)
		if err != nil {
			return err
		}
		dialer.DefaultDialer.InterfaceName.Store(iface.Name)
		dialer.DefaultDialer.InterfaceIndex.Store(int32(iface.Index))
		log.Infof("[DIALER] bind to interface: %s", k.Interface)
	}

	if k.Mark != 0 {
		dialer.DefaultDialer.RoutingMark.Store(int32(k.Mark))
		log.Infof("[DIALER] set fwmark: %#x", k.Mark)
	}

	if k.UDPTimeout > 0 {
		if k.UDPTimeout < time.Second {
			return errors.New("invalid udp timeout value")
		}
		tunnel.T().SetUDPTimeout(k.UDPTimeout)
	}
	return nil
}

func restAPI(k *Key) error {
	if k.RestAPI != "" {
		u, err := parseRestAPI(k.RestAPI)
		if err != nil {
			return err
		}
		host, token := u.Host, u.User.String()

		// Set initial proxy config from config file
		restapi.SetProxyConfig(k.Proxy)

		// Register callbacks
		restapi.UpdateProxyFunc = UpdateProxy

		restapi.SetStatsFunc(func() tcpip.Stats {
			_engineMu.Lock()
			defer _engineMu.Unlock()

			// default stack is not initialized.
			if _defaultStack == nil {
				return tcpip.Stats{}
			}
			return _defaultStack.Stats()
		})

		go func() {
			if err := restapi.Start(host, token); err != nil {
				log.Errorf("[RESTAPI] failed to start: %v", err)
			}
		}()
		log.Infof("[RESTAPI] serve at: %s", u)
	}
	return nil
}

func netstack(k *Key) (err error) {
	if k.Proxy == "" {
		log.Warnf("[STACK] No proxy configured. Setting dummy proxy. Please configure proxy via Web API.")
		k.Proxy = "socks5://127.0.0.1:0"
	}
	if k.Device == "" {
		return errors.New("empty device")
	}

	// Parse device name from URL format (e.g., "tun://tunsocks" -> "tunsocks")
	deviceName := k.Device
	if strings.Contains(k.Device, "://") {
		parts := strings.SplitN(k.Device, "://", 2)
		if len(parts) == 2 {
			deviceName = parts[1]
		}
	}

	// Delete existing device if present (to allow tun.Open() to create fresh device)
	if link, err := netlink.LinkByName(deviceName); err == nil {
		log.Infof("[TUN] Deleting existing device: %s", deviceName)
		if err := netlink.LinkDel(link); err != nil {
			log.Warnf("[TUN] Failed to delete existing device: %v", err)
		}
	}

	if k.TUNPreUp != "" {
		log.Infof("[TUN] pre-execute command: `%s`", k.TUNPreUp)
		if preUpErr := execCommand(k.TUNPreUp); preUpErr != nil {
			log.Errorf("[TUN] failed to pre-execute: %s: %v", k.TUNPreUp, preUpErr)
		}
	}

	defer func() {
		if k.TUNPostUp == "" || err != nil {
			return
		}
		log.Infof("[TUN] post-execute command: `%s`", k.TUNPostUp)
		if postUpErr := execCommand(k.TUNPostUp); postUpErr != nil {
			log.Errorf("[TUN] failed to post-execute: %s: %v", k.TUNPostUp, postUpErr)
		}
	}()

	multicastGroups, err := parseMulticastGroups(k.MulticastGroups)
	if err != nil {
		return err
	}

	if _defaultProxy, err = parseProxy(k.Proxy); err != nil {
		return err
	}
	tunnel.T().SetProxy(_defaultProxy)

	if _defaultDevice, err = parseDevice(k.Device, uint32(k.MTU)); err != nil {
		return err
	}

	// Add IP address and set device UP
	if deviceName != "" {
		link, linkErr := netlink.LinkByName(deviceName)
		if linkErr == nil {
			// Add default IP address (198.18.0.1/15)
			addr, addrErr := netlink.ParseAddr("198.18.0.1/15")
			if addrErr != nil {
				log.Warnf("[TUN] Failed to parse IP address: %v", addrErr)
			} else {
				if addErr := netlink.AddrAdd(link, addr); addErr != nil {
					log.Warnf("[TUN] Failed to add IP address: %v", addErr)
				} else {
					log.Infof("[TUN] Added IP address: 198.18.0.1/15 to %s", deviceName)
				}
			}

			// Set device UP
			if upErr := netlink.LinkSetUp(link); upErr != nil {
				log.Warnf("[TUN] Failed to set device UP: %v", upErr)
			} else {
				log.Infof("[TUN] Device %s is now UP", deviceName)
			}
		}
	}

	var opts []option.Option
	if k.TCPModerateReceiveBuffer {
		opts = append(opts, option.WithTCPModerateReceiveBuffer(true))
	}

	if k.TCPSendBufferSize != "" {
		size, err := units.RAMInBytes(k.TCPSendBufferSize)
		if err != nil {
			return err
		}
		opts = append(opts, option.WithTCPSendBufferSize(int(size)))
	}

	if k.TCPReceiveBufferSize != "" {
		size, err := units.RAMInBytes(k.TCPReceiveBufferSize)
		if err != nil {
			return err
		}
		opts = append(opts, option.WithTCPReceiveBufferSize(int(size)))
	}

	if _defaultStack, err = core.CreateStack(&core.Config{
		LinkEndpoint:     _defaultDevice,
		TransportHandler: tunnel.T(),
		MulticastGroups:  multicastGroups,
		Options:          opts,
	}); err != nil {
		return err
	}

	log.Infof("[STACK] %s <-> %s", k.Device, k.Proxy)
	return nil
}
