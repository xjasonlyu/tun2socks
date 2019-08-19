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

type Resolver struct {
	b []string
	h *T.Trie
	p *F.Pool
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

	msg := resolve(r.h, r.p, r.b, req)
	if msg == nil {
		return nil, errors.New("cannot resolve dns query: msg is nil")
	}
	resp, err := msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("failed to pack dns answer: %v", err)
	}
	return resp, nil
}

func NewResolver(h, b string) (*Resolver, error) {
	_, ipnet, _ := net.ParseCIDR(dnsFakeIPRange)

	// fake ip should start with "198.18.0.3".
	pool, err := F.New(ipnet, 3, dnsCacheSize)
	if err != nil {
		return nil, err
	}

	hosts := func(str string) *T.Trie {
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

	return &Resolver{
		b: strings.Split(b, ","),
		h: hosts,
		p: pool,
	}, nil
}
