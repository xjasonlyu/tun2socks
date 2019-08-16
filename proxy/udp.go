package proxy

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/common/lsof"
	"github.com/xjasonlyu/tun2socks/common/pool"
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/log"

	S "github.com/xjasonlyu/tun2socks/component/session"
)

type udpHandler struct {
	proxyHost string
	proxyPort int

	timeout   time.Duration
	hijackDNS []string

	remoteAddrMap sync.Map
	remoteConnMap sync.Map
}

func NewUDPHandler(proxyHost string, proxyPort int, timeout time.Duration, hijackDNS string) core.UDPConnHandler {
	return &udpHandler{
		proxyHost: proxyHost,
		proxyPort: proxyPort,
		timeout:   timeout,
		hijackDNS: strings.Split(hijackDNS, ","),
	}
}

func (h *udpHandler) isHijacked(target *net.UDPAddr) bool {
	for _, addr := range h.hijackDNS {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			continue
		}
		portInt, _ := strconv.Atoi(port)
		if (host == "*" && portInt == target.Port) || addr == target.String() {
			return true
		}
	}
	return false
}

func (h *udpHandler) fetchUDPInput(conn core.UDPConn, input net.PacketConn, addr *net.UDPAddr) {
	buf := pool.BufPool.Get().([]byte)

	defer func() {
		h.Close(conn)
		pool.BufPool.Put(buf[:cap(buf)])
	}()

	for {
		input.SetDeadline(time.Now().Add(h.timeout))
		n, _, err := input.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
				log.Warnf("failed to read UDP data from remote: %v", err)
			}
			return
		}

		if _, err := conn.WriteFrom(buf[:n], addr); err != nil {
			log.Warnf("failed to write UDP data: %v", err)
			return
		}
	}
}

func (h *udpHandler) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	// Check hijackDNS
	if h.isHijacked(target) {
		return nil
	}

	// Lookup fakeDNS host record
	targetHost, err := lookupHost(target)
	if err != nil {
		log.Warnf("lookup target host: %v", err)
		return err
	}

	proxyAddr := net.JoinHostPort(h.proxyHost, strconv.Itoa(h.proxyPort))
	targetAddr := net.JoinHostPort(targetHost, strconv.Itoa(target.Port))
	// Dial
	remoteConn, remoteAddr, err := dialUDP(proxyAddr, targetAddr)
	if err != nil {
		log.Infof("DialUDP: %v", err)
		return err
	}

	// Get name of the process
	var process = lsof.GetProcessName(conn.LocalAddr())
	if monitor != nil {
		session := &S.Session{
			Process:       process,
			Network:       conn.LocalAddr().Network(),
			DialerAddr:    remoteConn.LocalAddr().String(),
			ClientAddr:    conn.LocalAddr().String(),
			TargetAddr:    targetAddr,
			UploadBytes:   0,
			DownloadBytes: 0,
			SessionStart:  time.Now(),
		}
		addSession(conn, session)
		remoteConn = &S.PacketConn{Session: session, PacketConn: remoteConn}
	}

	h.remoteAddrMap.Store(conn, remoteAddr)
	h.remoteConnMap.Store(conn, remoteConn)

	go h.fetchUDPInput(conn, remoteConn, target)

	log.Access(process, "proxy", "udp", conn.LocalAddr().String(), targetAddr)
	return nil
}

func (h *udpHandler) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) (err error) {
	// Close if return error
	defer func() {
		if err != nil {
			h.Close(conn)
		}
	}()

	// Check hijackDNS
	if h.isHijacked(addr) {
		resp, err := fakeDNS.Resolve(data)
		if err != nil {
			return fmt.Errorf("hijack DNS request error: %v", err)
		} else {
			if _, err = conn.WriteFrom(resp, addr); err != nil {
				return fmt.Errorf("write dns answer failed: %v", err)
			}
			h.Close(conn)
			return nil
		}
	}

	var remoteAddr net.Addr
	var remoteConn net.PacketConn

	if value, ok := h.remoteAddrMap.Load(conn); ok {
		remoteAddr = value.(net.Addr)
	}

	if value, ok := h.remoteConnMap.Load(conn); ok {
		remoteConn = value.(net.PacketConn)
	}

	if remoteAddr == nil || remoteConn == nil {
		return fmt.Errorf("proxy connection %v->%v does not exists", conn.LocalAddr(), addr)
	}

	if _, err = remoteConn.WriteTo(data, remoteAddr); err != nil {
		return fmt.Errorf("write remote failed: %v", err)
	}

	return nil
}

func (h *udpHandler) Close(conn core.UDPConn) {
	// Close
	conn.Close()

	// Load from remoteConnMap
	if remoteConn, ok := h.remoteConnMap.Load(conn); ok {
		remoteConn.(net.PacketConn).Close()
	} else {
		return
	}

	h.remoteAddrMap.Delete(conn)
	h.remoteConnMap.Delete(conn)

	// Remove session
	removeSession(conn)
}
