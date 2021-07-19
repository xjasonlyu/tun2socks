package tunnel

import (
	"errors"
	"net"
	"os"
	"time"

	"github.com/xjasonlyu/tun2socks/common/pool"
	"github.com/xjasonlyu/tun2socks/component/nat"
	M "github.com/xjasonlyu/tun2socks/constant"
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/log"
	"github.com/xjasonlyu/tun2socks/proxy"
	"github.com/xjasonlyu/tun2socks/tunnel/statistic"
)

var (
	// _natTable uses source udp packet information
	// as key to store destination udp packetConn.
	_natTable = nat.NewTable()

	// _udpSessionTimeout is the default timeout for
	// each UDP session.
	_udpSessionTimeout = 60 * time.Second
)

func SetUDPTimeout(v int) {
	_udpSessionTimeout = time.Duration(v) * time.Second
}

func newUDPTracker(conn net.PacketConn, metadata *M.Metadata) net.PacketConn {
	return statistic.NewUDPTracker(conn, metadata, statistic.DefaultManager)
}

func handleUDP(packet core.UDPPacket) {
	id := packet.ID()
	metadata := &M.Metadata{
		Net:     M.UDP,
		SrcIP:   net.IP(id.RemoteAddress),
		SrcPort: id.RemotePort,
		DstIP:   net.IP(id.LocalAddress),
		DstPort: id.LocalPort,
	}

	generateNATKey := func(m *M.Metadata) string {
		return m.SourceAddress() /* as Full Cone NAT Key */
	}
	key := generateNATKey(metadata)

	handle := func(drop bool) bool {
		pc := _natTable.Get(key)
		if pc != nil {
			handleUDPToRemote(packet, pc, metadata /* as net.Addr */, drop)
			return true
		}
		return false
	}

	if handle(true /* drop */) {
		return
	}

	lockKey := key + "-lock"
	cond, loaded := _natTable.GetOrCreateLock(lockKey)
	go func() {
		if loaded {
			cond.L.Lock()
			cond.Wait()
			handle(true) /* drop after sending data to remote */
			cond.L.Unlock()
			return
		}

		defer func() {
			_natTable.Delete(lockKey)
			cond.Broadcast()
		}()

		pc, err := proxy.DialUDP(metadata)
		if err != nil {
			log.Warnf("[UDP] dial %s error: %v", metadata.DestinationAddress(), err)
			return
		}

		if dialerAddr, ok := pc.LocalAddr().(*net.UDPAddr); ok {
			metadata.MidIP = dialerAddr.IP
			metadata.MidPort = uint16(dialerAddr.Port)
		} else { /* fallback */
			metadata.MidIP, metadata.MidPort = parseAddr(pc.LocalAddr().String())
		}

		pc = newUDPTracker(pc, metadata)

		go func() {
			defer pc.Close()
			defer packet.Drop()
			defer _natTable.Delete(key)

			handleUDPToLocal(packet, pc)
		}()

		_natTable.Set(key, pc)
		handle(false /* drop */)
	}()
}

func handleUDPToRemote(packet core.UDPPacket, pc net.PacketConn, remote net.Addr, drop bool) {
	defer func() {
		if drop {
			packet.Drop()
		}
	}()

	if _, err := pc.WriteTo(packet.Data() /* data */, remote); err != nil {
		log.Warnf("[UDP] write to %s error: %v", remote, err)
	}
	pc.SetReadDeadline(time.Now().Add(_udpSessionTimeout)) /* reset timeout */

	log.Infof("[UDP] %s --> %s", packet.RemoteAddr(), remote)
}

func handleUDPToLocal(packet core.UDPPacket, pc net.PacketConn) {
	buf := pool.Get(pool.MaxSegmentSize)
	defer pool.Put(buf)

	for /* just loop */ {
		pc.SetReadDeadline(time.Now().Add(_udpSessionTimeout))
		n, from, err := pc.ReadFrom(buf)
		if err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) /* ignore i/o timeout */ {
				log.Warnf("[UDP] read error: %v", err)
			}
			return
		}

		if _, err := packet.WriteBack(buf[:n], from); err != nil {
			log.Warnf("[UDP] write back from %s error: %v", from, err)
			return
		}

		log.Infof("[UDP] %s <-- %s", packet.RemoteAddr(), from)
	}
}
