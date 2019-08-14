package fakedns

import (
	"errors"
	"net"

	D "github.com/miekg/dns"
	"github.com/xjasonlyu/tun2socks/common/fakeip"
	"github.com/xjasonlyu/tun2socks/log"
)

const (
	dnsFakeTTL    uint32 = 1
	dnsDefaultTTL uint32 = 600
)

var (
	ServeAddr = "127.0.0.1:5353"
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

func (s *Server) Start() error {
	log.Debugf("Start fake DNS server")
	_, port, err := net.SplitHostPort(ServeAddr)
	if port == "0" || port == "" || err != nil {
		return errors.New("address format error")
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ServeAddr)
	if err != nil {
		return err
	}

	p, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	s.Server = &D.Server{Addr: ServeAddr, PacketConn: p, Handler: s}
	go s.ActivateAndServe()
	return nil
}

func (s *Server) Stop() error {
	log.Debugf("Stop fake DNS server")
	return s.Shutdown()
}

func (s *Server) IPToHost(ip net.IP) (string, bool) {
	return s.p.LookBack(ip)
}

func NewServer(fakeIPRange, hostsLine string, size int) (*Server, error) {
	_, ipnet, err := net.ParseCIDR(fakeIPRange)
	if err != nil {
		return nil, err
	}
	pool, err := fakeip.New(ipnet, size)
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
