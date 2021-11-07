package remotedns

import (
	"github.com/miekg/dns"
	M "github.com/xjasonlyu/tun2socks/constant"
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/log"
	"net"
)

func RewriteMetadata(metadata *M.Metadata) bool {
	if !IsEnabled() {
		return false
	}
	dstName, found := getCachedName(metadata.DstIP)
	if !found {
		return false
	}
	metadata.VirtualIP = metadata.DstIP
	metadata.DstIP = nil
	metadata.DstName = dstName.(string)
	return true
}

func HandleDNSQuery(packet *core.UDPPacket) bool {
	if !IsEnabled() {
		return false
	}

	msg := dns.Msg{}
	err := msg.Unpack((*packet).Data())

	// Ignore UDP packets that are not IP queries to a recursive resolver
	if (*packet).ID().LocalPort != 53 || err != nil || len(msg.Question) == 0 || msg.Question[0].Qtype != dns.TypeA &&
		msg.Question[0].Qtype != dns.TypeAAAA || msg.Question[0].Qclass != dns.ClassINET ||
		!msg.RecursionDesired {
		return false
	}

	qname := msg.Question[0].Name
	qtype := msg.Question[0].Qtype

	msg.RecursionDesired = false
	msg.RecursionAvailable = true
	var hdr *dns.RR_Header
	var ip net.IP
	if qtype == dns.TypeA {
		rr := dns.A{}
		ip = insertNameIntoCache(4, qname)
		rr.A = ip
		hdr = &rr.Hdr
		msg.Answer = append(msg.Answer, &rr)
	} else {
		rr := dns.AAAA{}
		hdr = &rr.Hdr
		ip = insertNameIntoCache(6, qname)
		rr.AAAA = ip
		msg.Answer = append(msg.Answer, &rr)
	}
	if ip == nil {
		log.Warnf("[DNS] IP space exhausted")
		return true
	}
	(*hdr).Name = qname
	(*hdr).Ttl = _ttl
	(*hdr).Class = dns.ClassINET
	(*hdr).Rrtype = qtype

	packed, err := msg.Pack()
	if err != nil {
		return true
	}

	_, _ = (*packet).WriteBack(packed, (*packet).LocalAddr(), nil)

	log.Infof("[DNS] query %s %s", dns.TypeToString[qtype], qname)
	return true
}
