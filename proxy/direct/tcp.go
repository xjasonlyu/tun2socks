package direct

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/pool"
	"github.com/xjasonlyu/tun2socks/core"
)

type tcpHandler struct{}

func NewTCPHandler() core.TCPConnHandler {
	return &tcpHandler{}
}

func (h *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	c, err := net.DialTCP("tcp", nil, target)
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
