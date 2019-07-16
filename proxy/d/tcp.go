package d

import (
	"io"
	"net"
	"strconv"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/lsof"
	"github.com/xjasonlyu/tun2socks/core"
)

// This handler allows you chain another proxy behind tun2socks locally, typically a rule-based proxy client, e.g. V2Ray.
//
// Rule-based proxy clients are very useful, they are able to dispatch requests to different servers based on powerful rule filters.
// By using this setup, you are able to make all your TCP/UDP traffic under control with your favorite rule-based proxy client.
//
// Here's an example setup on macOS:
//
// tun2socks -tunGw 10.255.0.1 -fakeDns -proxyType d -proxyServer 127.0.0.1:1086 -exceptionSendThrough 192.168.1.189:0 -exceptionApps "v2ray"
//
// route delete default
// route add default 10.255.0.1
// route add default 192.168.1.1 -ifscope en0
//
// Where 192.168.1.189 is the default interface address, in my case, it's the WiFi interface and it's en0.
// 192.168.1.1 is the default gateway.
// It's very important to have two default routes, and the default route to TUN should has the highest priority.
//
// Start v2ray (or any other chainable proxy clients) and has SOCKS inbound listen on 127.0.0.1:1086.
//
// Optinally with all outbounds have sendThrough set to 192.168.1.189, if applicable.
// https://v2ray.com/chapter_02/01_overview.html#outboundobject

type tcpHandler struct {
	proxyHandler  core.TCPConnHandler
	exceptionApps []string
	sendThrough   net.Addr
}

func NewTCPHandler(proxyHandler core.TCPConnHandler, exceptionApps []string, sendThrough net.Addr) core.TCPConnHandler {
	return &tcpHandler{
		proxyHandler,
		exceptionApps,
		sendThrough,
	}
}

func (h *tcpHandler) isExceptionApp(name string) bool {
	for _, app := range h.exceptionApps {
		if name == app {
			return true
		}
	}
	return false
}

func (h *tcpHandler) relay(lhs, rhs net.Conn) {
	cls := func() {
		rhs.Close()
		lhs.Close()
	}

	go func() {
		io.Copy(rhs, lhs)
		cls()
	}()

	io.Copy(lhs, rhs)
	cls()
}

func (h *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	localHost, localPortStr, _ := net.SplitHostPort(conn.LocalAddr().String())
	localPortInt, _ := strconv.Atoi(localPortStr)
	cmd, err := lsof.GetCommandNameBySocket("tcp", localHost, uint16(localPortInt))
	if err != nil {
		cmd = "unknown process"
	}

	if h.isExceptionApp(cmd) {
		dialer := net.Dialer{LocalAddr: h.sendThrough}
		rc, err := dialer.Dial("tcp", target.String())
		if err != nil {
			return err
		}

		go h.relay(conn, rc)

		log.Access(cmd, "direct", target.Network(), conn.LocalAddr().String(), target.String())

		return nil
	} else {
		return h.proxyHandler.Handle(conn, target)
	}
}
