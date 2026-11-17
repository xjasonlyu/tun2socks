package core_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/core"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func regressionSYN(v6 bool, port uint16, isn uint32) []byte {
	size := header.IPv4MinimumSize
	src, dst := tcpAddr("10.0.0.2"), tcpAddr("192.0.2.10")
	if v6 {
		size = header.IPv6MinimumSize
		src, dst = tcpAddr("fd00::2"), tcpAddr("2001:db8::10")
	}
	b := make([]byte, size+header.TCPMinimumSize)
	if v6 {
		header.IPv6(b).Encode(&header.IPv6Fields{PayloadLength: header.TCPMinimumSize, TransportProtocol: header.TCPProtocolNumber, HopLimit: 64, SrcAddr: src, DstAddr: dst})
	} else {
		ip := header.IPv4(b)
		ip.Encode(&header.IPv4Fields{TotalLength: uint16(len(b)), TTL: 64, Protocol: uint8(header.TCPProtocolNumber), SrcAddr: src, DstAddr: dst})
		ip.SetChecksum(^ip.CalculateChecksum())
	}
	h := header.TCP(b[size:])
	h.Encode(&header.TCPFields{SrcPort: port, DstPort: 443, SeqNum: isn, DataOffset: header.TCPMinimumSize, Flags: header.TCPFlagSyn, WindowSize: 65535})
	h.SetChecksum(^h.CalculateChecksum(header.PseudoHeaderChecksum(header.TCPProtocolNumber, src, dst, header.TCPMinimumSize)))
	return b
}

type stalledResetLink struct {
	*channel.Endpoint
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (l *stalledResetLink) WritePackets(pkts stack.PacketBufferList) (int, tcpip.Error) {
	for _, pkt := range pkts.AsSlice() {
		if pkt.TransportProtocolNumber == header.TCPProtocolNumber {
			h := header.TCP(pkt.TransportHeader().Slice())
			if h.Flags().Contains(header.TCPFlagRst) && h.DestinationPort() == 40001 {
				l.once.Do(func() { close(l.entered) })
				<-l.release
			}
		}
	}
	return l.Endpoint.WritePackets(pkts)
}

func TestDeferredBlockedResetDoesNotBlockOtherTuples(t *testing.T) {
	link := &stalledResetLink{Endpoint: channel.New(64, 1500, ""), entered: make(chan struct{}), release: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	called := make(chan uint16, 10)
	tun := tunnel.New(testProxy{func(_ context.Context, m *M.Metadata) (net.Conn, error) {
		calls.Add(1)
		called <- m.SrcPort
		return nil, errors.New("refused")
	}}, statistic.DefaultManager)
	h := tunnel.NewDeferredTCPHandler(ctx, tun)
	s, err := core.CreateStack(&core.Config{LinkEndpoint: link, TransportHandler: h, DeferredTCP: &core.DeferredTCPConfig{Context: ctx}})
	if err != nil {
		t.Fatal(err)
	}
	var releaseOnce sync.Once
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(link.release) })
		cancel()
		h.Close()
		tun.Close()
		s.Close()
		s.Wait()
		link.Close()
	})
	injectPolicyPacket(link.Endpoint, regressionSYN(false, 40001, 100), false)
	select {
	case <-link.entered:
	case <-time.After(time.Second):
		t.Fatal("RST never reached output")
	}
	<-called
	finished := make(chan struct{})
	go func() {
		// Same tuple must remain reserved even though gVisor already removed
		// its in-flight record before attempting the blocked RST write.
		injectPolicyPacket(link.Endpoint, regressionSYN(false, 40001, 100), false)
		injectPolicyPacket(link.Endpoint, regressionSYN(false, 40001, 200), false)
		injectPolicyPacket(link.Endpoint, regressionSYN(false, 40002, 300), false)
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("RST holds global SYN admission lock")
	}
	select {
	case port := <-called:
		if port != 40002 {
			t.Fatalf("readmitted pending tuple: %d", port)
		}
	case <-time.After(time.Second):
		t.Fatal("other tuple not admitted")
	}
	if calls.Load() != 2 {
		t.Fatalf("duplicate dial: %d", calls.Load())
	}
	releaseOnce.Do(func() { close(link.release) })
}

type countedConnector struct {
	*tunnel.DeferredTCPHandler
	calls atomic.Int32
}

func (c *countedConnector) DialTCP(ctx context.Context, id stack.TransportEndpointID) (net.Conn, error) {
	c.calls.Add(1)
	return c.DeferredTCPHandler.DialTCP(ctx, id)
}

func TestDeferredCancellationNotCachedButTimeoutIs(t *testing.T) {
	for _, mode := range []string{"handler-close", "wrapped-cancel", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			link := channel.New(64, 1500, "")
			entered := make(chan struct{}, 1)
			tun := tunnel.New(testProxy{func(ctx context.Context, _ *M.Metadata) (net.Conn, error) {
				if mode == "handler-close" {
					select {
					case entered <- struct{}{}:
					default:
					}
					<-ctx.Done()
					return nil, ctx.Err()
				}
				if mode == "wrapped-cancel" {
					return nil, fmt.Errorf("shutdown: %w", context.Canceled)
				}
				return nil, context.DeadlineExceeded
			}}, statistic.DefaultManager)
			// Intentionally NOT the stack context: Close cancels the handler only.
			h := &countedConnector{DeferredTCPHandler: tunnel.NewDeferredTCPHandler(context.Background(), tun)}
			s, err := core.CreateStack(&core.Config{LinkEndpoint: link, TransportHandler: h, DeferredTCP: &core.DeferredTCPConfig{Context: ctx}})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				h.Close()
				cancel()
				tun.Close()
				s.Close()
				s.Wait()
				link.Close()
			})
			packet := regressionSYN(false, 40003, 123)
			injectPolicyPacket(link, packet, false)
			if mode == "handler-close" {
				select {
				case <-entered:
				case <-time.After(time.Second):
					t.Fatal("not dialing")
				}
				h.Close()
			}
			readPolicyReply(t, link)
			// Replays cover the tiny interval before Complete releases the tuple.
			for i := 0; i < 20; i++ {
				time.Sleep(5 * time.Millisecond)
				injectPolicyPacket(link, packet, false)
			}
			if ctx.Err() != nil {
				t.Fatal("stack context unexpectedly canceled")
			}
			if mode == "timeout" {
				if h.calls.Load() != 1 {
					t.Fatal("real deadline failure not cached")
				}
			} else if h.calls.Load() < 2 {
				t.Fatal("lifecycle cancellation cached as remote failure")
			}
		})
	}
}

func TestRejectIPv6ChecksumOffloadAndMalformedPackets(t *testing.T) {
	for _, tcpPacket := range []bool{false, true} {
		for _, validated := range []bool{false, true} {
			for _, malformed := range []bool{false, true} {
				t.Run(fmt.Sprintf("tcp=%v/validated=%v/malformed=%v", tcpPacket, validated, malformed), func(t *testing.T) {
					link, _, _ := newRejectPacketStack(t, true)
					if validated {
						// NIC input derives RXChecksumValidated from the device capability.
						link.LinkEPCapabilities |= stack.CapabilityRXChecksumOffload
					}
					b := udpPolicyPacket(true)
					if tcpPacket {
						b = regressionSYN(true, 40004, 321)
						h := header.TCP(b[header.IPv6MinimumSize:])
						h.SetChecksum(h.Checksum() ^ 1)
						if malformed {
							b[header.IPv6MinimumSize+12] = 0x10
						}
					} else {
						h := header.UDP(b[header.IPv6MinimumSize:])
						h.SetChecksum(h.Checksum() ^ 1)
						if malformed {
							b[header.IPv6MinimumSize+4], b[header.IPv6MinimumSize+5] = 0, 1
						}
					}
					pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{Payload: buffer.MakeWithData(b)})
					pkt.RXChecksumValidated = validated
					link.InjectInbound(header.IPv6ProtocolNumber, pkt)
					pkt.DecRef()
					if validated && !malformed {
						reply := readPolicyReply(t, link)
						if tcpPacket {
							if !header.TCP(reply[header.IPv6MinimumSize:]).Flags().Contains(header.TCPFlagRst) {
								t.Fatal("missing RST")
							}
						} else if header.ICMPv6(reply[header.IPv6MinimumSize:]).Code() != header.ICMPv6Prohibited {
							t.Fatal("missing ICMP rejection")
						}
					} else if p := link.Read(); p != nil {
						p.DecRef()
						t.Fatal("replied to malformed or unchecked corrupt packet")
					}
				})
			}
		}
	}
}

func TestRejectIPv6AtomicFragmentStillValidatesTCP(t *testing.T) {
	for _, corrupt := range []bool{false, true} {
		t.Run(fmt.Sprint(corrupt), func(t *testing.T) {
			link, _, _ := newRejectPacketStack(t, true)
			b := regressionSYN(true, 40005, 456)
			if corrupt {
				b[len(b)-1] ^= 1
			}
			b[6] = 44 // An atomic IPv6 fragment (offset=0, M=0).
			b[5] += 8
			fragment := []byte{byte(header.TCPProtocolNumber), 0, 0, 0, 0, 0, 0, 1}
			packet := append(append(append([]byte(nil), b[:header.IPv6MinimumSize]...), fragment...), b[header.IPv6MinimumSize:]...)
			injectPolicyPacket(link, packet, true)
			if corrupt {
				if p := link.Read(); p != nil {
					p.DecRef()
					t.Fatal("replied to corrupt fragmented TCP")
				}
			} else {
				readPolicyReply(t, link)
			}
		})
	}
}
