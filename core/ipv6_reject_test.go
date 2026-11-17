package core_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/checksum"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/adapter"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
)

func TestRejectIPv6TCPIndependentOfDeferredHandshake(t *testing.T) {
	for _, deferred := range []bool{false, true} {
		for _, reject := range []bool{false, true} {
			t.Run(fmt.Sprintf("deferred=%v/reject=%v", deferred, reject), func(t *testing.T) {
				var dials atomic.Int32
				h := newTCPPolicyHarness(t, func(context.Context, *M.Metadata) (net.Conn, error) {
					dials.Add(1)
					return echoRemote(), nil
				}, deferred, reject)
				for i := 0; i < 3; i++ {
					c, err := h.dial("2001:db8::10")
					if reject {
						if c != nil {
							c.Close()
						}
						if err == nil || h.resets.Load() != int32(i+1) || h.synACK.Load() != 0 || dials.Load() != 0 {
							t.Fatalf("not locally rejected: err=%v resets=%d synACK=%d dials=%d", err, h.resets.Load(), h.synACK.Load(), dials.Load())
						}
					} else {
						if err != nil {
							t.Fatal(err)
						}
						assertEcho(t, c)
					}
				}
				before := dials.Load()
				c, err := h.dial("192.0.2.10")
				if err != nil {
					t.Fatal("IPv4 affected:", err)
				}
				assertEcho(t, c)
				if dials.Load() != before+1 {
					t.Fatal("IPv4 did not use proxy")
				}
			})
		}
	}
}

func TestRejectIPv6AlonePreservesIPv4EarlyHandshake(t *testing.T) {
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	h := newTCPPolicyHarness(t, func(ctx context.Context, _ *M.Metadata) (net.Conn, error) {
		select {
		case <-release:
			return echoRemote(), nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}, false, true)
	c, err := h.dial("192.0.2.10")
	if err != nil {
		t.Fatal("RejectIPv6 enabled deferred behavior for IPv4:", err)
	}
	defer c.Close()
	if h.synACK.Load() != 1 {
		t.Fatal("IPv4 early SYN-ACK missing")
	}
	once.Do(func() { close(release) })
	assertEcho(t, c)
}

type rejectPacketHandler struct{ udp atomic.Int32 }

func (*rejectPacketHandler) HandleTCP(c adapter.TCPConn)   { c.Close() }
func (h *rejectPacketHandler) HandleUDP(c adapter.UDPConn) { h.udp.Add(1); c.Close() }

func newRejectPacketStack(t *testing.T, reject bool) (*channel.Endpoint, *rejectPacketHandler, *stack.Stack) {
	t.Helper()
	link := channel.New(128, 1500, "")
	h := &rejectPacketHandler{}
	s, err := core.CreateStack(&core.Config{LinkEndpoint: link, TransportHandler: h, RejectIPv6: reject})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close(); s.Wait(); link.Close() })
	if s.IPTables().Modified() {
		t.Fatal("transport policy must not enable unrelated iptables/conntrack machinery")
	}
	return link, h, s
}

func udpPolicyPacket(v6 bool) []byte {
	ipSize := header.IPv4MinimumSize
	src, dst := tcpAddr("10.0.0.2"), tcpAddr("192.0.2.10")
	if v6 {
		ipSize = header.IPv6MinimumSize
		src, dst = tcpAddr("fd00::2"), tcpAddr("2001:db8::10")
	}
	payload := []byte("UDP including QUIC is opaque")
	b := make([]byte, ipSize+header.UDPMinimumSize+len(payload))
	if v6 {
		header.IPv6(b).Encode(&header.IPv6Fields{PayloadLength: uint16(len(b) - ipSize), TransportProtocol: header.UDPProtocolNumber, HopLimit: 64, SrcAddr: src, DstAddr: dst})
	} else {
		ip := header.IPv4(b)
		ip.Encode(&header.IPv4Fields{TotalLength: uint16(len(b)), TTL: 64, Protocol: uint8(header.UDPProtocolNumber), SrcAddr: src, DstAddr: dst})
		ip.SetChecksum(^ip.CalculateChecksum())
	}
	u := header.UDP(b[ipSize:])
	u.Encode(&header.UDPFields{SrcPort: 49152, DstPort: 443, Length: uint16(len(b) - ipSize)})
	copy(b[ipSize+header.UDPMinimumSize:], payload)
	u.SetChecksum(^checksum.Checksum(b[ipSize:], header.PseudoHeaderChecksum(header.UDPProtocolNumber, src, dst, uint16(len(b)-ipSize))))
	return b
}

func injectPolicyPacket(link *channel.Endpoint, b []byte, v6 bool) {
	proto := header.IPv4ProtocolNumber
	if v6 {
		proto = header.IPv6ProtocolNumber
	}
	p := stack.NewPacketBuffer(stack.PacketBufferOptions{Payload: buffer.MakeWithData(b)})
	link.InjectInbound(proto, p)
	p.DecRef()
}

func readPolicyReply(t *testing.T, link *channel.Endpoint) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	p := link.ReadContext(ctx)
	if p == nil {
		t.Fatal("missing local rejection reply")
	}
	defer p.DecRef()
	b := p.ToBuffer()
	defer b.Release()
	return append([]byte(nil), b.Flatten()...)
}

func TestRejectIPv6UDPAndRecovery(t *testing.T) {
	// Recreating the stack is the caller's reconnect policy update. Verify both
	// transitions and no global/sticky rejection state across stack instances.
	for _, reject := range []bool{true, false, true} {
		for _, v6 := range []bool{true, false} {
			t.Run(fmt.Sprintf("reject=%v/ipv6=%v", reject, v6), func(t *testing.T) {
				link, h, _ := newRejectPacketStack(t, reject)
				original := udpPolicyPacket(v6)
				injectPolicyPacket(link, original, v6)
				if !reject || !v6 {
					if h.udp.Load() != 1 {
						t.Fatal("normal UDP forwarding affected")
					}
					return
				}
				if h.udp.Load() != 0 {
					t.Fatal("rejected UDP reached transport handler")
				}
				reply := readPolicyReply(t, link)
				ip := header.IPv6(reply)
				proto, ok := ip.TryParseTransportProtocol()
				icmp := header.ICMPv6(reply[header.IPv6MinimumSize:])
				if !ok || proto != header.ICMPv6ProtocolNumber || icmp.Type() != header.ICMPv6DstUnreachable || icmp.Code() != header.ICMPv6Prohibited {
					t.Fatalf("wrong ICMPv6 reply: %x", reply)
				}
				if ip.SourceAddress() != tcpAddr("2001:db8::10") || ip.DestinationAddress() != tcpAddr("fd00::2") {
					t.Fatal("wrong reply addresses")
				}
				if checksum.Checksum(icmp, header.PseudoHeaderChecksum(header.ICMPv6ProtocolNumber, ip.SourceAddress(), ip.DestinationAddress(), uint16(len(icmp)))) != 0xffff {
					t.Fatal("invalid ICMPv6 checksum")
				}
				if !bytes.Equal(icmp[header.ICMPv6MinimumSize:], original) {
					t.Fatal("ICMPv6 must quote original IP and UDP datagram")
				}
			})
		}
	}
}

func TestRejectIPv6RawSYNReplay(t *testing.T) {
	link, _, _ := newRejectPacketStack(t, true)
	src, dst := tcpAddr("fd00::2"), tcpAddr("2001:db8::10")
	const isn = 3361995830
	b := make([]byte, header.IPv6MinimumSize+header.TCPMinimumSize)
	header.IPv6(b).Encode(&header.IPv6Fields{PayloadLength: header.TCPMinimumSize, TransportProtocol: header.TCPProtocolNumber, HopLimit: 64, SrcAddr: src, DstAddr: dst})
	tcp := header.TCP(b[header.IPv6MinimumSize:])
	tcp.Encode(&header.TCPFields{SrcPort: 63293, DstPort: 5, SeqNum: isn, DataOffset: header.TCPMinimumSize, Flags: header.TCPFlagSyn, WindowSize: 65535})
	tcp.SetChecksum(^tcp.CalculateChecksum(header.PseudoHeaderChecksum(header.TCPProtocolNumber, src, dst, header.TCPMinimumSize)))
	for i := 0; i < 3; i++ {
		injectPolicyPacket(link, b, true)
		reply := readPolicyReply(t, link)
		ip := header.IPv6(reply)
		r := header.TCP(reply[header.IPv6MinimumSize:])
		if ip.SourceAddress() != dst || ip.DestinationAddress() != src || r.Flags() != header.TCPFlagRst|header.TCPFlagAck || r.SequenceNumber() != 0 || r.AckNumber() != isn+1 || !r.IsChecksumValid(dst, src, 0, 0) {
			t.Fatal("invalid local RST")
		}
	}
}

func TestRejectIPv6UDPWithExtensionAndFragments(t *testing.T) {
	for _, fragmented := range []bool{false, true} {
		t.Run(fmt.Sprint(fragmented), func(t *testing.T) {
			link, h, _ := newRejectPacketStack(t, true)
			original := udpPolicyPacket(true)
			payload := original[header.IPv6MinimumSize:]
			if !fragmented {
				b := append([]byte(nil), original[:header.IPv6MinimumSize]...)
				b[6] = 0 // Hop-by-hop options.
				b = append(b, byte(header.UDPProtocolNumber), 0, 0, 0, 0, 0, 0, 0)
				b = append(b, payload...)
				binary.BigEndian.PutUint16(b[4:6], uint16(len(b)-header.IPv6MinimumSize))
				injectPolicyPacket(link, b, true)
			} else {
				for _, offset := range []int{0, 16} {
					end := len(payload)
					flags := uint16(offset)
					if offset == 0 {
						end = 16
						flags |= 1
					}
					b := append([]byte(nil), original[:header.IPv6MinimumSize]...)
					b[6] = 44 // Fragment header.
					frag := make([]byte, 8)
					frag[0] = byte(header.UDPProtocolNumber)
					binary.BigEndian.PutUint16(frag[2:4], flags)
					binary.BigEndian.PutUint32(frag[4:8], 12345)
					b = append(b, frag...)
					b = append(b, payload[offset:end]...)
					binary.BigEndian.PutUint16(b[4:6], uint16(len(b)-header.IPv6MinimumSize))
					injectPolicyPacket(link, b, true)
				}
			}
			reply := readPolicyReply(t, link)
			icmp := header.ICMPv6(reply[header.IPv6MinimumSize:])
			if h.udp.Load() != 0 || icmp.Type() != header.ICMPv6DstUnreachable || icmp.Code() != header.ICMPv6Prohibited {
				t.Fatal("UDP bypassed rejection")
			}
		})
	}
}

func TestRejectIPv6LeavesICMPv6Alone(t *testing.T) {
	link, _, s := newRejectPacketStack(t, true)
	b := make([]byte, header.IPv6MinimumSize+header.ICMPv6MinimumSize)
	src, dst := tcpAddr("fd00::2"), tcpAddr("2001:db8::10")
	header.IPv6(b).Encode(&header.IPv6Fields{PayloadLength: header.ICMPv6MinimumSize, TransportProtocol: header.ICMPv6ProtocolNumber, HopLimit: 64, SrcAddr: src, DstAddr: dst})
	i := header.ICMPv6(b[header.IPv6MinimumSize:])
	i.SetType(header.ICMPv6EchoRequest)
	i.SetChecksum(^checksum.Checksum(i, header.PseudoHeaderChecksum(header.ICMPv6ProtocolNumber, src, dst, uint16(len(i)))))
	before := s.Stats().IP.IPTablesInputDropped.Value()
	injectPolicyPacket(link, b, true)
	if s.Stats().IP.IPTablesInputDropped.Value() != before {
		t.Fatal("ICMPv6 blocked by TCP/UDP policy")
	}
}

func TestRejectIPv6UDPRateLimitStillDropsPackets(t *testing.T) {
	link, h, s := newRejectPacketStack(t, true)
	s.SetICMPLimit(0)
	s.SetICMPBurst(1)
	for i := 0; i < 20; i++ {
		injectPolicyPacket(link, udpPolicyPacket(true), true)
	}
	count := 0
	for {
		p := link.Read()
		if p == nil {
			break
		}
		count++
		p.DecRef()
	}
	if count != 1 || h.udp.Load() != 0 {
		t.Fatalf("rate limit responses=%d forwarded=%d", count, h.udp.Load())
	}
}

func TestRejectIPv6DoesNotReplyToCorruptUDP(t *testing.T) {
	link, h, _ := newRejectPacketStack(t, true)
	b := udpPolicyPacket(true)
	b[len(b)-1] ^= 1
	injectPolicyPacket(link, b, true)
	if p := link.Read(); p != nil {
		p.DecRef()
		t.Fatal("replied to invalid UDP checksum")
	}
	if h.udp.Load() != 0 {
		t.Fatal("invalid UDP forwarded")
	}
}
