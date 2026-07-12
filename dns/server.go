package dns

import (
	"errors"
	"net"
	"net/netip"

	D "github.com/miekg/dns"

	"github.com/xjasonlyu/tun2socks/v2/common/sockopt"
	"github.com/xjasonlyu/tun2socks/v2/component/fakeip"
	"github.com/xjasonlyu/tun2socks/v2/log"
)

var server = &Server{}

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

// ReCreateServer (re)starts the fake DNS UDP listener at addr using pool.
// Passing an empty addr or a nil pool stops any running server.
func ReCreateServer(addr string, pool *fakeip.Pool) {
	fakeMu.Lock()
	defer fakeMu.Unlock()

	fakePool = pool
	fakeListenAddr = netip.AddrPort{}
	if server.Server != nil {
		server.Shutdown()
		server = &Server{}
	}

	if addr == "" || pool == nil {
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
	// Recorded regardless of whether the OS-level bind below succeeds, so
	// that fake DNS queries arriving through the tunnel (see
	// IsFakeDNSQuery/HandleQuery) are still answered even on platforms
	// where binding this address as a real local socket isn't possible.
	// Unmap in case ResolveUDPAddr returned an IPv4-in-6 form: addresses
	// coming from the tunnel (gVisor) are always plain IPv4, so comparing
	// against a mapped address in IsFakeDNSQuery would never match.
	fakeListenAddr = netip.AddrPortFrom(udpAddr.AddrPort().Addr().Unmap(), uint16(udpAddr.Port))

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return
	}

	err = sockopt.UDPReuseaddr(udpConn)
	if err != nil {
		log.Warnf("Failed to Reuse UDP Address: %s", err)

		err = nil
	}

	server = &Server{handler: fakeipHandler()}
	server.Server = &D.Server{Addr: addr, PacketConn: udpConn, Handler: server}

	// Capture the current *Server so a concurrent ReCreateServer call
	// swapping the package-level `server` var can't redirect this goroutine
	// to serve on the wrong instance.
	srv := server
	go func() {
		srv.ActivateAndServe()
	}()

	log.Infof("DNS server listening at: %s", udpConn.LocalAddr().String())
}
