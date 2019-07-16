package shadowsocks

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	sscore "github.com/shadowsocks/go-shadowsocks2/core"
	sssocks "github.com/shadowsocks/go-shadowsocks2/socks"

	"github.com/xjasonlyu/tun2socks/common/dns"
	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/core"
)

type udpHandler struct {
	sync.Mutex

	cipher     sscore.Cipher
	remoteAddr net.Addr
	conns      map[core.UDPConn]net.PacketConn
	dnsCache   dns.DnsCache
	fakeDns    dns.FakeDns
	timeout    time.Duration
}

func NewUDPHandler(server, cipher, password string, timeout time.Duration, dnsCache dns.DnsCache, fakeDns dns.FakeDns) core.UDPConnHandler {
	ciph, err := sscore.PickCipher(cipher, []byte{}, password)
	if err != nil {
		log.Errorf("failed to pick a cipher: %v", err)
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		log.Errorf("failed to resolve udp address: %v", err)
	}

	return &udpHandler{
		cipher:     ciph,
		remoteAddr: remoteAddr,
		conns:      make(map[core.UDPConn]net.PacketConn, 16),
		dnsCache:   dnsCache,
		fakeDns:    fakeDns,
		timeout:    timeout,
	}
}

func (h *udpHandler) fetchUDPInput(conn core.UDPConn, input net.PacketConn) {
	buf := core.NewBytes(core.BufSize)

	defer func() {
		h.Close(conn)
		core.FreeBytes(buf)
	}()

	for {
		input.SetDeadline(time.Now().Add(h.timeout))
		n, _, err := input.ReadFrom(buf)
		if err != nil {
			// log.Printf("read remote failed: %v", err)
			return
		}

		addr := sssocks.SplitAddr(buf[:])
		resolvedAddr, err := net.ResolveUDPAddr("udp", addr.String())
		if err != nil {
			return
		}
		_, err = conn.WriteFrom(buf[int(len(addr)):n], resolvedAddr)
		if err != nil {
			log.Warnf("write local failed: %v", err)
			return
		}

		if h.dnsCache != nil {
			_, port, err := net.SplitHostPort(addr.String())
			if err != nil {
				panic("impossible error")
			}
			if port == strconv.Itoa(dns.CommonDnsPort) {
				h.dnsCache.Store(buf[int(len(addr)):n])
				return // DNS response
			}
		}
	}
}

func (h *udpHandler) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	pc, err := net.ListenPacket("udp", "")
	if err != nil {
		return err
	}
	pc = h.cipher.PacketConn(pc)

	h.Lock()
	h.conns[conn] = pc
	h.Unlock()
	go h.fetchUDPInput(conn, pc)
	if target != nil {
		log.Infof("new proxy connection for target: %s:%s", target.Network(), target.String())
	}
	return nil
}

func (h *udpHandler) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	h.Lock()
	pc, ok1 := h.conns[conn]
	h.Unlock()

	if addr.Port == dns.CommonDnsPort {
		if h.fakeDns != nil {
			resp, err := h.fakeDns.GenerateFakeResponse(data)
			if err == nil {
				_, err = conn.WriteFrom(resp, addr)
				if err != nil {
					return errors.New(fmt.Sprintf("write dns answer failed: %v", err))
				}
				h.Close(conn)
				return nil
			}
		}

		if h.dnsCache != nil {
			if answer := h.dnsCache.Query(data); answer != nil {
				_, err := conn.WriteFrom(answer, addr)
				if err != nil {
					return errors.New(fmt.Sprintf("cache dns answer failed: %v", err))
				}
				h.Close(conn)
				return nil
			}
		}
	}

	if ok1 {
		// Replace with a domain name if target address IP is a fake IP.
		var targetHost string
		if h.fakeDns != nil && h.fakeDns.IsFakeIP(addr.IP) {
			targetHost = h.fakeDns.QueryDomain(addr.IP)
		} else {
			targetHost = addr.IP.String()
		}
		dest := net.JoinHostPort(targetHost, strconv.Itoa(addr.Port))

		buf := append([]byte{0, 0, 0}, sssocks.ParseAddr(dest)...)
		buf = append(buf, data[:]...)
		_, err := pc.WriteTo(buf[3:], h.remoteAddr)
		if err != nil {
			h.Close(conn)
			return errors.New(fmt.Sprintf("write remote failed: %v", err))
		}
		return nil
	} else {
		h.Close(conn)
		return errors.New(fmt.Sprintf("proxy connection %v->%v does not exists", conn.LocalAddr(), addr))
	}
}

func (h *udpHandler) Close(conn core.UDPConn) {
	conn.Close()

	h.Lock()
	defer h.Unlock()

	if pc, ok := h.conns[conn]; ok {
		pc.Close()
		delete(h.conns, conn)
	}
}
