package tunnel

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/common/pool"
	"github.com/xjasonlyu/tun2socks/v2/core/adapter"
	"github.com/xjasonlyu/tun2socks/v2/log"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
)

const (
	// tcpWaitTimeout implements a TCP half-close timeout.
	tcpWaitTimeout = 60 * time.Second
)

func handleTCPConn(originConn adapter.TCPConn) {
	defer originConn.Close()

	id := originConn.ID()
	metadata := &M.Metadata{
		Network: M.TCP,
		SrcIP:   net.IP(id.RemoteAddress),
		SrcPort: id.RemotePort,
		DstIP:   net.IP(id.LocalAddress),
		DstPort: id.LocalPort,
	}

	remoteConn, err := proxy.Dial(metadata)
	if err != nil {
		log.Warnf("[TCP] dial %s: %v", metadata.DestinationAddress(), err)
		return
	}
	metadata.MidIP, metadata.MidPort = parseAddr(remoteConn.LocalAddr())

	remoteConn = statistic.DefaultTCPTracker(remoteConn, metadata)
	defer remoteConn.Close()

	log.Infof("[TCP] %s <-> %s", metadata.SourceAddress(), metadata.DestinationAddress())
	if err = pipe(originConn, remoteConn); err != nil {
		log.Debugf("[TCP] %s <-> %s: %v", metadata.SourceAddress(), metadata.DestinationAddress(), err)
	}
}

// pipe copies copy data to & from provided net.Conn(s) bidirectionally.
func pipe(origin, remote net.Conn) error {
	wg := sync.WaitGroup{}
	wg.Add(2)

	var leftErr, rightErr error

	go func() {
		defer wg.Done()
		if err := copyBuffer(remote, origin); err != nil {
			leftErr = errors.Join(leftErr, err)
		}
		remote.SetReadDeadline(time.Now().Add(tcpWaitTimeout))
	}()

	go func() {
		defer wg.Done()
		if err := copyBuffer(origin, remote); err != nil {
			rightErr = errors.Join(rightErr, err)
		}
		origin.SetReadDeadline(time.Now().Add(tcpWaitTimeout))
	}()

	wg.Wait()
	return errors.Join(leftErr, rightErr)
}

func copyBuffer(dst io.Writer, src io.Reader) error {
	buf := pool.Get(pool.RelayBufferSize)
	defer pool.Put(buf)

	_, err := io.CopyBuffer(dst, src, buf)
	return err
}
