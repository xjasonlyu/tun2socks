//go:build linux

package tun

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/checksum"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type closeTestDispatcher struct{ received chan struct{} }

func (d *closeTestDispatcher) DeliverNetworkPacket(tcpip.NetworkProtocolNumber, *stack.PacketBuffer) {
	select {
	case d.received <- struct{}{}:
	default:
	}
}
func (*closeTestDispatcher) DeliverLinkPacket(tcpip.NetworkProtocolNumber, *stack.PacketBuffer) {}

// A datagram pair exercises the same Readv/ppoll dispatcher as a Linux TUN,
// without requiring root or changing the host's network configuration.
func TestCloseUnblocksIdleLinuxDispatcher(t *testing.T) {
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_DGRAM|unix.SOCK_NONBLOCK, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer unix.Close(fds[1])
	ep, err := fdbased.New(&fdbased.Options{FDs: []int{fds[0]}, MTU: 1500, PacketDispatchMode: fdbased.Readv})
	if err != nil {
		unix.Close(fds[0])
		t.Fatal(err)
	}
	dev := &TUN{LinkEndpoint: ep, fd: fds[0]}
	dispatcher := &closeTestDispatcher{received: make(chan struct{}, 1)}
	dev.Attach(dispatcher)
	packet := []byte{0x45, 0, 0, 28, 0, 0, 0, 0, 64, 17, 0, 0, 127, 0, 0, 1, 127, 0, 0, 1, 4, 210, 22, 46, 0, 8, 0, 0}
	header.IPv4(packet).SetChecksum(^checksum.Checksum(packet[:20], 0))
	if _, err := unix.Write(fds[1], packet); err != nil {
		t.Fatal(err)
	}
	select {
	case <-dispatcher.received:
	case <-time.After(2 * time.Second):
		t.Fatal("dispatcher did not receive the probe")
	}
	// Let the reader return to an idle poll, rather than close between reads.
	time.Sleep(50 * time.Millisecond)
	done := make(chan struct{})
	go func() { dev.Close(); dev.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Close/Wait left the idle dispatcher blocked")
	}

	// Reusing the old descriptor must be safe even if Stop/Close is repeated.
	if err := unix.Dup3(fds[1], fds[0], unix.O_CLOEXEC); err != nil {
		t.Fatal(err)
	}
	defer unix.Close(fds[0])
	var callers sync.WaitGroup
	for range 8 {
		callers.Add(1)
		go func() { defer callers.Done(); dev.Close() }()
	}
	callers.Wait()
	if _, err := unix.FcntlInt(uintptr(fds[0]), unix.F_GETFD, 0); err != nil {
		t.Fatalf("repeated Close closed a reused descriptor: %v", err)
	}
}
