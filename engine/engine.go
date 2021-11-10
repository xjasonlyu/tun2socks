package engine

import (
	"errors"
	"os"

	"github.com/xjasonlyu/tun2socks/component/dialer"
	"github.com/xjasonlyu/tun2socks/core/device"
	"github.com/xjasonlyu/tun2socks/core/stack"
	"github.com/xjasonlyu/tun2socks/log"
	"github.com/xjasonlyu/tun2socks/proxy"
	"github.com/xjasonlyu/tun2socks/stats"
	"github.com/xjasonlyu/tun2socks/tunnel"

	"gopkg.in/yaml.v3"
)

var _engine = &engine{}

// Start starts the default engine up.
func Start() error {
	return _engine.start()
}

// Stop shuts the default engine down.
func Stop() error {
	return _engine.stop()
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
	Config     string `yaml:"-"`
	Version    bool   `yaml:"-"`
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

	if e.Version {
		showVersion()
		os.Exit(0)
	}

	for _, f := range []func() error{
		e.setConfig,
		e.setLogLevel,
		e.setMark,
		e.setInterface,
		e.setStats,
		e.setUDPTimeout,
		e.setProxy,
		e.setDevice,
		e.setStack,
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
	return nil
}

func (e *engine) insert(k *Key) {
	e.Key = k
}

func (e *engine) setConfig() error {
	if e.Config == "" {
		return nil
	}

	data, err := os.ReadFile(e.Config)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, e.Key)
}

func (e *engine) setLogLevel() error {
	level, err := log.ParseLevel(e.LogLevel)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	return nil
}

func (e *engine) setMark() error {
	if e.Mark != 0 {
		dialer.SetMark(e.Mark)
		log.Infof("[DIALER] set fwmark: %#x", e.Mark)
	}
	return nil
}

func (e *engine) setInterface() error {
	if e.Interface != "" {
		if err := dialer.BindToInterface(e.Interface); err != nil {
			return err
		}
		log.Infof("[DIALER] use interface: %s", e.Interface)
	}
	return nil
}

func (e *engine) setStats() error {
	if e.Stats != "" {
		go func() {
			_ = stats.Start(e.Stats, e.Token)
		}()
		log.Infof("[STATS] serve at: http://%s", e.Stats)
	}
	return nil
}

func (e *engine) setUDPTimeout() error {
	if e.UDPTimeout > 0 {
		tunnel.SetUDPTimeout(e.UDPTimeout)
	}
	return nil
}

func (e *engine) setProxy() (err error) {
	if e.Proxy == "" {
		return errors.New("empty proxy")
	}

	e.proxy, err = parseProxy(e.Proxy)
	proxy.SetDialer(e.proxy)
	return
}

func (e *engine) setDevice() (err error) {
	if e.Device == "" {
		return errors.New("empty device")
	}

	e.device, err = parseDevice(e.Device, uint32(e.MTU))
	return
}

func (e *engine) setStack() (err error) {
	defer func() {
		if err == nil {
			log.Infof(
				"[STACK] %s://%s <-> %s://%s",
				e.device.Type(), e.device.Name(),
				e.proxy.Proto(), e.proxy.Addr(),
			)
		}
	}()

	e.stack, err = stack.New(e.device, &fakeTunnel{}, stack.WithDefault())
	return
}
