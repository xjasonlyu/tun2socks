package proxy

import (
	"errors"
	"fmt"
	"net/url"
	"sync"

	"go.uber.org/atomic"
)

// ErrProtocol indicates that parsing encountered an unknown protocol.
var ErrProtocol = errors.New("proxy: unknown protocol")

// A protocol holds a proxy protocol's name and how to parse it.
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

// Parse parses proxy *url.URL that holds the proxy info into Proxy.
// Protocol registration is typically done by an init function in the
// proxy-specific package.
func Parse(proxyURL *url.URL) (Proxy, error) {
	if proxyURL == nil {
		return nil, errors.New("proxy: nil url")
	}
	if proxyURL.Scheme == "" {
		return nil, errors.New("proxy: protocol not specified")
	}
	p := pick(proxyURL.Scheme)
	if p.parse == nil {
		return nil, fmt.Errorf("%w: %s", ErrProtocol, proxyURL.Scheme)
	}
	return p.parse(proxyURL)
}

// ParseFromURL parses url string that holds the proxy info into Proxy.
func ParseFromURL(proxy string) (Proxy, error) {
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, err
	}
	return Parse(proxyURL)
}
