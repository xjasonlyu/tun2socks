package proxy

import (
	"fmt"
	"net"
	"strconv"
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

	remoteMap sync.Map
}

func NewUDPHandler(proxyHost string, proxyPort int, timeout time.Duration) core.UDPConnHandler {
	return &udpHandler{
		proxyHost: proxyHost,
		proxyPort: proxyPort,
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
			if !isTimeout(err) && !isClosed(err) {
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
	if isHijacked(target) {
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

	var process = "N/A"
	if monitor != nil {
		// Get name of the process
		process = lsof.GetProcessName(conn.LocalAddr())
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

	h.remoteMap.Store(conn, &udpElement{
		remoteAddr: remoteAddr,
		remoteConn: remoteConn,
	})

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
	if isHijacked(addr) {
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

	if elm, ok := h.remoteMap.Load(conn); ok {
		remoteAddr = elm.(*udpElement).remoteAddr
		remoteConn = elm.(*udpElement).remoteConn
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
	if elm, ok := h.remoteMap.Load(conn); ok {
		elm.(*udpElement).remoteConn.Close()
		h.remoteMap.Delete(conn)
	} else {
		return
	}

	// Remove session
	removeSession(conn)
}
