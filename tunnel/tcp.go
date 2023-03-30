package tunnel

import (
	"errors"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/common/pool"
	"github.com/xjasonlyu/tun2socks/v2/core/adapter"
	"github.com/xjasonlyu/tun2socks/v2/log"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
)

var _tcpWaitTimeout = 5 * time.Second

func SetTCPWaitTimeout(t time.Duration) {
	_tcpWaitTimeout = t
}

func handleTCPConn(localConn adapter.TCPConn) {
	defer localConn.Close()

	id := localConn.ID()
	metadata := &M.Metadata{
		Network: M.TCP,
		SrcIP:   net.IP(id.RemoteAddress),
		SrcPort: id.RemotePort,
		DstIP:   net.IP(id.LocalAddress),
		DstPort: id.LocalPort,
	}

	targetConn, err := proxy.Dial(metadata)
	if err != nil {
		log.Warnf("[TCP] dial %s: %v", metadata.DestinationAddress(), err)
		return
	}
	metadata.MidIP, metadata.MidPort = parseAddr(targetConn.LocalAddr())

	targetConn = statistic.DefaultTCPTracker(targetConn, metadata)
	defer targetConn.Close()

	log.Infof("[TCP] %s <-> %s", metadata.SourceAddress(), metadata.DestinationAddress())
	if err = relay(localConn, targetConn); err != nil {
		log.Debugf("[TCP] %s <-> %s: %v", metadata.SourceAddress(), metadata.DestinationAddress(), err)
	}
}

// relay copies between left and right bidirectionally.
func relay(left, right net.Conn) error {
	wg := sync.WaitGroup{}
	wg.Add(2)

	var leftErr, rightErr error

	go func() {
		defer wg.Done()
		if err := copyBuffer(right, left); err != nil {
			leftErr = errors.Join(leftErr, err)
		}

		cwOk := false
		if cw, ok := right.(statistic.CloseWriter); ok {
			if err := cw.CloseWrite(); err != nil {
				if !isIgnorable(err) {
					leftErr = errors.Join(leftErr, err)
				}
			} else {
				cwOk = true
			}
		}
		if !cwOk {
			right.SetReadDeadline(time.Now().Add(_tcpWaitTimeout))
		}
	}()

	go func() {
		defer wg.Done()
		if err := copyBuffer(left, right); err != nil {
			rightErr = errors.Join(rightErr, err)
		}

		cwOk := false
		if cw, ok := left.(statistic.CloseWriter); ok {
			if err := cw.CloseWrite(); err != nil {
				if !isIgnorable(err) {
					rightErr = errors.Join(rightErr, err)
				}
			} else {
				cwOk = true
			}
		}
		if !cwOk {
			left.SetReadDeadline(time.Now().Add(_tcpWaitTimeout))
		}
	}()

	wg.Wait()
	return errors.Join(leftErr, rightErr)
}

func copyBuffer(dst io.Writer, src io.Reader) error {
	buf := pool.Get(pool.RelayBufferSize)
	defer pool.Put(buf)

	_, err := io.CopyBuffer(dst, src, buf)
	if err != nil && !isIgnorable(err) {
		return err
	}
	return nil
}

func isIgnorable(err error) bool {
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return true /* ignore I/O timeout */
	} else if errors.Is(err, syscall.EPIPE) {
		return true /* ignore broken pipe */
	} else if errors.Is(err, syscall.ECONNRESET) {
		return true /* ignore connection reset by peer */
	}
	return false
}
