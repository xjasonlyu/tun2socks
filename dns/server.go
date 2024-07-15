package dns

import (
	"errors"
	"net"

	D "github.com/miekg/dns"

	"github.com/xjasonlyu/tun2socks/v2/common/sockopt"
	"github.com/xjasonlyu/tun2socks/v2/component/fakeip"
	"github.com/xjasonlyu/tun2socks/v2/log"
)

var (
	server               = &Server{}
	dnsDefaultTTL uint32 = 600
)

type (
	handler func(r *D.Msg) (*D.Msg, error)
)

type Server struct {
	*D.Server
	handler handler
}

// ServeDNS implement D.Handler ServeDNS
func (s *Server) ServeDNS(w D.ResponseWriter, r *D.Msg) {
	msg, err := handlerWithContext(s.handler, r)
	if err != nil {
		D.HandleFailed(w, r)
		return
	}
	msg.Compress = true
	w.WriteMsg(msg)
}

func handlerWithContext(handler handler, msg *D.Msg) (*D.Msg, error) {
	if len(msg.Question) == 0 {
		return nil, errors.New("at least one question is required")
	}

	return handler(msg)
}

func ReCreateServer(addr string, pool *fakeip.Pool) {
	fakePool = pool
	if server.Server != nil {
		server.Shutdown()
		server = &Server{}
	}

	if addr == "" {
		return
	}

	var err error
	defer func() {
		if err != nil {
			log.Errorf("Start DNS server error: %s", err.Error())
		}
	}()

	_, port, err := net.SplitHostPort(addr)
	if port == "0" || port == "" || err != nil {
		return
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return
	}

	err = sockopt.UDPReuseaddr(udpConn)
	if err != nil {
		log.Warnf("Failed to Reuse UDP Address: %s", err)

		err = nil
	}

	server = &Server{handler: fakeipHandler(fakePool)}
	server.Server = &D.Server{Addr: addr, PacketConn: udpConn, Handler: server}

	go func() {
		server.ActivateAndServe()
	}()

	log.Infof("DNS server listening at: %s", udpConn.LocalAddr().String())
}
