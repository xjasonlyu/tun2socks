package utils

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/common/pool"
)

type duplexConn interface {
	net.Conn
	CloseRead() error
	CloseWrite() error
}

func TCPCloseRead(conn net.Conn) {
	if c, ok := conn.(duplexConn); ok {
		c.CloseRead()
	}
}

func TCPCloseWrite(conn net.Conn) {
	if c, ok := conn.(duplexConn); ok {
		c.CloseWrite()
	}
}

func TCPKeepAlive(conn net.Conn) {
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(30 * time.Second)
	}
}

func TCPRelay(localConn, remoteConn net.Conn) {
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
			TCPCloseRead(remoteConn)
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
		TCPCloseRead(localConn)
	}
	pool.BufPool.Put(buf[:cap(buf)])

	wg.Wait() // Wait for Up Link done
}
