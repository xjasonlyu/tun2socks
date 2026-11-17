package proxy

import (
	"errors"
	"time"
)

// TCPTimeouts configures opt-in HTTP CONNECT/SOCKS4/SOCKS5 handshake waiting.
// Proxy connection keeps the fixed 5s timeout. The handshake timer starts only
// after connection succeeds. Zero selects 21s; negatives are invalid.
// These settings do not change Proxy.DialContext's legacy behavior.
type TCPTimeouts struct {
	ProxyHandshakeTimeout time.Duration
}

func (c TCPTimeouts) WithDefaults() (TCPTimeouts, error) {
	if c.ProxyHandshakeTimeout < 0 {
		return c, errors.New("proxy handshake timeout must not be negative")
	}
	if c.ProxyHandshakeTimeout == 0 {
		c.ProxyHandshakeTimeout = 21 * time.Second
	}
	return c, nil
}
