package tunnel

import (
	"errors"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/xjasonlyu/clash/common/pool"
	"github.com/xjasonlyu/clash/component/resolver"
	"github.com/xjasonlyu/tun2socks/internal/adapter"
	"github.com/xjasonlyu/tun2socks/internal/manager"
	"github.com/xjasonlyu/tun2socks/internal/proxy"
	"github.com/xjasonlyu/tun2socks/pkg/log"
	"github.com/xjasonlyu/tun2socks/pkg/nat"
)

const (
	udpTimeout    = 30 * time.Second
	udpBufferSize = (1 << 16) - 1 // largest possible UDP datagram
)

var (
	// natTable uses source udp packet information
	// as key to store destination udp packetConn.
	natTable = nat.NewTable()
)

func handleUDP(packet adapter.UDPPacket) {
	metadata := packet.Metadata()
	if !metadata.Valid() {
		log.Warnf("[Metadata] not valid: %#v", metadata)
		return
	}

	// make a fAddr if request ip is fake ip.
	var fAddr net.Addr
	if resolver.IsExistFakeIP(metadata.DstIP) {
		fAddr = metadata.UDPAddr()
	}

	err := resolveMetadata(metadata)
	if err != nil {
		log.Warnf("[Metadata] resolve metadata error: %v", err)
		return
	}

	key := generateNATKey(metadata)

	handle := func(drop bool) bool {
		pc := natTable.Get(key)
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
	cond, loaded := natTable.GetOrCreateLock(lockKey)
	go func() {
		if loaded {
			cond.L.Lock()
			cond.Wait()
			handle(true) /* drop after sending data to remote */
			cond.L.Unlock()
			return
		}

		defer func() {
			natTable.Delete(lockKey)
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
		} else {
			ip, p, _ := net.SplitHostPort(pc.LocalAddr().String())
			port, _ := strconv.ParseUint(p, 10, 16)
			metadata.MidIP = net.ParseIP(ip)
			metadata.MidPort = uint16(port)
		}

		pc = manager.NewUDPTracker(pc, metadata)

		go func() {
			defer pc.Close()
			defer packet.Drop()
			defer natTable.Delete(key)

			handleUDPToLocal(packet, pc, fAddr, udpTimeout)
		}()

		natTable.Set(key, pc)
		handle(false /* drop */)
	}()
}

func handleUDPToRemote(packet adapter.UDPPacket, pc net.PacketConn, remote net.Addr, drop bool) {
	defer func() {
		if drop {
			packet.Drop()
		}
	}()

	if _, err := pc.WriteTo(packet.Data() /* data */, remote); err != nil {
		log.Warnf("[UDP] write to %s error: %v", remote, err)
	}

	log.Infof("[UDP] %s --> %s", packet.RemoteAddr(), remote)
}

func handleUDPToLocal(packet adapter.UDPPacket, pc net.PacketConn, fAddr net.Addr, timeout time.Duration) {
	buf := pool.Get(udpBufferSize)
	defer pool.Put(buf)

	for /* just loop */ {
		pc.SetReadDeadline(time.Now().Add(timeout))
		n, from, err := pc.ReadFrom(buf)
		if err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) /* ignore i/o timeout */ {
				log.Warnf("[UDP] ReadFrom error: %v", err)
			}
			return
		}

		if fAddr != nil {
			from = fAddr
		}

		if _, err := packet.WriteBack(buf[:n], from); err != nil {
			log.Warnf("[UDP] write back from %s error: %v", from, err)
			return
		}

		log.Infof("[UDP] %s <-- %s", packet.RemoteAddr(), from)
	}
}
