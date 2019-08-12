package d

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/lsof"
	"github.com/xjasonlyu/tun2socks/common/pool"
	"github.com/xjasonlyu/tun2socks/core"
)

type udpHandler struct {
	timeout time.Duration

	exceptionConnMap sync.Map

	exceptionApps []string
	sendThrough   net.Addr
	proxyHandler  core.UDPConnHandler
}

func (h *udpHandler) isExceptionApp(name string) bool {
	for _, app := range h.exceptionApps {
		if name == app {
			return true
		}
	}
	return false
}

func NewUDPHandler(proxyHandler core.UDPConnHandler, exceptionApps []string, sendThrough net.Addr, timeout time.Duration) core.UDPConnHandler {
	return &udpHandler{
		proxyHandler:  proxyHandler,
		exceptionApps: exceptionApps,
		sendThrough:   sendThrough,
		timeout:       timeout,
	}
}

func (h *udpHandler) handleInput(conn core.UDPConn, pc *net.UDPConn) {
	buf := pool.BufPool.Get().([]byte)

	defer func() {
		h.Close(conn)
		pool.BufPool.Put(buf[:cap(buf)])
	}()

	for {
		pc.SetDeadline(time.Now().Add(h.timeout))
		n, addr, err := pc.ReadFromUDP(buf)
		if err != nil {
			return
		}

		if _, err := conn.WriteFrom(buf[:n], addr); err != nil {
			return
		}
	}
}

func (h *udpHandler) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	localHost, localPortStr, _ := net.SplitHostPort(conn.LocalAddr().String())
	localPortInt, _ := strconv.Atoi(localPortStr)
	cmd, err := lsof.GetCommandNameBySocket("udp", localHost, uint16(localPortInt))
	if err != nil {
		cmd = "N/A"
	}

	if h.isExceptionApp(cmd) {
		bindAddr, _ := net.ResolveUDPAddr("udp", h.sendThrough.String())
		pc, err := net.ListenUDP("udp", bindAddr)
		if err != nil {
			return err
		}

		h.exceptionConnMap.Store(conn, pc)

		go h.handleInput(conn, pc)

		log.Access(cmd, "direct", target.Network(), conn.LocalAddr().String(), target.String())
		return nil
	} else {
		return h.proxyHandler.Connect(conn, target)
	}
}

func (h *udpHandler) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	if pc, ok := h.exceptionConnMap.Load(conn); ok {
		_, err := pc.(*net.UDPConn).WriteTo(data, addr)
		if err != nil {
			return err
		}
		return nil
	} else {
		return h.proxyHandler.ReceiveTo(conn, data, addr)
	}
}

func (h *udpHandler) Close(conn core.UDPConn) {
	conn.Close()

	if pc, ok := h.exceptionConnMap.Load(conn); ok {
		pc.(*net.UDPConn).Close()
		h.exceptionConnMap.Delete(conn)
	}
}
