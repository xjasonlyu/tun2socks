package exception

import (
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/lsof"
	"github.com/xjasonlyu/tun2socks/common/pool"
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

type tcpHandler struct {
	exceptionApps []string
	sendThrough   net.Addr
	proxyHandler  core.TCPConnHandler
}

func NewTCPHandler(proxyHandler core.TCPConnHandler, exceptionApps []string, sendThrough net.Addr) core.TCPConnHandler {
	return &tcpHandler{
		exceptionApps: exceptionApps,
		sendThrough:   sendThrough,
		proxyHandler:  proxyHandler,
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

func (h *tcpHandler) relay(localConn, remoteConn net.Conn) {
	var once sync.Once
	closeOnce := func() {
		once.Do(func() {
			localConn.Close()
			remoteConn.Close()
		})
	}

	// Close
	defer closeOnce()

	// WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		buf := pool.BufPool.Get().([]byte)
		defer pool.BufPool.Put(buf[:cap(buf)])
		if _, err := io.CopyBuffer(remoteConn, localConn, buf); err != nil {
			closeOnce()
		} else {
			localConn.SetDeadline(time.Now())
			remoteConn.SetDeadline(time.Now())
		}
		wg.Done()
	}()

	buf := pool.BufPool.Get().([]byte)
	if _, err := io.CopyBuffer(localConn, remoteConn, buf); err != nil {
		closeOnce()
	} else {
		localConn.SetDeadline(time.Now())
		remoteConn.SetDeadline(time.Now())
	}
	pool.BufPool.Put(buf[:cap(buf)])

	wg.Wait() // Wait for Up Link done
}

func (h *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	localHost, localPortStr, _ := net.SplitHostPort(conn.LocalAddr().String())
	localPortInt, _ := strconv.Atoi(localPortStr)
	cmd, err := lsof.GetCommandNameBySocket("tcp", localHost, uint16(localPortInt))
	if err != nil {
		cmd = "N/A"
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
