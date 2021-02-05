package tunnel

import (
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/common/adapter"
	"github.com/xjasonlyu/tun2socks/common/pool"
	"github.com/xjasonlyu/tun2socks/log"
	"github.com/xjasonlyu/tun2socks/proxy"
)

const (
	tcpWaitTimeout = 5 * time.Second
)

func handleTCP(localConn adapter.TCPConn) {
	defer localConn.Close()

	metadata := localConn.Metadata()
	if !metadata.Valid() {
		log.Warnf("[Metadata] not valid: %#v", metadata)
		return
	}

	targetConn, err := proxy.Dial(metadata)
	if err != nil {
		log.Warnf("[TCP] dial %s error: %v", metadata.DestinationAddress(), err)
		return
	}

	if dialerAddr, ok := targetConn.LocalAddr().(*net.TCPAddr); ok {
		metadata.MidIP = dialerAddr.IP
		metadata.MidPort = uint16(dialerAddr.Port)
	} else { /* fallback */
		ip, p, _ := net.SplitHostPort(targetConn.LocalAddr().String())
		port, _ := strconv.ParseUint(p, 10, 16)
		metadata.MidIP = net.ParseIP(ip)
		metadata.MidPort = uint16(port)
	}

	targetConn = newTCPTracker(targetConn, metadata)
	defer targetConn.Close()

	log.Infof("[TCP] %s <-> %s", metadata.SourceAddress(), metadata.DestinationAddress())
	relay(localConn, targetConn) /* relay connections */
}

// relay copies between left and right bidirectionally.
func relay(left, right net.Conn) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = copyBuffer(right, left) /* ignore error */
		right.SetReadDeadline(time.Now().Add(tcpWaitTimeout))
	}()

	go func() {
		defer wg.Done()
		_ = copyBuffer(left, right) /* ignore error */
		left.SetReadDeadline(time.Now().Add(tcpWaitTimeout))
	}()

	wg.Wait()
}

func copyBuffer(dst io.Writer, src io.Reader) error {
	buf := pool.Get(pool.RelayBufferSize)
	defer pool.Put(buf)

	_, err := io.CopyBuffer(dst, src, buf)
	return err
}
