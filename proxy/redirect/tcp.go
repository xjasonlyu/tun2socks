package redirect

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/pool"
	"github.com/xjasonlyu/tun2socks/core"
)

// To do a benchmark using iperf3 locally, you may follow these steps:
//
// 1. Setup and configure the TUN device and start tun2socks with the
//    redirect handler using the following command:
//      tun2socks -proxyType redirect -proxyServer 127.0.0.1:1234
//    Tun2socks will redirect all traffic to 127.0.0.1:1234.
//
// 2. Route traffic targeting 1.2.3.4 to the TUN interface (240.0.0.1):
//      route add 1.2.3.4/32 240.0.0.1
//
// 3. Run iperf3 server locally and listening on 1234 port:
//      iperf3 -s -p 1234
//
// 4. Run iperf3 client locally and connect to 1.2.3.4:1234:
//      iperf3 -c 1.2.3.4 -p 1234
//
// It works this way:
// iperf3 client -> 1.2.3.4:1234 -> routing table -> TUN (240.0.0.1) -> tun2socks -> tun2socks redirect anything to 127.0.0.1:1234 -> iperf3 server
//
type tcpHandler struct {
	target string
}

type duplexConn interface {
	net.Conn
	CloseWrite() error
	CloseRead() error
}

func NewTCPHandler(target string) core.TCPConnHandler {
	return &tcpHandler{target: target}
}

func (h *tcpHandler) handleOutput(conn net.Conn, output io.WriteCloser) {
	defer func() {
		if tcpConn, ok := conn.(core.TCPConn); ok {
			tcpConn.CloseRead()
		} else {
			conn.Close()
		}
		if tcpOutput, ok := output.(duplexConn); ok {
			tcpOutput.CloseWrite()
		} else {
			output.Close()
		}
	}()

	buf := pool.BufPool.Get().([]byte)
	io.CopyBuffer(output, conn, buf)
	pool.BufPool.Put(buf[:cap(buf)])
}

func (h *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	c, err := net.Dial("tcp", h.target)
	if err != nil {
		return err
	}

	// WaitGroup
	var wg sync.WaitGroup
	wg.Add(2)

	var once sync.Once
	relayCopy := func(dst, src net.Conn) {
		closeOnce := func() {
			once.Do(func() {
				src.Close()
				dst.Close()
			})
		}

		// Close
		defer closeOnce()

		buf := pool.BufPool.Get().([]byte)
		defer pool.BufPool.Put(buf[:cap(buf)])
		if _, err := io.CopyBuffer(dst, src, buf); err != nil {
			closeOnce()
		} else {
			src.SetDeadline(time.Now())
			dst.SetDeadline(time.Now())
		}
		wg.Done()

		wg.Wait() // Wait for another goroutine
	}

	go relayCopy(conn, c)
	go relayCopy(c, conn)

	log.Infof("new proxy connection for target: %s:%s", target.Network(), target.String())
	return nil
}
