package proxy

import (
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/common/lsof"
	"github.com/xjasonlyu/tun2socks/common/pool"
	"github.com/xjasonlyu/tun2socks/component/dns"
	"github.com/xjasonlyu/tun2socks/component/stats"
	C "github.com/xjasonlyu/tun2socks/constant"
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/log"
)

type tcpHandler struct {
	proxyHost string
	proxyPort int

	fakeDNS       dns.FakeDNS
	sessionStater stats.SessionStater
}

func NewTCPHandler(proxyHost string, proxyPort int, fakeDNS dns.FakeDNS, sessionStater stats.SessionStater) core.TCPConnHandler {
	return &tcpHandler{
		proxyHost:     proxyHost,
		proxyPort:     proxyPort,
		fakeDNS:       fakeDNS,
		sessionStater: sessionStater,
	}
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

	// Up Link
	go func() {
		buf := pool.BufPool.Get().([]byte)
		defer pool.BufPool.Put(buf[:cap(buf)])
		if _, err := io.CopyBuffer(remoteConn, localConn, buf); err != nil {
			closeOnce()
		} else {
			localConn.SetDeadline(time.Now())
			remoteConn.SetDeadline(time.Now())
			tcpCloseRead(remoteConn)
		}
		wg.Done()
	}()

	// Down Link
	buf := pool.BufPool.Get().([]byte)
	if _, err := io.CopyBuffer(localConn, remoteConn, buf); err != nil {
		closeOnce()
	} else {
		localConn.SetDeadline(time.Now())
		remoteConn.SetDeadline(time.Now())
		tcpCloseRead(localConn)
	}
	pool.BufPool.Put(buf[:cap(buf)])

	wg.Wait() // Wait for Up Link done

	// Remove session
	if h.sessionStater != nil {
		h.sessionStater.RemoveSession(localConn)
	}
}

func (h *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	// Alias
	var localConn = conn

	// Lookup fakeDNS host record
	targetHost, err := lookupHost(h.fakeDNS, target)
	if err != nil {
		log.Warnf("lookup target host error: %v", err)
		return err
	}

	proxyAddr := net.JoinHostPort(h.proxyHost, strconv.Itoa(h.proxyPort))
	targetAddr := net.JoinHostPort(targetHost, strconv.Itoa(target.Port))
	// Dial
	remoteConn, err := dial(proxyAddr, targetAddr)
	if err != nil {
		log.Infof("Dial: %v", err)
		return err
	}

	// Get name of the process
	var process = lsof.GetProcessName(localConn.LocalAddr())
	if h.sessionStater != nil {
		sess := &C.Session{
			Process:       process,
			Network:       localConn.LocalAddr().Network(),
			DialerAddr:    remoteConn.LocalAddr().String(),
			ClientAddr:    localConn.LocalAddr().String(),
			TargetAddr:    targetAddr,
			UploadBytes:   0,
			DownloadBytes: 0,
			SessionStart:  time.Now(),
		}
		h.sessionStater.AddSession(localConn, sess)

		remoteConn = C.NewSessionConn(remoteConn, sess)
	}

	// Set keepalive
	tcpKeepAlive(localConn)
	tcpKeepAlive(remoteConn)

	// Relay connections
	go h.relay(localConn, remoteConn)

	log.Access(process, "proxy", "tcp", localConn.LocalAddr().String(), targetAddr)
	return nil
}
