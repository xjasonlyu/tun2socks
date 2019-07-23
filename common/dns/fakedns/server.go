package fakedns

import (
	"errors"
	"net"
	"strings"

	D "github.com/miekg/dns"
	"github.com/xjasonlyu/tun2socks/common/fakeip"
	cache "github.com/xjasonlyu/tun2socks/common/lru-cache"
)

const (
	dnsFakeTTL    uint32 = 1
	dnsDefaultTTL uint32 = 600
)

// var cacheDuration = time.Duration(dnsDefaultTTL) * time.Second

type Server struct {
	*D.Server
	c *cache.Cache
	h handler
}

func (s *Server) ServeDNS(w D.ResponseWriter, r *D.Msg) {
	if len(r.Question) == 0 {
		D.HandleFailed(w, r)
		return
	}

	s.h(w, r)
}

func (s *Server) StartServer(addr string) error {
	_, port, err := net.SplitHostPort(addr)
	if port == "0" || port == "" || err != nil {
		return errors.New("address format error")
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	p, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	s.Server = &D.Server{Addr: addr, PacketConn: p, Handler: s}

	go func() {
		_ = s.ActivateAndServe()
	}()
	return nil
}

func (s *Server) IPToHost(ip net.IP) (string, bool) {
	c, ok := s.c.Peek(ip.String())
	if !ok {
		return "", false
	}
	fqdn := c.(*D.Msg).Question[0].Name
	return strings.TrimRight(fqdn, "."), true
}

func NewServer(fakeIPRange, hostsLine string, size int) (*Server, error) {
	_, ipnet, err := net.ParseCIDR(fakeIPRange)
	if err != nil {
		return nil, err
	}
	pool, err := fakeip.New(ipnet)
	if err != nil {
		return nil, err
	}

	var cacheItem *cache.Cache
	evictCallback := func(_ interface{}, value interface{}) {
		msg := value.(*D.Msg).Copy()
		q := msg.Question[0]
		ip := msg.Answer[0].(*D.A).A
		cacheItem.Remove("fakeip:" + q.String())
		cacheItem.Remove(ip.String())
	}
	cacheItem = cache.New(size, evictCallback)

	hosts := lineToHosts(hostsLine)
	handler := newHandler(hosts, cacheItem, pool)

	return &Server{
		c: cacheItem,
		h: handler,
	}, nil
}
