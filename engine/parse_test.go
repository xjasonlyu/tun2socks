package engine

import (
	"strings"
	"testing"

	"github.com/xjasonlyu/tun2socks/v2/proxy"
)

// TestBuildProxySplit guards against regression of the netstack()
// guard that used to reject tcp-proxy/udp-proxy configurations when
// --proxy was empty. buildProxy should accept that pairing and return
// a *proxy.Split.
func TestBuildProxySplit(t *testing.T) {
	p, err := buildProxy(&Key{TCPProxy: "direct://", UDPProxy: "direct://"})
	if err != nil {
		t.Fatalf("buildProxy split: %v", err)
	}
	if _, ok := p.(*proxy.Split); !ok {
		t.Fatalf("buildProxy split: got %T, want *proxy.Split", p)
	}
}

func TestBuildProxyEmpty(t *testing.T) {
	_, err := buildProxy(&Key{})
	if err == nil {
		t.Fatal("buildProxy empty: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no proxy configured") {
		t.Fatalf("buildProxy empty: unexpected error: %v", err)
	}
}

func TestBuildProxyBaseOnly(t *testing.T) {
	p, err := buildProxy(&Key{Proxy: "direct://"})
	if err != nil {
		t.Fatalf("buildProxy base-only: %v", err)
	}
	if _, ok := p.(*proxy.Split); ok {
		t.Fatalf("buildProxy base-only: got *proxy.Split, want direct proxy")
	}
}

func TestBuildProxyHalfSplit(t *testing.T) {
	// udp-proxy set, tcp-proxy empty, no base -> error
	_, err := buildProxy(&Key{UDPProxy: "direct://"})
	if err == nil {
		t.Fatal("buildProxy half-split: expected error, got nil")
	}
}
