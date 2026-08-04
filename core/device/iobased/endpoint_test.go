package iobased

import (
	"bytes"
	"syscall"
	"testing"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func TestWritePacketDarwinAFHeader(t *testing.T) {
	tests := []struct {
		name   string
		proto  tcpip.NetworkProtocolNumber
		packet []byte
		wantAF byte
	}{
		{
			name:   "ipv4",
			proto:  header.IPv4ProtocolNumber,
			packet: []byte{0x45, 0x00, 0x00, 0x14, 0x00, 0x00, 0x00, 0x00, 0x40, 0x11, 0x00, 0x00, 0x7f, 0x00, 0x00, 0x01, 0x7f, 0x00, 0x00, 0x01},
			wantAF: byte(syscall.AF_INET),
		},
		{
			name:   "ipv6",
			proto:  header.IPv6ProtocolNumber,
			packet: []byte{0x60, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantAF: byte(syscall.AF_INET6),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w bytes.Buffer
			ep, err := New(&w, 1500, 4)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
				Payload: buffer.MakeWithData(tt.packet),
			})
			pkt.NetworkProtocolNumber = tt.proto
			if err := ep.writePacket(pkt); err != nil {
				t.Fatalf("writePacket() error = %v", err)
			}

			got := w.Bytes()
			if len(got) != 4+len(tt.packet) {
				t.Fatalf("written length = %d, want %d", len(got), 4+len(tt.packet))
			}
			if got[3] != tt.wantAF {
				t.Errorf("header byte 3 = %d, want address family %d", got[3], tt.wantAF)
			}
			if !bytes.Equal(got[4:], tt.packet) {
				t.Errorf("packet after header mismatch:\n got %x\nwant %x", got[4:], tt.packet)
			}
		})
	}
}

func TestWritePacketNoHeaderWithoutOffset(t *testing.T) {
	packet := []byte{0x45, 0x00, 0x00, 0x14, 0x00, 0x00, 0x00, 0x00, 0x40, 0x11, 0x00, 0x00, 0x7f, 0x00, 0x00, 0x01, 0x7f, 0x00, 0x00, 0x01}

	var w bytes.Buffer
	ep, err := New(&w, 1500, 0)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
		Payload: buffer.MakeWithData(packet),
	})
	pkt.NetworkProtocolNumber = header.IPv4ProtocolNumber
	if err := ep.writePacket(pkt); err != nil {
		t.Fatalf("writePacket() error = %v", err)
	}

	if got := w.Bytes(); !bytes.Equal(got, packet) {
		t.Errorf("written bytes = %x, want %x", got, packet)
	}
}
