//go:build unix && !linux

package fdbased

import (
	"strconv"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type testDispatcher struct {
	got chan struct{}
}

func (d *testDispatcher) DeliverNetworkPacket(tcpip.NetworkProtocolNumber, *stack.PacketBuffer) {
	select {
	case d.got <- struct{}{}:
	default:
	}
}

func (d *testDispatcher) DeliverLinkPacket(tcpip.NetworkProtocolNumber, *stack.PacketBuffer) {}

func TestCloseUnblocksDispatchLoop(t *testing.T) {
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_DGRAM, 0)
	if err != nil {
		t.Fatalf("Socketpair() error = %v", err)
	}
	defer unix.Close(fds[1])

	// os.NewFile only registers the fd with the runtime netpoller when the
	// fd is non-blocking, which is the iobased path this fix targets.
	if err := unix.SetNonblock(fds[0], true); err != nil {
		t.Fatalf("SetNonblock() error = %v", err)
	}

	dev, err := Open(strconv.Itoa(fds[0]), 1500, 0)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	// Attach a dispatcher and send a probe packet so we know the dispatch
	// loop is running and then blocked on the next Read.
	dispatcher := &testDispatcher{got: make(chan struct{}, 1)}
	dev.Attach(dispatcher)

	ipv4 := []byte{0x45, 0x00, 0x00, 0x14, 0x00, 0x00, 0x00, 0x00, 0x40, 0x11, 0x00, 0x00, 0x7f, 0x00, 0x00, 0x01, 0x7f, 0x00, 0x00, 0x01}
	if _, err := unix.Write(fds[1], ipv4); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	select {
	case <-dispatcher.got:
	case <-time.After(5 * time.Second):
		t.Fatal("probe packet not dispatched; dispatch loop not running")
	}

	// Close the device and verify Wait() returns instead of hanging.
	done := make(chan struct{})
	go func() {
		dev.Wait()
		close(done)
	}()

	dev.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Wait() did not return after Close(); dispatchLoop still blocked")
	}
}
