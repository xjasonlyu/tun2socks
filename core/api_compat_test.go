package core

import (
	"testing"

	"github.com/xjasonlyu/tun2socks/v2/core/adapter"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// Existing signatures and named-field configurations remain compatible.
// Positional struct literals are intentionally not a compatibility requirement.
var _ func(*Config) (*stack.Stack, error) = CreateStack
var _ = Config{TransportHandler: legacyOnlyHandler{}}

type legacyOnlyHandler struct{}

func (legacyOnlyHandler) HandleTCP(c adapter.TCPConn) { c.Close() }
func (legacyOnlyHandler) HandleUDP(c adapter.UDPConn) { c.Close() }

var _ adapter.TransportHandler = legacyOnlyHandler{}

func TestDeferredModeRejectsUnsupportedHandler(t *testing.T) {
	if s, err := CreateStack(&Config{TransportHandler: legacyOnlyHandler{}, DeferredTCP: &DeferredTCPConfig{}}); err == nil || s != nil {
		t.Fatal("explicit opt-in must not silently fall back to early handshake")
	}
}
