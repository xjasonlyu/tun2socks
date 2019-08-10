package socks

import (
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/proxy"

	"github.com/xjasonlyu/tun2socks/common/dns"
	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/lsof"
	"github.com/xjasonlyu/tun2socks/common/stats"
	"github.com/xjasonlyu/tun2socks/common/stats/session"
	"github.com/xjasonlyu/tun2socks/core"
)

var activeTCPConnections int64

func init() {
	session.ActiveTCPConnections = &activeTCPConnections
}

type tcpHandler struct {
	sync.Mutex

	proxyHost string
	proxyPort uint16

	fakeDns       dns.FakeDns
	sessionStater stats.SessionStater
}

func NewTCPHandler(proxyHost string, proxyPort uint16, fakeDns dns.FakeDns, sessionStater stats.SessionStater) core.TCPConnHandler {
	return &tcpHandler{
		proxyHost:     proxyHost,
		proxyPort:     proxyPort,
		fakeDns:       fakeDns,
		sessionStater: sessionStater,
	}
}

func (h *tcpHandler) relay(localConn, remoteConn net.Conn, sess *stats.Session) {
	upCh := make(chan struct{})

	// Close
	defer func() {
		localConn.Close()
		remoteConn.Close()
	}()

	// UpLink
	go func() {
		io.Copy(remoteConn, localConn)
		remoteConn.SetReadDeadline(time.Now())
		upCh <- struct{}{}
	}()

	// DownLink
	io.Copy(localConn, remoteConn)
	localConn.SetReadDeadline(time.Now())

	<-upCh // Wait for UpLink done.

	// add -1
	atomic.AddInt64(&activeTCPConnections, -1)

	if h.sessionStater != nil {
		h.sessionStater.RemoveSession(localConn)
	}
}

func (h *tcpHandler) Handle(localConn net.Conn, target *net.TCPAddr) error {
	dialer, err := proxy.SOCKS5("tcp", core.ParseTCPAddr(h.proxyHost, h.proxyPort).String(), nil, nil)
	if err != nil {
		return err
	}

	// Replace with a domain name if target address IP is a fake IP.
	var targetHost = target.IP.String()
	if h.fakeDns != nil {
		if host, exist := h.fakeDns.IPToHost(target.IP); exist {
			targetHost = host
		}
	}

	targetAddr := net.JoinHostPort(targetHost, strconv.Itoa(target.Port))
	remoteConn, err := dialer.Dial(target.Network(), targetAddr)
	if err != nil {
		return err
	}

	var process string
	var sess *stats.Session
	if h.sessionStater != nil {
		// Get name of the process.
		localHost, localPortStr, _ := net.SplitHostPort(localConn.LocalAddr().String())
		localPortInt, _ := strconv.Atoi(localPortStr)
		process, err = lsof.GetCommandNameBySocket(target.Network(), localHost, uint16(localPortInt))
		if err != nil {
			process = "N/A"
		}

		sess = &stats.Session{
			ProcessName:   process,
			Network:       target.Network(),
			DialerAddr:    remoteConn.LocalAddr().String(),
			ClientAddr:    localConn.LocalAddr().String(),
			TargetAddr:    targetAddr,
			UploadBytes:   0,
			DownloadBytes: 0,
			SessionStart:  time.Now(),
		}
		h.sessionStater.AddSession(localConn, sess)

		remoteConn = stats.NewSessionConn(remoteConn, sess)
	}

	// set keepalive
	tcpKeepAlive(localConn)
	tcpKeepAlive(remoteConn)

	// relay connections
	go h.relay(localConn, remoteConn, sess)

	// add 1
	atomic.AddInt64(&activeTCPConnections, 1)

	log.Access(process, "proxy", target.Network(), localConn.LocalAddr().String(), targetAddr)
	return nil
}

func tcpKeepAlive(conn net.Conn) {
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(30 * time.Second)
	}
}
