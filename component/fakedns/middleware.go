package fakedns

import (
	"net"
	"strings"

	trie "github.com/xjasonlyu/tun2socks/common/domain-trie"
	"github.com/xjasonlyu/tun2socks/common/fakeip"

	D "github.com/miekg/dns"
)

var (
	nameserver = "8.8.8.8:53"
)

func dnsExchange(r *D.Msg) *D.Msg {
	c := new(D.Client)
	c.Net = "tcp"
	msg, _, err := c.Exchange(r, nameserver)
	if err != nil {
		// empty DNS response
		rr := &D.A{}
		rr.Hdr = D.RR_Header{Name: r.Question[0].Name, Rrtype: D.TypeA, Class: D.ClassINET, Ttl: dnsDefaultTTL}
		msg = r.Copy()
		msg.Answer = []D.RR{rr}
		setMsgTTL(msg, dnsDefaultTTL)
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
