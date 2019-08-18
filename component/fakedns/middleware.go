package fakedns

import (
	"net"
	"strings"

	trie "github.com/xjasonlyu/tun2socks/common/domain-trie"
	"github.com/xjasonlyu/tun2socks/common/fakeip"
	"github.com/xjasonlyu/tun2socks/log"

	D "github.com/miekg/dns"
)

var (
	BackendDNS []string
)

func dnsExchange(r *D.Msg) (msg *D.Msg) {
	defer func() {
		if msg == nil {
			// empty DNS response
			rr := &D.A{}
			rr.Hdr = D.RR_Header{Name: r.Question[0].Name, Rrtype: D.TypeA, Class: D.ClassINET, Ttl: dnsDefaultTTL}
			msg = r.Copy()
			msg.Answer = []D.RR{rr}
			setMsgTTL(msg, dnsDefaultTTL)
		}
	}()

	c := new(D.Client)
	c.Net = "tcp"
	for _, dns := range BackendDNS {
		msg, _, _ = c.Exchange(r, dns)
		if msg != nil {
			break
		}
	}
	return msg
}

func resolve(hosts *trie.Trie, pool *fakeip.Pool, r *D.Msg) (msg *D.Msg) {
	defer func() {
		if msg != nil {
			msg.SetReply(r)
		}
	}()

	if msg = hostResolve(hosts, r); msg != nil {
		return msg
	}

	q := r.Question[0]
	if q.Qtype != D.TypeA || q.Qclass != D.ClassINET {
		log.Debugf("DNS Query: %v %v %v", q.Name, q.Qclass, q.Qtype)
		return dnsExchange(r)
	}

	return fakeResolve(pool, r)
}

func fakeResolve(pool *fakeip.Pool, r *D.Msg) *D.Msg {
	q := r.Question[0]
	host := strings.TrimRight(q.Name, ".")

	rr := &D.A{}
	rr.Hdr = D.RR_Header{Name: q.Name, Rrtype: D.TypeA, Class: D.ClassINET, Ttl: dnsFakeTTL}
	ip := pool.Lookup(host)
	rr.A = ip

	msg := r.Copy()
	msg.Answer = []D.RR{rr}
	setMsgTTL(msg, dnsFakeTTL)
	return msg
}

func hostResolve(hosts *trie.Trie, r *D.Msg) *D.Msg {
	if hosts == nil {
		return nil
	}

	q := r.Question[0]
	if q.Qtype != D.TypeA && q.Qtype != D.TypeAAAA {
		return nil
	}

	domain := strings.TrimRight(q.Name, ".")
	host := hosts.Search(domain)
	if host == nil {
		return nil
	}

	ip := host.Data.(net.IP)
	if q.Qtype == D.TypeAAAA && ip.To16() == nil {
		return nil
	} else if q.Qtype == D.TypeA && ip.To4() == nil {
		return nil
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
	setMsgTTL(msg, dnsDefaultTTL)
	return msg
}

func newHandler(hosts *trie.Trie, pool *fakeip.Pool) D.HandlerFunc {
	return func(w D.ResponseWriter, r *D.Msg) {
		msg := resolve(hosts, pool, r)
		w.WriteMsg(msg)
		return
	}
}
