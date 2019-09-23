package proxy

import (
	"io"
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

type tcpHandler struct {
	proxyHost string
	proxyPort int
}

func NewTCPHandler(proxyHost string, proxyPort int) core.TCPConnHandler {
	return &tcpHandler{
		proxyHost: proxyHost,
		proxyPort: proxyPort,
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

	// Cleanup
	defer func() {
		// Close
		closeOnce()
		// Remove session
		removeSession(localConn)
	}()

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
}

func (h *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	// Alias
	var localConn = conn

	// Lookup fakeDNS host record
	targetHost, err := lookupHost(target)
	if err != nil {
		log.Warnf("lookup target host: %v", err)
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

	var process = "N/A"
	if monitor != nil {
		// Get name of the process
		process = lsof.GetProcessName(localConn.LocalAddr())
		session := &S.Session{
			Process:       process,
			Network:       localConn.LocalAddr().Network(),
			DialerAddr:    remoteConn.LocalAddr().String(),
			ClientAddr:    localConn.LocalAddr().String(),
			TargetAddr:    targetAddr,
			UploadBytes:   0,
			DownloadBytes: 0,
			SessionStart:  time.Now(),
		}
		addSession(localConn, session)
		remoteConn = &S.Conn{Session: session, Conn: remoteConn}
	}

	// Set keepalive
	tcpKeepAlive(localConn)
	tcpKeepAlive(remoteConn)

	// Relay connections
	go h.relay(localConn, remoteConn)

	log.Access(process, "proxy", "tcp", localConn.LocalAddr().String(), targetAddr)
	return nil
}
