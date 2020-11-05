package core

import (
	"fmt"
	"net"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"

	"github.com/xjasonlyu/tun2socks/internal/adapter"
)

const udpNoChecksum = true

type udpHandleFunc func(adapter.UDPPacket)

func WithUDPHandler(handle udpHandleFunc) Option {
	return func(s *stack.Stack) error {
		udpHandlePacket := func(r *stack.Route, id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
			// Ref: gVisor pkg/tcpip/transport/udp/endpoint.go HandlePacket
			hdr := header.UDP(pkt.TransportHeader().View())
			if int(hdr.Length()) > pkt.Data.Size()+header.UDPMinimumSize {
				// Malformed packet.
				s.Stats().UDP.MalformedPacketsReceived.Increment()
				return true
			}

			if !verifyChecksum(r, hdr, pkt) {
				// Checksum error.
				s.Stats().UDP.ChecksumErrors.Increment()
				return true
			}

			s.Stats().UDP.PacketsReceived.Increment()

			// make a clone here.
			route := r.Clone()
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
		return sendUDP(p.r, data, p.id.LocalPort, p.id.RemotePort, udpNoChecksum)
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
	return sendUDP(&r, data, uint16(udpAddr.Port), p.id.RemotePort, udpNoChecksum)
}

// sendUDP sends a UDP segment via the provided network endpoint and under the
// provided identity.
func sendUDP(r *stack.Route, data buffer.VectorisedView, localPort, remotePort uint16, noChecksum bool) (int, error) {
	// Allocate a buffer for the UDP header.
	pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
		ReserveHeaderBytes: header.UDPMinimumSize + int(r.MaxHeaderLength()),
		Data:               data,
	})

	// Initialize the UDP header.
	udpHdr := header.UDP(pkt.TransportHeader().Push(header.UDPMinimumSize))
	pkt.TransportProtocolNumber = udp.ProtocolNumber

	length := uint16(pkt.Size())
	udpHdr.Encode(&header.UDPFields{
		SrcPort: localPort,
		DstPort: remotePort,
		Length:  length,
	})

	// Set the checksum field unless TX checksum offload is enabled.
	// On IPv4, UDP checksum is optional, and a zero value indicates the
	// transmitter skipped the checksum generation (RFC768).
	// On IPv6, UDP checksum is not optional (RFC2460 Section 8.1).
	if r.Capabilities()&stack.CapabilityTXChecksumOffload == 0 &&
		(!noChecksum || r.NetProto == header.IPv6ProtocolNumber) {
		xsum := r.PseudoHeaderChecksum(udp.ProtocolNumber, length)
		for _, v := range data.Views() {
			xsum = header.Checksum(v, xsum)
		}
		udpHdr.SetChecksum(^udpHdr.CalculateChecksum(xsum))
	}

	ttl := r.DefaultTTL()
	if err := r.WritePacket(nil /* gso */, stack.NetworkHeaderParams{
		Protocol: udp.ProtocolNumber,
		TTL:      ttl,
		TOS:      0, /* default */
	}, pkt); err != nil {
		r.Stats().UDP.PacketSendErrors.Increment()
		return 0, fmt.Errorf("%s", err)
	}

	// Track count of packets sent.
	r.Stats().UDP.PacketsSent.Increment()
	return data.Size(), nil
}

// verifyChecksum verifies the checksum unless RX checksum offload is enabled.
// On IPv4, UDP checksum is optional, and a zero value means the transmitter
// omitted the checksum generation (RFC768).
// On IPv6, UDP checksum is not optional (RFC2460 Section 8.1).
func verifyChecksum(r *stack.Route, hdr header.UDP, pkt *stack.PacketBuffer) bool {
	if r.Capabilities()&stack.CapabilityRXChecksumOffload == 0 &&
		(hdr.Checksum() != 0 || r.NetProto == header.IPv6ProtocolNumber) {
		xsum := r.PseudoHeaderChecksum(udp.ProtocolNumber, hdr.Length())
		for _, v := range pkt.Data.Views() {
			xsum = header.Checksum(v, xsum)
		}
		return hdr.CalculateChecksum(xsum) == 0xffff
	}
	return true
}
