//go:build unix

package iobased

import (
	"bytes"
	"os"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type readTestDispatcher struct {
	proto tcpip.NetworkProtocolNumber
	data  []byte
	got   chan struct{}
}

func (d *readTestDispatcher) DeliverNetworkPacket(protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	d.proto = protocol
	d.data = pkt.ToView().AsSlice()
	select {
	case d.got <- struct{}{}:
	default:
	}
}

func (d *readTestDispatcher) DeliverLinkPacket(tcpip.NetworkProtocolNumber, *stack.PacketBuffer) {}

func newSocketpairEndpoint(t *testing.T, offset int) (*Endpoint, int) {
	t.Helper()
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_DGRAM, 0)
	if err != nil {
		t.Fatalf("Socketpair() error = %v", err)
	}
	t.Cleanup(func() { _ = unix.Close(fds[1]) })

	ep, err := New(os.NewFile(uintptr(fds[0]), "test"), 1500, offset)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return ep, fds[1]
}

func TestReadStripsOffset(t *testing.T) {
	ep, writer := newSocketpairEndpoint(t, 4)
	dispatcher := &readTestDispatcher{got: make(chan struct{}, 1)}
	ep.Attach(dispatcher)

	// A darwin utun datagram is a 4-byte header (byte 3 = address family)
	// followed by the IP packet. With the offset applied, the injected
	// packet must be the IP packet itself with the header stripped.
	ipv4 := []byte{0x45, 0x00, 0x00, 0x18, 0x00, 0x00, 0x00, 0x00, 0x40, 0x11, 0x00, 0x00, 0x7f, 0x00, 0x00, 0x01, 0x7f, 0x00, 0x00, 0x01, 0x12, 0x34, 0x56, 0x78}
	datagram := append([]byte{0x00, 0x00, 0x00, 0x02}, ipv4...)

	if _, err := unix.Write(writer, datagram); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	select {
	case <-dispatcher.got:
	case <-time.After(5 * time.Second):
		t.Fatal("packet not dispatched; 4-byte TUN header not stripped")
	}

	if dispatcher.proto != header.IPv4ProtocolNumber {
		t.Errorf("protocol = %d, want IPv4", dispatcher.proto)
	}
	// The 4-byte utun header must be stripped: the injected packet has to
	// start with the IP packet itself. (The read path may append up to
	// `offset` stale buffer bytes after the packet.)
	if !bytes.HasPrefix(dispatcher.data, ipv4) {
		t.Errorf("injected packet = %x, want prefix %x (4-byte TUN header must be stripped)", dispatcher.data, ipv4)
	}
}

func TestReadNoOffset(t *testing.T) {
	ep, writer := newSocketpairEndpoint(t, 0)
	dispatcher := &readTestDispatcher{got: make(chan struct{}, 1)}
	ep.Attach(dispatcher)

	ipv4 := []byte{0x45, 0x00, 0x00, 0x18, 0x00, 0x00, 0x00, 0x00, 0x40, 0x11, 0x00, 0x00, 0x7f, 0x00, 0x00, 0x01, 0x7f, 0x00, 0x00, 0x01, 0x12, 0x34, 0x56, 0x78}
	if _, err := unix.Write(writer, ipv4); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	select {
	case <-dispatcher.got:
	case <-time.After(5 * time.Second):
		t.Fatal("raw IPv4 packet not dispatched")
	}

	if dispatcher.proto != header.IPv4ProtocolNumber {
		t.Errorf("protocol = %d, want IPv4", dispatcher.proto)
	}
	if !bytes.Equal(dispatcher.data, ipv4) {
		t.Errorf("injected packet = %x, want %x", dispatcher.data, ipv4)
	}
}
