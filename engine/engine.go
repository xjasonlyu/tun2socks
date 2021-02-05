package engine

import (
	"github.com/xjasonlyu/tun2socks/component/dialer"
	"github.com/xjasonlyu/tun2socks/device"
	"github.com/xjasonlyu/tun2socks/log"
	"github.com/xjasonlyu/tun2socks/proxy"
	"github.com/xjasonlyu/tun2socks/stack"
	"github.com/xjasonlyu/tun2socks/stats"
)

type Engine struct {
	mtu       uint32
	iface     string
	secret    string
	stats     string
	logLevel  string
	rawProxy  string
	rawDevice string

	device device.Device
}

func New(opts ...Option) *Engine {
	e := &Engine{}

	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Engine) Start() error {
	for _, set := range []func() error{
		e.setLogLevel,
		e.setInterface,
		e.setStats,
		e.setProxy,
		e.setDevice,
		e.setStack,
	} {
		if err := set(); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) Stop() {
	if e.device != nil {
		if err := e.device.Close(); err != nil {
			log.Fatalf("%v", err)
		}
	}
}

func (e *Engine) setLogLevel() error {
	level, err := log.ParseLevel(e.logLevel)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	return nil
}

func (e *Engine) setInterface() error {
	if e.iface != "" {
		if err := dialer.BindToInterface(e.iface); err != nil {
			return err
		}
		log.Infof("[BOUND] bind to interface: %s", e.iface)
	}
	return nil
}

func (e *Engine) setStats() error {
	if e.stats != "" {
		go func() {
			_ = stats.Start(e.stats, e.secret)
		}()
		log.Infof("[STATS] listen and serve at: http://%s", e.stats)
	}
	return nil
}

func (e *Engine) setProxy() error {
	d, err := parseProxy(e.rawProxy)
	if err != nil {
		return err
	}
	proxy.SetDialer(d)
	return nil
}

func (e *Engine) setDevice() error {
	d, err := parseDevice(e.rawDevice, e.mtu)
	if err != nil {
		return err
	}
	e.device = d
	return nil
}

func (e *Engine) setStack() error {
	handler := &fakeTunnel{}
	if _, err := stack.New(e.device, handler, stack.WithDefault()); err != nil {
		return err
	}
	log.Infof("[STACK] %s <-> %s", e.rawDevice, e.rawProxy)
	return nil
}
