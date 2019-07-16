package d

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/lsof"
	"github.com/xjasonlyu/tun2socks/core"
)

type udpHandler struct {
	sync.Mutex

	proxyHandler   core.UDPConnHandler
	exceptionApps  []string
	sendThrough    net.Addr
	exceptionConns map[core.UDPConn]*net.UDPConn
	timeout        time.Duration
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
		proxyHandler:   proxyHandler,
		exceptionApps:  exceptionApps,
		sendThrough:    sendThrough,
		exceptionConns: make(map[core.UDPConn]*net.UDPConn),
		timeout:        timeout,
	}
}

func (h *udpHandler) handleInput(conn core.UDPConn, pc *net.UDPConn) {
	buf := core.NewBytes(core.BufSize)

	defer func() {
		h.Close(conn)
		core.FreeBytes(buf)
	}()

	for {
		pc.SetDeadline(time.Now().Add(h.timeout))
		n, addr, err := pc.ReadFromUDP(buf)
		if err != nil {
			return
		}

		_, err = conn.WriteFrom(buf[:n], addr)
		if err != nil {
			return
		}
	}
}

func (h *udpHandler) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	localHost, localPortStr, _ := net.SplitHostPort(conn.LocalAddr().String())
	localPortInt, _ := strconv.Atoi(localPortStr)
	cmd, err := lsof.GetCommandNameBySocket("udp", localHost, uint16(localPortInt))
	if err != nil {
		cmd = "unknown process"
	}

	if h.isExceptionApp(cmd) {
		bindAddr, _ := net.ResolveUDPAddr(
			"udp",
			h.sendThrough.String(),
		)
		pc, err := net.ListenUDP("udp", bindAddr)
		if err != nil {
			return err
		}
		h.Lock()
		h.exceptionConns[conn] = pc
		h.Unlock()

		go h.handleInput(conn, pc)

		log.Access(cmd, "direct", target.Network(), conn.LocalAddr().String(), target.String())

		return nil
	} else {
		return h.proxyHandler.Connect(conn, target)
	}
}

func (h *udpHandler) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	h.Lock()
	defer h.Unlock()

	if pc, found := h.exceptionConns[conn]; found {
		_, err := pc.WriteTo(data, addr)
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

	h.Lock()
	defer h.Unlock()

	if pc, ok := h.exceptionConns[conn]; ok {
		pc.Close()
		delete(h.exceptionConns, conn)
	}
}
