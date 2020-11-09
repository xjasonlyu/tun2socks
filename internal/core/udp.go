package core

import (
	"fmt"
	"net"
	_ "unsafe"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"

	"github.com/xjasonlyu/tun2socks/internal/adapter"
	"github.com/xjasonlyu/tun2socks/pkg/log"
)

const udpNoChecksum = true

type udpHandleFunc func(adapter.UDPPacket)

func WithUDPHandler(handle udpHandleFunc) Option {
	return func(s *stack.Stack) error {
		udpHandlePacket := func(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
			// Ref: gVisor pkg/tcpip/transport/udp/endpoint.go HandlePacket
			udpHdr := header.UDP(pkt.TransportHeader().View())
			if int(udpHdr.Length()) > pkt.Data.Size()+header.UDPMinimumSize {
				// Malformed packet.
				s.Stats().UDP.MalformedPacketsReceived.Increment()
				return true
			}

			if !verifyChecksum(udpHdr, pkt) {
				// Checksum error.
				s.Stats().UDP.ChecksumErrors.Increment()
				return true
			}

			s.Stats().UDP.PacketsReceived.Increment()

			netHdr := pkt.Network()
			route, err := s.FindRoute(pkt.NICID, netHdr.DestinationAddress(), netHdr.SourceAddress(), pkt.NetworkProtocolNumber, false /* multicastLoop */)
			if err != nil {
				log.Warnf("[STACK] find route error: %v", err)
				return true
			}
			route.ResolveWith(pkt.SourceLinkAddress())

			packet := &udpPacket{
				id: id,
				r:  &route,
				metadata: &adapter.Metadata{
					Net:     adapter.UDP,
					SrcIP:   net.IP(id.RemoteAddress),
					SrcPort: id.RemotePort,
					DstIP:   net.IP(id.LocalAddress),
					DstPort: id.LocalPort,
				},
				payload: pkt.Data.ToView(),
			}

			handle(packet)
			return true
		}
		s.SetTransportProtocolHandler(udp.ProtocolNumber, udpHandlePacket)
		return nil
	}
}

type udpPacket struct {
	id       stack.TransportEndpointID
	r        *stack.Route
	metadata *adapter.Metadata
	payload  []byte
}

func (p *udpPacket) Data() []byte {
	return p.payload
}

func (p *udpPacket) Drop() {
	p.r.Release()
}

func (p *udpPacket) LocalAddr() net.Addr {
	return &net.UDPAddr{IP: net.IP(p.id.LocalAddress), Port: int(p.id.LocalPort)}
}

func (p *udpPacket) Metadata() *adapter.Metadata {
	return p.metadata
}

func (p *udpPacket) RemoteAddr() net.Addr {
	return &net.UDPAddr{IP: net.IP(p.id.RemoteAddress), Port: int(p.id.RemotePort)}
}

func (p *udpPacket) WriteBack(b []byte, addr net.Addr) (int, error) {
	v := buffer.View(b)
	if len(v) > header.UDPMaximumPacketSize {
		// Payload can't possibly fit in a packet.
		return 0, fmt.Errorf("%s", tcpip.ErrMessageTooLong)
	}

	data := v.ToVectorisedView()
	// if addr is not provided, write back use original dst Addr as src Addr.
	if addr == nil {
		return _sendUDP(p.r, data, p.id.LocalPort, p.id.RemotePort, udpNoChecksum)
	}

	udpAddr, ok := addr.(*net.UDPAddr)
	if !ok {
		return 0, fmt.Errorf("type %T is not a valid udp address", addr)
	}

	r := p.r.Clone()
	defer r.Release()

	if ipv4 := udpAddr.IP.To4(); ipv4 != nil {
		r.LocalAddress = tcpip.Address(ipv4)
	} else {
		r.LocalAddress = tcpip.Address(udpAddr.IP)
	}
	return _sendUDP(&r, data, uint16(udpAddr.Port), p.id.RemotePort, udpNoChecksum)
}

// _sendUDP wraps sendUDP with some default parameters.
func _sendUDP(r *stack.Route, data buffer.VectorisedView, localPort, remotePort uint16, noChecksum bool) (int, error) {
	if err := sendUDP(r, data, localPort, remotePort, 0 /* ttl */, true /* useDefaultTTL */, 0 /* tos */, nil /* owner */, noChecksum); err != nil {
		return 0, fmt.Errorf("%s", err)
	}
	return data.Size(), nil
}

// sendUDP sends a UDP segment via the provided network endpoint and under the
// provided identity.
//
//go:linkname sendUDP gvisor.dev/gvisor/pkg/tcpip/transport/udp.sendUDP
func sendUDP(r *stack.Route, data buffer.VectorisedView, localPort, remotePort uint16, ttl uint8, useDefaultTTL bool, tos uint8, owner tcpip.PacketOwner, noChecksum bool) *tcpip.Error

// verifyChecksum verifies the checksum unless RX checksum offload is enabled.
// On IPv4, UDP checksum is optional, and a zero value means the transmitter
// omitted the checksum generation (RFC768).
// On IPv6, UDP checksum is not optional (RFC2460 Section 8.1).
//
//go:linkname verifyChecksum gvisor.dev/gvisor/pkg/tcpip/transport/udp.verifyChecksum
func verifyChecksum(hdr header.UDP, pkt *stack.PacketBuffer) bool
