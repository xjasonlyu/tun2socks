package dns

import (
	"net"
	"net/netip"
	"testing"

	D "github.com/miekg/dns"
	"github.com/stretchr/testify/assert"

	"github.com/xjasonlyu/tun2socks/v2/component/fakeip"
	"github.com/xjasonlyu/tun2socks/v2/component/trie"
)

func newTestPool(t *testing.T) *fakeip.Pool {
	t.Helper()
	_, cidr, err := net.ParseCIDR("198.18.0.0/16")
	if err != nil {
		t.Fatal(err)
	}
	pool, err := fakeip.New(fakeip.Options{
		IPNet: cidr,
		Size:  10,
		Host:  trie.New(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return pool
}

// TestIsFakeDNSQuery_TunnelHijack verifies that a query addressed to the
// configured fake DNS listen address is recognized, independent of whether
// ReCreateServer's real OS-level socket bind succeeds. This is what allows
// platforms like Android — where all traffic, including DNS, is captured by
// the TUN device and never reaches a real socket — to answer fake DNS
// queries via the in-tunnel path instead.
func TestIsFakeDNSQuery_TunnelHijack(t *testing.T) {
	t.Cleanup(func() { DisableFakeDNS() })

	pool := newTestPool(t)

	want := netip.MustParseAddrPort("127.0.0.1:15353")
	assert.False(t, IsFakeDNSQuery(want), "must be false before fake DNS is configured")

	ReCreateServer("127.0.0.1:15353", pool)
	EnableFakeDNS()

	assert.True(t, IsFakeDNSQuery(want))
	assert.False(t, IsFakeDNSQuery(netip.MustParseAddrPort("127.0.0.1:9999")))
}

// TestHandleQuery_TunnelPath verifies that a raw DNS query byte slice, as
// would be read directly from a UDP/TCP flow intercepted in the tunnel, is
// answered with a fake IP from the pool, and that IP resolves back to the
// queried hostname via the pool's LookBack — exactly the round trip
// tunnel/udp.go's handleFakeDNSUDP and tunnel/tcp.go's handleFakeDNSTCP rely
// on.
func TestHandleQuery_TunnelPath(t *testing.T) {
	t.Cleanup(func() { DisableFakeDNS() })

	pool := newTestPool(t)
	ReCreateServer("127.0.0.1:0", pool)
	EnableFakeDNS()

	query := new(D.Msg)
	query.SetQuestion(D.Fqdn("example.com"), D.TypeA)

	raw, err := query.Pack()
	assert.NoError(t, err)

	respRaw, err := HandleQuery(raw)
	assert.NoError(t, err)

	resp := new(D.Msg)
	assert.NoError(t, resp.Unpack(respRaw))
	assert.Len(t, resp.Answer, 1)

	a, ok := resp.Answer[0].(*D.A)
	assert.True(t, ok)

	host, found := pool.LookBack(a.A)
	assert.True(t, found)
	assert.Equal(t, "example.com", host)
}
