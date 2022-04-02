package tunnel

import (
	"errors"
	"net"
	"os"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/common/pool"
	"github.com/xjasonlyu/tun2socks/v2/core/adapter"
	"github.com/xjasonlyu/tun2socks/v2/log"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
)

// _udpSessionTimeout is the default timeout for each UDP session.
var _udpSessionTimeout = 60 * time.Second

func SetUDPTimeout(t time.Duration) {
	_udpSessionTimeout = t
}

func newUDPTracker(conn net.PacketConn, metadata *M.Metadata) net.PacketConn {
	return statistic.NewUDPTracker(conn, metadata, statistic.DefaultManager)
}

// TODO: Port Restricted NAT support.
func handleUDPConn(uc adapter.UDPConn) {
	defer uc.Close()

	id := uc.ID()
	metadata := &M.Metadata{
		Network: M.UDP,
		SrcIP:   net.IP(id.RemoteAddress),
		SrcPort: id.RemotePort,
		DstIP:   net.IP(id.LocalAddress),
		DstPort: id.LocalPort,
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

func handleUDPToRemote(uc adapter.UDPConn, pc net.PacketConn, remote net.Addr) {
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
		pc.SetReadDeadline(time.Now().Add(_udpSessionTimeout)) /* reset timeout */

		log.Infof("[UDP] %s --> %s", uc.RemoteAddr(), remote)
	}
}

func handleUDPToLocal(uc adapter.UDPConn, pc net.PacketConn, remote net.Addr) {
	buf := pool.Get(pool.MaxSegmentSize)
	defer pool.Put(buf)

	for {
		pc.SetReadDeadline(time.Now().Add(_udpSessionTimeout))
		n, from, err := pc.ReadFrom(buf)
		if err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) /* ignore I/O timeout */ {
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

		log.Infof("[UDP] %s <-- %s", uc.RemoteAddr(), from)
	}
}
