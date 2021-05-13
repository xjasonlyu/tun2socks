package stack

import (
	"fmt"
	"net"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

const (
	// udpNoChecksum disables UDP checksum.
	udpNoChecksum = true
)

func withUDPHandler() Option {
	return func(s *Stack) error {
		udpHandlePacket := func(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
			// Ref: gVisor pkg/tcpip/transport/udp/endpoint.go HandlePacket
			udpHdr := header.UDP(pkt.TransportHeader().View())
			if int(udpHdr.Length()) > pkt.Data().Size()+header.UDPMinimumSize {
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

			packet := &udpPacket{
				s:        s,
				id:       &id,
				data:     pkt.Data().ExtractVV(),
				nicID:    pkt.NICID,
				netHdr:   pkt.Network(),
				netProto: pkt.NetworkProtocolNumber,
			}

			s.handler.AddPacket(packet)
			return true
		}
		s.SetTransportProtocolHandler(udp.ProtocolNumber, udpHandlePacket)
		return nil
	}
}

type udpPacket struct {
	s        *Stack
	id       *stack.TransportEndpointID
	data     buffer.VectorisedView
	nicID    tcpip.NICID
	netHdr   header.Network
	netProto tcpip.NetworkProtocolNumber
}

func (p *udpPacket) Data() []byte {
	return p.data.ToView()
}

func (p *udpPacket) Drop() {}

func (p *udpPacket) ID() *stack.TransportEndpointID {
	return p.id
}

func (p *udpPacket) LocalAddr() net.Addr {
	return &net.UDPAddr{IP: net.IP(p.id.LocalAddress), Port: int(p.id.LocalPort)}
}

func (p *udpPacket) RemoteAddr() net.Addr {
	return &net.UDPAddr{IP: net.IP(p.id.RemoteAddress), Port: int(p.id.RemotePort)}
}

func (p *udpPacket) WriteBack(b []byte, addr net.Addr) (int, error) {
	v := buffer.View(b)
	if len(v) > header.UDPMaximumPacketSize {
		// Payload can't possibly fit in a packet.
		return 0, fmt.Errorf("%s", &tcpip.ErrMessageTooLong{})
	}

	var (
		localAddress tcpip.Address
		localPort    uint16
	)

	if udpAddr, ok := addr.(*net.UDPAddr); !ok {
		localAddress = p.netHdr.DestinationAddress()
		localPort = p.id.LocalPort
	} else if ipv4 := udpAddr.IP.To4(); ipv4 != nil {
		localAddress = tcpip.Address(ipv4)
		localPort = uint16(udpAddr.Port)
	} else {
		localAddress = tcpip.Address(udpAddr.IP)
		localPort = uint16(udpAddr.Port)
	}

	route, err := p.s.FindRoute(p.nicID, localAddress, p.netHdr.SourceAddress(), p.netProto, false /* multicastLoop */)
	if err != nil {
		return 0, fmt.Errorf("%#v find route: %s", p.id, err)
	}
	defer route.Release()

	data := v.ToVectorisedView()
	if err = sendUDP(route, data, localPort, p.id.RemotePort, udpNoChecksum); err != nil {
		return 0, fmt.Errorf("%v", err)
	}
	return data.Size(), nil
}

// sendUDP sends a UDP segment via the provided network endpoint and under the
// provided identity.
func sendUDP(r *stack.Route, data buffer.VectorisedView, localPort, remotePort uint16, noChecksum bool) tcpip.Error {
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
	if r.RequiresTXTransportChecksum() &&
		(!noChecksum || r.NetProto() == header.IPv6ProtocolNumber) {
		xsum := r.PseudoHeaderChecksum(udp.ProtocolNumber, length)
		for _, v := range data.Views() {
			xsum = header.Checksum(v, xsum)
		}
		udpHdr.SetChecksum(^udpHdr.CalculateChecksum(xsum))
	}

	ttl := r.DefaultTTL()

	if err := r.WritePacket(stack.NetworkHeaderParams{
		Protocol: udp.ProtocolNumber,
		TTL:      ttl,
		TOS:      0, /* default */
	}, pkt); err != nil {
		r.Stats().UDP.PacketSendErrors.Increment()
		return err
	}

	// Track count of packets sent.
	r.Stats().UDP.PacketsSent.Increment()
	return nil
}

// verifyChecksum verifies the checksum unless RX checksum offload is enabled.
// On IPv4, UDP checksum is optional, and a zero value means the transmitter
// omitted the checksum generation (RFC768).
// On IPv6, UDP checksum is not optional (RFC2460 Section 8.1).
func verifyChecksum(hdr header.UDP, pkt *stack.PacketBuffer) bool {
	if !pkt.RXTransportChecksumValidated &&
		(hdr.Checksum() != 0 || pkt.NetworkProtocolNumber == header.IPv6ProtocolNumber) {
		netHdr := pkt.Network()
		xsum := header.PseudoHeaderChecksum(udp.ProtocolNumber, netHdr.DestinationAddress(), netHdr.SourceAddress(), hdr.Length())
		for _, v := range pkt.Data().Views() {
			xsum = header.Checksum(v, xsum)
		}
		return hdr.CalculateChecksum(xsum) == 0xffff
	}
	return true
}
