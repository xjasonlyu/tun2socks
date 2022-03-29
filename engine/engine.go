package engine

import (
	"errors"

	"github.com/xjasonlyu/tun2socks/v2/component/dialer"
	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	_ "github.com/xjasonlyu/tun2socks/v2/dns"
	"github.com/xjasonlyu/tun2socks/v2/log"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/restapi"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"

	"gvisor.dev/gvisor/pkg/tcpip"
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
	RestAPI    string `yaml:"restapi"`
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
		e.applyDialer,
		e.applyRestAPI,
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

func (e *engine) stop() (err error) {
	if e.device != nil {
		err = e.device.Close()
	}
	if e.stack != nil {
		e.stack.Close()
		e.stack.Wait()
	}
	return err
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

func (e *engine) applyDialer() error {
	if e.Interface != "" {
		dialer.DefaultInterfaceName.Store(e.Interface)
		log.Infof("[DIALER] bind to interface: %s", e.Interface)
	}
	if e.Mark != 0 {
		dialer.DefaultRoutingMark.Store(int32(e.Mark))
		log.Infof("[DIALER] set fwmark: %#x", e.Mark)
	}
	return nil
}

func (e *engine) applyRestAPI() error {
	if e.RestAPI != "" {
		u, err := parseRestAPI(e.RestAPI)
		if err != nil {
			return err
		}
		host, token := u.Host, u.User.String()

		go func() {
			if err := restapi.Start(host, token); err != nil {
				log.Warnf("[RESTAPI] failed to start: %v", err)
			}
		}()
		log.Infof("[RESTAPI] serve at: %s", u)
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

	e.stack, err = core.CreateStack(&core.Config{
		LinkEndpoint:     e.device,
		TransportHandler: &fakeTunnel{},
		ErrorFunc: func(err tcpip.Error) {
			log.Warnf("[STACK] %s", err)
		},
	})
	return
}
