package proxy

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/common/lsof"
	"github.com/xjasonlyu/tun2socks/common/pool"
	D "github.com/xjasonlyu/tun2socks/component/fakedns"
	S "github.com/xjasonlyu/tun2socks/component/session"
	C "github.com/xjasonlyu/tun2socks/constant"
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/log"
)

type udpHandler struct {
	proxyHost string
	proxyPort int
	timeout   time.Duration

	remoteAddrMap sync.Map
	remoteConnMap sync.Map

	fakeDNS D.FakeDNS
	monitor S.Monitor
}

func NewUDPHandler(proxyHost string, proxyPort int, timeout time.Duration, fakeDNS D.FakeDNS, monitor S.Monitor) core.UDPConnHandler {
	return &udpHandler{
		proxyHost: proxyHost,
		proxyPort: proxyPort,
		fakeDNS:   fakeDNS,
		monitor:   monitor,
		timeout:   timeout,
	}
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
	// Lookup fakeDNS host record
	targetHost, err := lookupHost(h.fakeDNS, target)
	if err != nil {
		log.Warnf("lookup target host error: %v", err)
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
	if h.monitor != nil {
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
		h.monitor.AddSession(conn, session)

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
	// Load from remoteConnMap
	if remoteConn, ok := h.remoteConnMap.Load(conn); ok {
		remoteConn.(net.PacketConn).Close()
	} else {
		return
	}

	h.remoteAddrMap.Delete(conn)
	h.remoteConnMap.Delete(conn)

	// Close
	conn.Close()

	// Remove session
	if h.monitor != nil {
		h.monitor.RemoveSession(conn)
	}
}
