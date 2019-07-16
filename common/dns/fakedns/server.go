package fakedns

import (
	"errors"
	"net"
	"strings"
	"time"

	D "github.com/miekg/dns"
)

var (
	cacheDuration = 60 * time.Second

	dnsFakeTTL    uint32 = 1
	dnsDefaultTTL uint32 = 600
)

type Server struct {
	*D.Server
	p *Pool
	c *Cache
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

func (s *Server) IPToHost(ip net.IP) string {
	c := s.c.Get(ip.String())
	if c == nil {
		return ""
	}
	fqdn := c.(*D.Msg).Question[0].Name
	return strings.TrimRight(fqdn, ".")
}

func (s *Server) IsFakeIP(ip net.IP) bool {
	c := ip2uint32(ip)
	if c >= s.p.min && c <= s.p.max {
		return true
	}
	return false
}

func NewServer(fakeIPRange, hostsMap string) (*Server, error) {
	_, ipnet, err := net.ParseCIDR(fakeIPRange)
	if err != nil {
		return nil, err
	}
	pool, err := NewPool(ipnet)
	if err != nil {
		return nil, err
	}

	cache := NewCache(cacheDuration)
	hosts := strToHosts(hostsMap)
	handler := newHandler(hosts, cache, pool)

	return &Server{
		p: pool,
		c: cache,
		h: handler,
	}, nil
}
