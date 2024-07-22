package proxy

import (
	"errors"
	"net/url"
	"sync"

	"go.uber.org/atomic"
)

// ErrProtocol indicates that parsing encountered an unknown protocol.
var ErrProtocol = errors.New("proxy: unknown protocol")

// A proxy holds a proxy's protocol and how to parse it.
type protocol struct {
	name  string
	parse func(*url.URL) (Proxy, error)
}

// Protocols is the list of registered proxy protocols.
var (
	protocolsMu     sync.Mutex
	atomicProtocols atomic.Value
)

// RegisterProtocol registers a proxy protocol for use by [Parse].
// Name is the name of the proxy protocol, like "http" or "socks5".
// [Parse] is the function that parses the proxy url.
func RegisterProtocol(name string, parse func(*url.URL) (Proxy, error)) {
	protocolsMu.Lock()
	formats, _ := atomicProtocols.Load().([]protocol)
	atomicProtocols.Store(append(formats, protocol{name, parse}))
	protocolsMu.Unlock()
}

// pick determines the protocol by the given name.
func pick(name string) protocol {
	protocols, _ := atomicProtocols.Load().([]protocol)
	for _, p := range protocols {
		if p.name == name {
			return p
		}
	}
	return protocol{}
}

// Parse parses a proxy url that has been encoded in a registered format.
// The string returned is the format name used during format registration.
// Format registration is typically done by an init function in the codec-
// specific package.
func Parse(proxyURL *url.URL) (Proxy, error) {
	if proxyURL == nil {
		return nil, errors.New("proxy: nil url")
	}
	if proxyURL.Scheme == "" {
		return nil, errors.New("proxy: protocol not specified")
	}
	p := pick(proxyURL.Scheme)
	if p.parse == nil {
		return nil, ErrProtocol
	}
	return p.parse(proxyURL)
}

// ParseFromURL parses a
func ParseFromURL(proxy string) (Proxy, error) {
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, err
	}
	return Parse(proxyURL)
}
