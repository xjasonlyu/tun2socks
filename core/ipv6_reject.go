package core

import (
	"errors"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// Register during stack initialization, before the NIC can deliver packets.
// The transport hook runs AFTER IPv6 extension parsing and fragment reassembly.
// Rejected flows never create endpoints, so subsequent packets hit this same
// hook. The policy is immutable for this stack's lifetime.
func setTransportHandler(s *stack.Stack, protocol tcpip.TransportProtocolNumber, next func(stack.TransportEndpointID, *stack.PacketBuffer) bool, rejectIPv6 bool) error {
	if !rejectIPv6 {
		// Default mode installs exactly the original handler, with no wrapper.
		s.SetTransportProtocolHandler(protocol, next)
		return nil
	}
	h, ok := s.NetworkProtocolInstance(ipv6.ProtocolNumber).(stack.RejectIPv6WithHandler)
	if !ok {
		return errors.New("IPv6 protocol does not support local rejection")
	}
	reason := stack.RejectIPv6WithICMPAdminProhibited
	if protocol == header.TCPProtocolNumber {
		reason = stack.RejectIPv6WithTCPReset
	}
	s.SetTransportProtocolHandler(protocol, func(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
		if pkt.NetworkProtocolNumber != ipv6.ProtocolNumber {
			return next(id, pkt)
		}
		// The transport dispatcher parsed headers but has not yet validated
		// their checksums. Do not respond to malformed TCP/UDP datagrams.
		if !validRejectedPacket(pkt, protocol) {
			return true
		}
		// gVisor constructs checksummed replies and applies ICMP suppression
		// and rate limiting. Failure to reply must never permit forwarding.
		_ = h.SendRejectionError(pkt, reason, stack.Input)
		return true
	})
	return nil
}

func validRejectedPacket(pkt *stack.PacketBuffer, protocol tcpip.TransportProtocolNumber) bool {
	ip := header.IPv6(pkt.NetworkHeader().Slice())
	dataSize := uint16(pkt.Data().Size())
	switch protocol {
	case header.TCPProtocolNumber:
		h := header.TCP(pkt.TransportHeader().Slice())
		if len(h) < header.TCPMinimumSize || int(h.DataOffset()) < header.TCPMinimumSize || int(h.DataOffset()) > len(h) {
			return false
		}
		if !pkt.RXChecksumValidated {
			// Validate here even for reassembled/atomic fragments: gVisor's
			// reset helper skips checksum checks when a fragment header remains.
			if !h.IsChecksumValid(ip.SourceAddress(), ip.DestinationAddress(), pkt.Data().Checksum(), dataSize) {
				return false
			}
			// RejectWithTCPReset honors this flag; do not checksum twice.
			pkt.RXChecksumValidated = true
		}
		return true
	case header.UDPProtocolNumber:
		h := header.UDP(pkt.TransportHeader().Slice())
		if len(h) < header.UDPMinimumSize || h.Length() != uint16(len(h))+dataSize {
			return false
		}
		return pkt.RXChecksumValidated || (h.Checksum() != 0 && h.IsChecksumValid(ip.SourceAddress(), ip.DestinationAddress(), pkt.Data().Checksum()))
	default:
		return false
	}
}
