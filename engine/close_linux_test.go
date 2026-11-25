//go:build linux

package engine

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

// Opt-in: creates only a temporary TUN, without routes, DNS or an external
// proxy connection. Run on a native Linux host with CAP_NET_ADMIN.
func TestLinuxTUNRepeatedStartStop(t *testing.T) {
	if os.Getenv("TUN2SOCKS_TEST_TUN") != "1" {
		t.Skip("set TUN2SOCKS_TEST_TUN=1 to test real Linux TUN lifecycle")
	}
	name := fmt.Sprintf("gct%d", os.Getpid())
	if _, err := net.InterfaceByName(name); err == nil {
		t.Fatal("test interface already exists")
	}
	baseline := 0
	for i := 0; i < 20; i++ {
		Insert(&Key{Device: "tun://" + name, Proxy: "socks5://127.0.0.1:1", MTU: 1400, LogLevel: "error", TCPDeferredHandshake: true})
		if err := Start(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(50 * time.Millisecond)
		started := time.Now()
		done := make(chan error, 1)
		go func() { done <- Stop() }()
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("iteration %d: Stop blocked", i+1)
		}
		if _, err := net.InterfaceByName(name); err == nil {
			t.Fatal("TUN remained after Stop")
		}
		if err := Stop(); err != nil {
			t.Fatal("repeated Stop:", err)
		}
		fds, err := os.ReadDir("/proc/self/fd")
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			baseline = len(fds)
		} else if len(fds) > baseline {
			t.Fatalf("descriptor leak: %d after first cycle, %d after cycle %d", baseline, len(fds), i+1)
		}
		t.Logf("cycle %d: Stop=%s, descriptors=%d", i+1, time.Since(started), len(fds))
	}
}
