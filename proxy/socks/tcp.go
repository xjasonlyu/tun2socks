package socks

import (
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/proxy"

	"github.com/xjasonlyu/tun2socks/common/dns"
	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/lsof"
	"github.com/xjasonlyu/tun2socks/common/stats"
	"github.com/xjasonlyu/tun2socks/core"
)

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

type direction byte

const (
	dirUplink direction = iota
	dirDownlink
)

func statsCopy(dst io.Writer, src io.Reader, sess *stats.Session, dir direction) (written int64, err error) {
	buf := make([]byte, 64*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				switch dir {
				case dirUplink:
					sess.AddUploadBytes(int64(nw))
				case dirDownlink:
					sess.AddDownloadBytes(int64(nw))
				default:
				}
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return
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
		if h.sessionStater != nil && sess != nil {
			statsCopy(remoteConn, localConn, sess, dirUplink)
		} else {
			io.Copy(remoteConn, localConn)
		}
		remoteConn.SetReadDeadline(time.Now())
		upCh <- struct{}{}
	}()

	// DownLink
	if h.sessionStater != nil && sess != nil {
		statsCopy(localConn, remoteConn, sess, dirDownlink)
	} else {
		io.Copy(localConn, remoteConn)
	}
	localConn.SetReadDeadline(time.Now())

	<-upCh // Wait for UpLink done.

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
	}

	// set keepalive
	tcpKeepAlive(localConn)
	tcpKeepAlive(remoteConn)

	// relay connections
	go h.relay(localConn, remoteConn, sess)

	log.Access(process, "proxy", target.Network(), localConn.LocalAddr().String(), targetAddr)
	return nil
}

func tcpKeepAlive(conn net.Conn) {
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(30 * time.Second)
	}
}
