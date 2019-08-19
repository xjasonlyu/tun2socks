package fakedns

import (
	"errors"
	"fmt"
	"net"
	"strings"

	D "github.com/miekg/dns"

	T "github.com/xjasonlyu/tun2socks/common/domain-trie"
	F "github.com/xjasonlyu/tun2socks/common/fakeip"
)

const (
	// TTL
	dnsFakeTTL    uint32 = 1
	dnsDefaultTTL uint32 = 600

	// Resolver default value
	dnsCacheSize   = 1000
	dnsFakeIPRange = "198.18.0.0/15"
)

type handler = D.HandlerFunc

type Resolver struct {
	h handler
	p *F.Pool
	t *T.Trie

	*D.Server
	ServeAddr string
}

func (r *Resolver) ServeDNS(w D.ResponseWriter, req *D.Msg) {
	if len(req.Question) == 0 {
		D.HandleFailed(w, req)
		return
	}
	r.h(w, req)
}

func (r *Resolver) Start() error {
	_, port, err := net.SplitHostPort(r.ServeAddr)
	if port == "0" || port == "" || err != nil {
		return errors.New("address format error")
	}

	udpAddr, err := net.ResolveUDPAddr("udp", r.ServeAddr)
	if err != nil {
		return err
	}

	pc, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	r.Server = &D.Server{Addr: r.ServeAddr, PacketConn: pc, Handler: r.h}
	go func() {
		r.ActivateAndServe()
	}()

	return nil
}

func (r *Resolver) Stop() error {
	return r.Shutdown()
}

func (r *Resolver) IPToHost(ip net.IP) (string, bool) {
	return r.p.LookBack(ip)
}

func (r *Resolver) Resolve(request []byte) ([]byte, error) {
	if err := D.IsMsg(request); err != nil {
		return nil, err
	}

	req := new(D.Msg)
	if err := req.Unpack(request); err != nil {
		return nil, errors.New("cannot handle dns query: failed to unpack")
	}

	if len(req.Question) == 0 {
		return nil, errors.New("cannot handle dns query: invalid question length")
	}

	msg := resolve(r.t, r.p, req)
	if msg == nil {
		return nil, errors.New("cannot resolve dns query: msg is nil")
	}
	resp, err := msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("failed to pack dns answer: %v", err)
	}
	return resp, nil
}

func NewResolver(a, h string) (*Resolver, error) {
	_, ipnet, _ := net.ParseCIDR(dnsFakeIPRange)

	pool, err := F.New(ipnet, dnsCacheSize)
	if err != nil {
		return nil, err
	}

	tree := func(str string) *T.Trie {
		tree := T.New()
		s := strings.Split(str, ",")
		for _, host := range s {
			m := strings.Split(host, "=")
			if len(m) != 2 {
				continue
			}
			domain := strings.TrimSpace(m[0])
			target := strings.TrimSpace(m[1])
			if err := tree.Insert(domain, net.ParseIP(target)); err != nil {
				panic(fmt.Sprintf("add hosts error: %v", err))
			}
		}
		return tree
	}(h)

	handler := newHandler(tree, pool)
	return &Resolver{
		h:         handler,
		p:         pool,
		t:         tree,
		ServeAddr: a,
	}, nil
}
