package tunnel

import (
	"errors"
	"net"
	"os"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/common/pool"
	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/log"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
)

// _udpSessionTimeout is the default timeout for each UDP session.
var _udpSessionTimeout = 60 * time.Second

func SetUDPTimeout(v int) {
	_udpSessionTimeout = time.Duration(v) * time.Second
}

func newUDPTracker(conn net.PacketConn, metadata *M.Metadata) net.PacketConn {
	return statistic.NewUDPTracker(conn, metadata, statistic.DefaultManager)
}

func handleUDPConn(uc core.UDPConn) {
	defer uc.Close()

	var (
		srcIP, srcPort = parseAddr(uc.RemoteAddr())
		dstIP, dstPort = parseAddr(uc.LocalAddr())
	)
	metadata := &M.Metadata{
		Network: M.UDP,
		SrcIP:   srcIP,
		SrcPort: srcPort,
		DstIP:   dstIP,
		DstPort: dstPort,
	}

	pc, err := proxy.DialUDP(metadata)
	if err != nil {
		log.Warnf("[UDP] dial %s error: %v", metadata.DestinationAddress(), err)
		return
	}
	metadata.MidIP, metadata.MidPort = parseAddr(pc.LocalAddr())

	pc = newUDPTracker(pc, metadata)
	defer pc.Close()

	remote := metadata.Addr()
	go handleUDPToRemote(uc, pc, remote)
	handleUDPToLocal(uc, pc, remote)
}

func handleUDPToRemote(uc core.UDPConn, pc net.PacketConn, remote net.Addr) {
	buf := pool.Get(pool.MaxSegmentSize)
	defer pool.Put(buf)

	for {
		n, err := uc.Read(buf)
		if err != nil {
			return
		}

		if _, err := pc.WriteTo(buf[:n], remote); err != nil {
			log.Warnf("[UDP] write to %s error: %v", remote, err)
		}

		log.Infof("[UDP] %s --> %s", uc.RemoteAddr(), remote)
	}
}

func handleUDPToLocal(uc core.UDPConn, pc net.PacketConn, remote net.Addr) {
	buf := pool.Get(pool.MaxSegmentSize)
	defer pool.Put(buf)

	for {
		pc.SetReadDeadline(time.Now().Add(_udpSessionTimeout))
		n, from, err := pc.ReadFrom(buf)
		if err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) /* ignore i/o timeout */ {
				log.Warnf("[UDP] read error: %v", err)
			}
			return
		}

		if from.Network() != remote.Network() || from.String() != remote.String() {
			log.Warnf("[UDP] drop unknown packet from %s", from)
			return
		}

		if _, err := uc.Write(buf[:n]); err != nil {
			log.Warnf("[UDP] write back from %s error: %v", from, err)
			return
		}
	}
}
