//go:build unix

package engine

import (
	"bytes"
	"net/url"
	"runtime"
	"strconv"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/fdbased"
)

type fdDispatcher struct {
	proto tcpip.NetworkProtocolNumber
	data  []byte
	got   chan struct{}
}

func (d *fdDispatcher) DeliverNetworkPacket(protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	d.proto = protocol
	d.data = pkt.ToView().AsSlice()
	select {
	case d.got <- struct{}{}:
	default:
	}
}

func (d *fdDispatcher) DeliverLinkPacket(tcpip.NetworkProtocolNumber, *stack.PacketBuffer) {}

func TestParseFDDarwinTUNOffset(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin utun 4-byte TUN header offset is darwin-only")
	}

	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_DGRAM, 0)
	if err != nil {
		t.Fatalf("Socketpair() error = %v", err)
	}
	defer unix.Close(fds[1])

	// Match the fix's premise: the fd passed to fd:// is non-blocking so
	// os.File uses the runtime netpoller (as with real utun fds).
	if err := unix.SetNonblock(fds[0], true); err != nil {
		t.Fatalf("SetNonblock() error = %v", err)
	}

	dev, err := parseFDURL(fds[0])
	if err != nil {
		t.Fatalf("parseFD() error = %v", err)
	}
	defer dev.Close()

	dispatcher := &fdDispatcher{got: make(chan struct{}, 1)}
	dev.Attach(dispatcher)

	// 20-byte IPv4 header + 4-byte transport header (src/dst ports).
	// The linux fdbased path drops packets whose payload is shorter than
	// the transport ports (tcpipConnectionID in gvisor's processors.go),
	// so the test packet must be at least 24 bytes.
	ipv4 := []byte{0x45, 0x00, 0x00, 0x18, 0x00, 0x00, 0x00, 0x00, 0x40, 0x11, 0x00, 0x00, 0x7f, 0x00, 0x00, 0x01, 0x7f, 0x00, 0x00, 0x01, 0x12, 0x34, 0x56, 0x78}
	datagram := append([]byte{0x00, 0x00, 0x00, 0x02}, ipv4...)

	if _, err := unix.Write(fds[1], datagram); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	select {
	case <-dispatcher.got:
	case <-time.After(5 * time.Second):
		t.Fatal("packet not dispatched; 4-byte TUN header not stripped (offset not applied?)")
	}

	if dispatcher.proto != header.IPv4ProtocolNumber {
		t.Errorf("protocol = %d, want IPv4", dispatcher.proto)
	}
	// The 4-byte utun header must be stripped: the injected packet has to
	// start with the IP packet itself. (The iobased read path may append
	// up to `offset` stale buffer bytes after the packet.)
	if !bytes.HasPrefix(dispatcher.data, ipv4) {
		t.Errorf("injected packet = %x, want prefix %x (4-byte TUN header must be stripped)", dispatcher.data, ipv4)
	}
}

func TestParseFDNoOffsetOnOtherPlatforms(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "ios" {
		t.Skip("no-offset behavior does not apply on darwin/ios")
	}

	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_DGRAM, 0)
	if err != nil {
		t.Fatalf("Socketpair() error = %v", err)
	}
	defer unix.Close(fds[1])

	if err := unix.SetNonblock(fds[0], true); err != nil {
		t.Fatalf("SetNonblock() error = %v", err)
	}

	dev, err := parseFDURL(fds[0])
	if err != nil {
		t.Fatalf("parseFD() error = %v", err)
	}
	defer dev.Close()

	dispatcher := &fdDispatcher{got: make(chan struct{}, 1)}
	dev.Attach(dispatcher)

	ipv4 := []byte{0x45, 0x00, 0x00, 0x18, 0x00, 0x00, 0x00, 0x00, 0x40, 0x11, 0x00, 0x00, 0x7f, 0x00, 0x00, 0x01, 0x7f, 0x00, 0x00, 0x01, 0x12, 0x34, 0x56, 0x78}
	if _, err := unix.Write(fds[1], ipv4); err != nil {
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

func parseFDURL(fd int) (device.Device, error) {
	u := &url.URL{Scheme: fdbased.Driver, Host: strconv.Itoa(fd)}
	return parseFD(u, 1500)
}
