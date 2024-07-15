package dns

import (
	"errors"
	"net"
	"net/netip"
	"strings"

	D "github.com/miekg/dns"

	"github.com/xjasonlyu/tun2socks/v2/component/fakeip"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
)

var (
	fakePool       *fakeip.Pool
	fakeDNSenabled = false
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
	fakeDNSenabled = true
}

func ProcessMetadata(metadata *M.Metadata) bool {
	if !fakeDNSenabled {
		return false
	}
	dstName, found := fakePool.LookBack(net.IP(metadata.DstIP.AsSlice()))
	if !found {
		return false
	}
	metadata.DstIP = netip.Addr{}
	metadata.DstName = dstName
	return true
}

func fakeipHandler(fakePool *fakeip.Pool) handler {
	return func(r *D.Msg) (*D.Msg, error) {
		if len(r.Question) == 0 {
			return nil, errors.New("at least one question is required")
		}

		q := r.Question[0]

		host := strings.TrimRight(q.Name, ".")
		msg := r.Copy()

		if q.Qtype == D.TypeA {
			rr := &D.A{}
			rr.Hdr = D.RR_Header{Name: q.Name, Rrtype: D.TypeA, Class: D.ClassINET, Ttl: dnsDefaultTTL}
			ip := fakePool.Lookup(host)
			rr.A = ip
			msg.Answer = []D.RR{rr}
		}

		setMsgTTL(msg, 1)
		msg.SetRcode(r, D.RcodeSuccess)
		msg.RecursionAvailable = true
		msg.Response = true

		return msg, nil
	}
}
