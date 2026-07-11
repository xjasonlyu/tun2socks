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
	// fakeMu guards fakePool and fakeDNSenabled, which are read from every
	// TCP/UDP connection goroutine via ProcessMetadata and written once from
	// the engine goroutine via EnableFakeDNS/DisableFakeDNS/ReCreateServer.
	fakeMu         sync.RWMutex
	fakePool       *fakeip.Pool
	fakeDNSenabled bool
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
