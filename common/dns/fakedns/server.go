package fakedns

import (
	"errors"
	"net"

	D "github.com/miekg/dns"
	"github.com/xjasonlyu/tun2socks/common/fakeip"
)

const (
	lruCacheSize = 1000

	dnsFakeTTL    uint32 = 1
	dnsDefaultTTL uint32 = 600
)

type Server struct {
	*D.Server
	p *fakeip.Pool
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
	go s.ActivateAndServe()

	return nil
}

func (s *Server) IPToHost(ip net.IP) (string, bool) {
	return s.p.LookBack(ip)
}

func NewServer(fakeIPRange, hostsLine string) (*Server, error) {
	_, ipnet, err := net.ParseCIDR(fakeIPRange)
	if err != nil {
		return nil, err
	}
	pool, err := fakeip.New(ipnet, lruCacheSize)
	if err != nil {
		return nil, err
	}

	hosts := lineToHosts(hostsLine)
	handler := newHandler(hosts, pool)

	return &Server{
		p: pool,
		h: handler,
	}, nil
}
