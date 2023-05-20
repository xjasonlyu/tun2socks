package engine

import (
	"errors"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/docker/go-units"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/option"
	"github.com/xjasonlyu/tun2socks/v2/dialer"
	"github.com/xjasonlyu/tun2socks/v2/engine/mirror"
	"github.com/xjasonlyu/tun2socks/v2/log"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/restapi"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"
)

var (
	_engineMu sync.Mutex

	// _defaultKey holds the default key for the engine.
	_defaultKey *Key

	// _defaultProxy holds the default proxy for the engine.
	_defaultProxy proxy.Proxy

	// _defaultDevice holds the default device for the engine.
	_defaultDevice device.Device

	// _defaultStack holds the default stack for the engine.
	_defaultStack *stack.Stack
)

// Start starts the default engine up.
func Start() {
	if err := start(); err != nil {
		log.Fatalf("[ENGINE] failed to start: %v", err)
	}
}

// Stop shuts the default engine down.
func Stop() {
	if err := stop(); err != nil {
		log.Fatalf("[ENGINE] failed to stop: %v", err)
	}
}

// Insert loads *Key to the default engine.
func Insert(k *Key) {
	_engineMu.Lock()
	_defaultKey = k
	_engineMu.Unlock()
}

func start() error {
	_engineMu.Lock()
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
	_engineMu.Unlock()
	return nil
}

func stop() (err error) {
	_engineMu.Lock()
	if _defaultDevice != nil {
		err = _defaultDevice.Close()
	}
	if _defaultStack != nil {
		_defaultStack.Close()
		_defaultStack.Wait()
	}
	_engineMu.Unlock()
	return err
}

func execCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return errors.New("empty command")
	}
	_, err := exec.Command(parts[0], parts[1:]...).Output()
	return err
}

func general(k *Key) error {
	level, err := log.ParseLevel(k.LogLevel)
	if err != nil {
		return err
	}
	log.SetLevel(level)

	if k.Interface != "" {
		iface, err := net.InterfaceByName(k.Interface)
		if err != nil {
			return err
		}
		dialer.DefaultInterfaceName.Store(iface.Name)
		dialer.DefaultInterfaceIndex.Store(int32(iface.Index))
		log.Infof("[DIALER] bind to interface: %s", k.Interface)
	}

	if k.Mark != 0 {
		dialer.DefaultRoutingMark.Store(int32(k.Mark))
		log.Infof("[DIALER] set fwmark: %#x", k.Mark)
	}

	if k.UDPTimeout > 0 {
		if k.UDPTimeout < time.Second {
			return errors.New("invalid udp timeout value")
		}
		tunnel.SetUDPTimeout(k.UDPTimeout)
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
				log.Warnf("[RESTAPI] failed to start: %v", err)
			}
		}()
		log.Infof("[RESTAPI] serve at: %s", u)
	}
	return nil
}

func netstack(k *Key) (err error) {
	if k.Proxy == "" {
		return errors.New("empty proxy")
	}
	if k.Device == "" {
		return errors.New("empty device")
	}

	if k.TUNPreUp != "" {
		if preUpErr := execCommand(k.TUNPreUp); preUpErr != nil {
			log.Warnf("[TUN] failed to pre-execute: %s: %v", k.TUNPreUp, preUpErr)
		}
	}

	defer func() {
		if k.TUNPostUp == "" || err != nil {
			return
		}
		if postUpErr := execCommand(k.TUNPostUp); postUpErr != nil {
			log.Warnf("[TUN] failed to post-execute: %s: %v", k.TUNPostUp, postUpErr)
		}
	}()

	if _defaultProxy, err = parseProxy(k.Proxy); err != nil {
		return
	}
	proxy.SetDialer(_defaultProxy)

	if _defaultDevice, err = parseDevice(k.Device, uint32(k.MTU)); err != nil {
		return
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
		TransportHandler: &mirror.Tunnel{},
		Options:          opts,
	}); err != nil {
		return
	}

	log.Infof(
		"[STACK] %s://%s <-> %s://%s",
		_defaultDevice.Type(), _defaultDevice.Name(),
		_defaultProxy.Proto(), _defaultProxy.Addr(),
	)
	return nil
}
