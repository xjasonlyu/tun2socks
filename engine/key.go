package engine

import "time"

type Key struct {
	MTU                      int           `yaml:"mtu"`
	Mark                     int           `yaml:"fwmark"`
	Proxy                    string        `yaml:"proxy"`
	RestAPI                  string        `yaml:"restapi"`
	Device                   string        `yaml:"device"`
	LogLevel                 string        `yaml:"loglevel"`
	Interface                string        `yaml:"interface"`
	TCPModerateReceiveBuffer bool          `yaml:"tcp-moderate-receive-buffer"`
	TCPSendBufferSize        string        `yaml:"tcp-send-buffer-size"`
	TCPReceiveBufferSize     string        `yaml:"tcp-receive-buffer-size"`
	MulticastGroups          []string      `yaml:"multicast-groups"`
	TUNPreUp                 string        `yaml:"tun-pre-up"`
	TUNPostUp                string        `yaml:"tun-post-up"`
	UDPTimeout               time.Duration `yaml:"udp-timeout"`
	RejectIPv6               bool          `yaml:"reject-ipv6"`
	// Deferred handshake is opt-in; zero cache settings use core defaults.
	TCPDeferredHandshake     bool          `yaml:"tcp-deferred-handshake"`
	TCPProxyHandshakeTimeout time.Duration `yaml:"tcp-proxy-handshake-timeout"`
	TCPFailureCacheTTL       time.Duration `yaml:"tcp-failure-cache-ttl"`
	TCPFailureCacheCapacity  int           `yaml:"tcp-failure-cache-capacity"`
}
