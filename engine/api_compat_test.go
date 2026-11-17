package engine

import (
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	_ func() error = Start
	_ func() error = Stop
	_ func(*Key)   = Insert
	_              = Key{Device: "tun://test", Proxy: "socks5://127.0.0.1:1080"}
)

func TestStartRejectsMissingKey(t *testing.T) {
	// No device/route operations occur in this test.
	_engineMu.Lock()
	previous := _defaultKey
	_defaultKey = nil
	_engineMu.Unlock()
	defer func() { _engineMu.Lock(); _defaultKey = previous; _engineMu.Unlock() }()
	if err := Start(); err == nil {
		t.Fatal("start without a key succeeded")
	}
}

func TestDeferredKeyYAMLAndValidation(t *testing.T) {
	var key Key
	if err := yaml.Unmarshal([]byte("device: tun://test\n"), &key); err != nil {
		t.Fatal(err)
	}
	if key.RejectIPv6 || key.TCPDeferredHandshake || key.TCPFailureCacheTTL != 0 || key.TCPFailureCacheCapacity != 0 || key.TCPProxyHandshakeTimeout != 0 {
		t.Fatal("legacy config enabled new behavior")
	}
	if err := yaml.Unmarshal([]byte("tcp-deferred-handshake: true\ntcp-failure-cache-ttl: 12s\ntcp-failure-cache-capacity: 128\n"), &key); err != nil {
		t.Fatal(err)
	}
	if !key.TCPDeferredHandshake || key.TCPFailureCacheTTL != 12*time.Second || key.TCPFailureCacheCapacity != 128 {
		t.Fatalf("config not decoded: %+v", key)
	}
	previous := _defaultKey
	defer func() { _defaultKey = previous }()
	for _, invalid := range []Key{{TCPFailureCacheTTL: -time.Second}, {TCPFailureCacheCapacity: -1}, {TCPProxyHandshakeTimeout: -time.Second}} {
		Insert(&invalid)
		if err := Start(); err == nil || !strings.Contains(err.Error(), "must not be negative") {
			t.Fatalf("invalid cache config was not rejected before setup: %v", err)
		}
	}
}

func TestSeparateTCPTimeoutConfiguration(t *testing.T) {
	var key Key
	if err := yaml.Unmarshal([]byte("tcp-proxy-handshake-timeout: 45s\n"), &key); err != nil {
		t.Fatal(err)
	}
	if key.TCPDeferredHandshake {
		t.Fatal("timeouts must not enable deferred mode")
	}
	got, err := tcpTimeouts(&key).WithDefaults()
	if err != nil || got.ProxyHandshakeTimeout != 45*time.Second {
		t.Fatal(got, err)
	}
	got, err = tcpTimeouts(&Key{}).WithDefaults()
	if err != nil || got.ProxyHandshakeTimeout != 21*time.Second {
		t.Fatal(got, err)
	}
}

func TestRejectIPv6YAMLIndependentOfDeferredTCP(t *testing.T) {
	var key Key
	if err := yaml.Unmarshal([]byte("reject-ipv6: true\n"), &key); err != nil {
		t.Fatal(err)
	}
	if !key.RejectIPv6 || key.TCPDeferredHandshake {
		t.Fatal("rejection must not enable deferred handshakes")
	}
}

func TestRemovedSOCKSTimeoutKeyIsNotAnAlias(t *testing.T) {
	const old = "tcp-socks-handshake-timeout: 1s\n"
	var key Key
	if err := yaml.Unmarshal([]byte(old), &key); err != nil {
		t.Fatal(err)
	}
	if key.TCPProxyHandshakeTimeout != 0 {
		t.Fatal("removed key still configures timeout")
	}
	decoder := yaml.NewDecoder(strings.NewReader(old))
	decoder.KnownFields(true)
	if err := decoder.Decode(&key); err == nil {
		t.Fatal("removed YAML field still recognized")
	}
}
