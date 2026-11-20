package core_test

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/xjasonlyu/tun2socks/v2/core"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
)

// Replay the actual reported pattern: same tuple and raw ISN after each reset.
// Raw packets are necessary: gonet correctly aborts its dial on the first RST,
// whereas the Windows client retries the SYN after receiving it.
func TestFailedSYNReplayDoesNotRedial(t *testing.T) {
	for _, v6 := range []bool{false, true} {
		t.Run(map[bool]string{false: "IPv4", true: "IPv6"}[v6], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			link := channel.New(64, 1500, "")
			var calls atomic.Int32
			started, release := make(chan struct{}), make(chan struct{})
			proxy := testProxy{func(ctx context.Context, _ *M.Metadata) (net.Conn, error) {
				if calls.Add(1) == 1 {
					close(started)
					select {
					case <-release:
					case <-ctx.Done():
					}
				}
				return nil, errors.New("simulated remote connect failure")
			}}
			tun := tunnel.New(proxy, statistic.DefaultManager)
			tun.ProcessAsync()
			handler := tunnel.NewDeferredTCPHandler(ctx, tun)
			s, err := core.CreateStack(&core.Config{
				LinkEndpoint: link, TransportHandler: handler,
				DeferredTCP: &core.DeferredTCPConfig{Context: ctx},
			})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { cancel(); tun.Close(); s.Close(); s.Wait(); link.Close() })
			proto, ipSize := header.IPv4ProtocolNumber, header.IPv4MinimumSize
			src, dst := tcpAddr("10.60.173.33"), tcpAddr("1.2.3.4")
			if v6 {
				proto, ipSize = header.IPv6ProtocolNumber, header.IPv6MinimumSize
				src, dst = tcpAddr("fd00::2"), tcpAddr("2001:db8::10")
			}
			inject := func(isn uint32, corrupt bool) {
				bytes := make([]byte, ipSize+header.TCPMinimumSize)
				if v6 {
					header.IPv6(bytes).Encode(&header.IPv6Fields{PayloadLength: header.TCPMinimumSize, TransportProtocol: header.TCPProtocolNumber, HopLimit: 64, SrcAddr: src, DstAddr: dst})
				} else {
					ip := header.IPv4(bytes)
					ip.Encode(&header.IPv4Fields{TotalLength: uint16(len(bytes)), TTL: 64, Protocol: uint8(header.TCPProtocolNumber), SrcAddr: src, DstAddr: dst})
					ip.SetChecksum(^ip.CalculateChecksum())
				}
				tcp := header.TCP(bytes[ipSize:])
				tcp.Encode(&header.TCPFields{SrcPort: 63293, DstPort: 5, SeqNum: isn, DataOffset: header.TCPMinimumSize, Flags: header.TCPFlagSyn, WindowSize: 65535})
				tcp.SetChecksum(^tcp.CalculateChecksum(header.PseudoHeaderChecksum(header.TCPProtocolNumber, src, dst, header.TCPMinimumSize)))
				if corrupt {
					tcp.SetChecksum(tcp.Checksum() ^ 1)
				}
				pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{Payload: buffer.MakeWithData(bytes)})
				link.InjectInbound(proto, pkt)
				pkt.DecRef()
			}
			readReset := func(isn uint32) {
				waitCtx, stop := context.WithTimeout(ctx, time.Second)
				defer stop()
				for {
					pkt := link.ReadContext(waitCtx)
					if pkt == nil {
						t.Fatal("missing RST")
					}
					if pkt.TransportProtocolNumber != header.TCPProtocolNumber {
						pkt.DecRef()
						continue
					}
					tcp := header.TCP(pkt.TransportHeader().Slice())
					flags, seq, ack := tcp.Flags(), tcp.SequenceNumber(), tcp.AckNumber()
					valid := tcp.IsChecksumValid(dst, src, 0, 0)
					pkt.DecRef()
					if flags != header.TCPFlagRst|header.TCPFlagAck || seq != 0 || ack != isn+1 || !valid {
						t.Fatalf("invalid reset flags=%v seq=%d ack=%d checksum=%v", flags, seq, ack, valid)
					}
					return
				}
			}
			const isn uint32 = 3361995830
			inject(isn, true) // Invalid checksum must not reserve/poison a key.
			inject(isn, false)
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("dial not started")
			}
			for i := 0; i < 8; i++ {
				inject(isn, false)
			}
			close(release)
			readReset(isn)
			for i := 0; i < 3; i++ {
				inject(isn, false)
				readReset(isn)
			}
			if calls.Load() != 1 {
				t.Fatalf("same SYN caused %d proxy dials", calls.Load())
			}
			inject(isn+100, false)
			readReset(isn + 100)
			if calls.Load() != 2 {
				t.Fatal("new ISN was incorrectly cached")
			}
			inject(isn, false)
			readReset(isn)
			if calls.Load() != 2 {
				t.Fatal("new ISN overwrote old failure record")
			}
		})
	}
}
