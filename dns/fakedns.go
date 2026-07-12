package dns

import (
	"errors"
	"net"
	"net/netip"
	"strings"
	"sync"

	D "github.com/miekg/dns"

	"github.com/xjasonlyu/tun2socks/v2/component/fakeip"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
)

var (
	// fakeMu guards fakePool, fakeDNSenabled and fakeListenAddr, which are
	// read from every TCP/UDP connection goroutine via ProcessMetadata and
	// IsFakeDNSQuery, and written once from the engine goroutine via
	// EnableFakeDNS/DisableFakeDNS/ReCreateServer.
	fakeMu         sync.RWMutex
	fakePool       *fakeip.Pool
	fakeDNSenabled bool
	// fakeListenAddr is the address fake DNS queries are expected on. It is
	// used to recognize fake DNS traffic that arrives through the tunnel
	// (e.g. on platforms like Android, where all traffic including DNS is
	// captured by the TUN device and never reaches a real OS-level socket),
	// independent of whether ReCreateServer managed to bind a real listener.
	fakeListenAddr netip.AddrPort
)

func setMsgTTL(msg *D.Msg, ttl uint32) {
	for _, answer := range msg.Answer {
		answer.Header().Ttl = ttl
	}

	for _, ns := range msg.Ns {
		ns.Header().Ttl = ttl
	}

	for _, extra := range msg.Extra {
		extra.Header().Ttl = ttl
	}
}

func EnableFakeDNS() {
	fakeMu.Lock()
	fakeDNSenabled = true
	fakeMu.Unlock()
}

// DisableFakeDNS turns fake DNS off and stops the DNS listener, if any.
func DisableFakeDNS() {
	fakeMu.Lock()
	fakeDNSenabled = false
	fakeMu.Unlock()
	ReCreateServer("", nil)
}

// IsFakeDNSQuery reports whether dst is the configured fake DNS listen
// address, i.e. a UDP/TCP flow to dst should be answered locally via
// HandleQuery instead of being dialed out through the proxy.
func IsFakeDNSQuery(dst netip.AddrPort) bool {
	fakeMu.RLock()
	defer fakeMu.RUnlock()
	return fakeDNSenabled && fakeListenAddr.IsValid() && dst == fakeListenAddr
}

// HandleQuery answers a raw DNS query message using the fake IP pool, for
// callers that intercept DNS traffic directly from the tunnel rather than
// through the real OS-level socket managed by ReCreateServer.
func HandleQuery(data []byte) ([]byte, error) {
	var msg D.Msg
	if err := msg.Unpack(data); err != nil {
		return nil, err
	}
	resp, err := fakeipHandler()(&msg)
	if err != nil {
		return nil, err
	}
	return resp.Pack()
}

func ProcessMetadata(metadata *M.Metadata) bool {
	fakeMu.RLock()
	enabled, pool := fakeDNSenabled, fakePool
	fakeMu.RUnlock()

	if !enabled || pool == nil {
		return false
	}
	dstName, found := pool.LookBack(net.IP(metadata.DstIP.AsSlice()))
	if !found {
		return false
	}
	metadata.DstIP = netip.Addr{}
	metadata.DstName = dstName
	return true
}

func fakeipHandler() handler {
	return func(r *D.Msg) (*D.Msg, error) {
		if len(r.Question) == 0 {
			return nil, errors.New("at least one question is required")
		}

		fakeMu.RLock()
		pool := fakePool
		fakeMu.RUnlock()
		if pool == nil {
			return nil, errors.New("fake DNS pool not initialized")
		}

		q := r.Question[0]

		host := strings.TrimRight(q.Name, ".")
		msg := r.Copy()

		if q.Qtype == D.TypeA {
			rr := &D.A{}
			rr.Hdr = D.RR_Header{Name: q.Name, Rrtype: D.TypeA, Class: D.ClassINET}
			rr.A = pool.Lookup(host)
			msg.Answer = []D.RR{rr}
		}

		// Fake IPs can be recycled by the pool at any time, so clients must
		// not cache them for long.
		setMsgTTL(msg, 1)
		msg.SetRcode(r, D.RcodeSuccess)
		msg.RecursionAvailable = true
		msg.Response = true

		return msg, nil
	}
}
