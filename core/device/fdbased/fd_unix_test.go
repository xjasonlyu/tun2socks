//go:build unix

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

	// The iobased path on non-Linux systems needs a nonblocking fd for netpoll.
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

	ipv4 := []byte{
		0x45, 0, 0, 28, 0, 0, 0, 0, 64, 17, 0x7c, 0xcf, 127, 0, 0, 1, 127, 0, 0, 1,
		0x04, 0xd2, 0x16, 0x2e, 0, 8, 0, 0, // UDP header
	}
	if _, err := unix.Write(fds[1], ipv4); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	select {
	case <-dispatcher.got:
	case <-time.After(5 * time.Second):
		t.Fatal("probe packet not dispatched; dispatch loop not running")
	}

	// Let the reader settle into its next idle poll. Keep the peer open so
	// only shutting down the device can wake it.
	time.Sleep(50 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		dev.Close()
		dev.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Wait() did not return after Close(); dispatchLoop still blocked")
	}
}
