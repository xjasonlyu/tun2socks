package core

import (
	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/checksum"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"

	"github.com/xjasonlyu/tun2socks/v2/core/option"
)

func withICMPHandler(factory TransportProtocolHandlerFactory) option.Option {
	return func(s *stack.Stack) error {
		if factory != nil {
			handler := factory(s)
			if handler != nil {
				s.SetTransportProtocolHandler(icmp.ProtocolNumber4, handler.HandlePacket)
				return nil
			}
		}
		f := newICMPForwarder(s)
		s.SetTransportProtocolHandler(icmp.ProtocolNumber4, f.HandlePacket)
		return nil
	}
}

type icmpForwarder struct {
	s *stack.Stack
}

func newICMPForwarder(s *stack.Stack) *icmpForwarder {
	return &icmpForwarder{s: s}
}

func (f *icmpForwarder) HandlePacket(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
	switch pkt.NetworkProtocolNumber {
	case ipv4.ProtocolNumber:
		return f.handlePacket4(id, pkt)
	case ipv6.ProtocolNumber:
		return f.handlePacket6(id, pkt)
	default:
		return false
	}
}

// Ref: https://github.com/google/gvisor/blob/c58cb637/pkg/tcpip/network/ipv4/icmp.go#L345-L461
func (f *icmpForwarder) handlePacket4(_ stack.TransportEndpointID, pkt *stack.PacketBuffer) (handled bool) {
	if h := header.ICMPv4(pkt.TransportHeader().Slice()); h.Type() != header.ICMPv4Echo {
		return false
	}

	ipHdr := header.IPv4(pkt.NetworkHeader().Slice())
	localAddressBroadcast := pkt.NetworkPacketInfo.LocalAddressBroadcast

	// As per RFC 1122 section 3.2.1.3, when a host sends any datagram, the IP
	// source address MUST be one of its own IP addresses (but not a broadcast
	// or multicast address).
	localAddr := ipHdr.DestinationAddress()
	if localAddressBroadcast || header.IsV4MulticastAddress(localAddr) {
		localAddr = tcpip.Address{}
	}

	r, err := f.s.FindRoute(pkt.NICID, localAddr, ipHdr.SourceAddress(), ipv4.ProtocolNumber, false /* multicastLoop */)
	if err != nil {
		// If we cannot find a route to the destination, silently drop the packet.
		return false
	}
	defer r.Release()

	replyData := stack.PayloadSince(pkt.TransportHeader())
	defer replyData.Release()

	replyICMPHdr := header.ICMPv4(replyData.AsSlice())
	replyICMPHdr.SetType(header.ICMPv4EchoReply)
	replyICMPHdr.SetCode(0) // RFC 792: EchoReply must have Code=0.
	replyICMPHdr.SetChecksum(0)
	replyICMPHdr.SetChecksum(^checksum.Checksum(replyData.AsSlice(), 0))

	replyBuf := buffer.MakeWithView(replyData.Clone())
	replyPkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
		ReserveHeaderBytes: int(r.MaxHeaderLength()),
		Payload:            replyBuf,
	})
	defer replyPkt.DecRef()

	sent := f.s.Stats().ICMP.V4.PacketsSent
	if err := r.WritePacket(stack.NetworkHeaderParams{
		Protocol: header.ICMPv4ProtocolNumber,
		TTL:      r.DefaultTTL(),
	}, replyPkt); err != nil {
		sent.Dropped.Increment()
		return false
	}
	sent.EchoReply.Increment()
	return true
}

func (f *icmpForwarder) handlePacket6(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
	return false // not implemented
}
