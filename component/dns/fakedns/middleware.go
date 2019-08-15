package fakedns

import (
	"net"
	"strings"

	D "github.com/miekg/dns"
	trie "github.com/xjasonlyu/tun2socks/common/domain-trie"
	"github.com/xjasonlyu/tun2socks/common/fakeip"
)

type handler func(w D.ResponseWriter, r *D.Msg)

func withFakeIP(pool *fakeip.Pool) handler {
	return func(w D.ResponseWriter, r *D.Msg) {
		q := r.Question[0]
		host := strings.TrimRight(q.Name, ".")

		rr := &D.A{}
		rr.Hdr = D.RR_Header{Name: q.Name, Rrtype: D.TypeA, Class: D.ClassINET, Ttl: dnsDefaultTTL}
		ip := pool.Lookup(host)
		rr.A = ip
		msg := r.Copy()
		msg.Answer = []D.RR{rr}

		setMsgTTL(msg, dnsFakeTTL)
		msg.SetReply(r)
		w.WriteMsg(msg)
		return
	}
}

func withHost(hosts *trie.Trie, next handler) handler {
	if hosts == nil {
		panic("dns/withHost: hosts should not be nil")
	}

	return func(w D.ResponseWriter, r *D.Msg) {
		q := r.Question[0]
		if q.Qtype != D.TypeA && q.Qtype != D.TypeAAAA {
			next(w, r)
			return
		}

		domain := strings.TrimRight(q.Name, ".")
		host := hosts.Search(domain)
		if host == nil {
			next(w, r)
			return
		}

		ip := host.Data.(net.IP)
		if q.Qtype == D.TypeAAAA && ip.To16() == nil {
			next(w, r)
			return
		} else if q.Qtype == D.TypeA && ip.To4() == nil {
			next(w, r)
			return
		}

		var rr D.RR
		if q.Qtype == D.TypeAAAA {
			record := &D.AAAA{}
			record.Hdr = D.RR_Header{Name: q.Name, Rrtype: D.TypeAAAA, Class: D.ClassINET, Ttl: dnsDefaultTTL}
			record.AAAA = ip
			rr = record
		} else {
			record := &D.A{}
			record.Hdr = D.RR_Header{Name: q.Name, Rrtype: D.TypeA, Class: D.ClassINET, Ttl: dnsDefaultTTL}
			record.A = ip
			rr = record
		}

		msg := r.Copy()
		msg.Answer = []D.RR{rr}
		msg.SetReply(r)
		w.WriteMsg(msg)
		return
	}
}

func newHandler(hosts *trie.Trie, pool *fakeip.Pool) handler {
	if hosts != nil {
		return withHost(hosts, withFakeIP(pool))
	}
	return withFakeIP(pool)
}
