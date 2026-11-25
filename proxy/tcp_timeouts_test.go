package proxy

import (
	"testing"
	"time"
)

func TestTCPTimeoutDefaultsAndValidation(t *testing.T) {
	cfg, err := (TCPTimeouts{}).WithDefaults()
	if err != nil || cfg.ProxyHandshakeTimeout != 21*time.Second {
		t.Fatal(cfg, err)
	}
	for _, cfg := range []TCPTimeouts{{ProxyHandshakeTimeout: -1}} {
		if _, err := cfg.WithDefaults(); err == nil {
			t.Fatal("accepted negative timeout")
		}
	}
	for _, cfg := range []TCPTimeouts{{ProxyHandshakeTimeout: time.Second}} {
		got, err := cfg.WithDefaults()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.ProxyHandshakeTimeout != 0 && got.ProxyHandshakeTimeout != cfg.ProxyHandshakeTimeout {
			t.Fatal(got)
		}
	}
}
