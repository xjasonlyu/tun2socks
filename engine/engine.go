package engine

import (
	"errors"
	"net"

	"github.com/xjasonlyu/tun2socks/v2/component/dialer"
	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/option"
	_ "github.com/xjasonlyu/tun2socks/v2/dns"
	"github.com/xjasonlyu/tun2socks/v2/log"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/stats"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

var _engine = &engine{}

// Start starts the default engine up.
func Start() {
	if err := _engine.start(); err != nil {
		log.Fatalf("[ENGINE] failed to start: %v", err)
	}
}

// Stop shuts the default engine down.
func Stop() {
	if err := _engine.stop(); err != nil {
		log.Fatalf("[ENGINE] failed to stop: %v", err)
	}
}

// Insert loads *Key to the default engine.
func Insert(k *Key) {
	_engine.insert(k)
}

type Key struct {
	MTU        int    `yaml:"mtu"`
	Mark       int    `yaml:"fwmark"`
	UDPTimeout int    `yaml:"udp-timeout"`
	Proxy      string `yaml:"proxy"`
	Stats      string `yaml:"stats"`
	Token      string `yaml:"token"`
	Device     string `yaml:"device"`
	LogLevel   string `yaml:"loglevel"`
	Interface  string `yaml:"interface"`
}

type engine struct {
	*Key

	stack  *stack.Stack
	proxy  proxy.Proxy
	device device.Device
}

func (e *engine) start() error {
	if e.Key == nil {
		return errors.New("empty key")
	}

	for _, f := range []func() error{
		e.applyLogLevel,
		e.applyMark,
		e.applyInterface,
		e.applyStats,
		e.applyUDPTimeout,
		e.applyProxy,
		e.applyDevice,
		e.applyStack,
	} {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func (e *engine) stop() error {
	if e.device != nil {
		return e.device.Close()
	}
	e.stack.Close()
	e.stack.Wait()
	return nil
}

func (e *engine) insert(k *Key) {
	e.Key = k
}

func (e *engine) applyLogLevel() error {
	level, err := log.ParseLevel(e.LogLevel)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	return nil
}

func (e *engine) applyMark() error {
	if e.Mark != 0 {
		dialer.SetMark(e.Mark)
		log.Infof("[DIALER] set fwmark: %#x", e.Mark)
	}
	return nil
}

func (e *engine) applyInterface() error {
	if e.Interface != "" {
		if err := dialer.BindToInterface(e.Interface); err != nil {
			return err
		}
		log.Infof("[DIALER] bind to interface: %s", e.Interface)
	}
	return nil
}

func (e *engine) applyStats() error {
	if e.Stats != "" {
		addr, err := net.ResolveTCPAddr("tcp", e.Stats)
		if err != nil {
			return err
		}

		go func() {
			if err := stats.Start(addr.String(), e.Token); err != nil {
				log.Warnf("[STATS] failed to start: %v", err)
			}
		}()
		log.Infof("[STATS] serve at: http://%s", addr)
	}
	return nil
}

func (e *engine) applyUDPTimeout() error {
	if e.UDPTimeout > 0 {
		tunnel.SetUDPTimeout(e.UDPTimeout)
	}
	return nil
}

func (e *engine) applyProxy() (err error) {
	if e.Proxy == "" {
		return errors.New("empty proxy")
	}

	e.proxy, err = parseProxy(e.Proxy)
	proxy.SetDialer(e.proxy)
	return
}

func (e *engine) applyDevice() (err error) {
	if e.Device == "" {
		return errors.New("empty device")
	}

	e.device, err = parseDevice(e.Device, uint32(e.MTU))
	return
}

func (e *engine) applyStack() (err error) {
	defer func() {
		if err == nil {
			log.Infof(
				"[STACK] %s://%s <-> %s://%s",
				e.device.Type(), e.device.Name(),
				e.proxy.Proto(), e.proxy.Addr(),
			)
		}
	}()

	e.stack, err = core.CreateStackWithOptions(e.device, &fakeTunnel{}, option.WithDefault())
	return
}
