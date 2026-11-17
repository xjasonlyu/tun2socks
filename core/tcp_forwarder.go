package core

import (
	"sync"

	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
)

// gVisor's ForwarderRequest exposes the tuple, but not the initial sequence
// number. Preserve the first SYN's ISN at admission, without changing gVisor.
// Keep a pending record through completion (including RST output), so neither a
// retransmission nor a new SYN can overwrite an outstanding request's key.
type trackedTCPForwarder struct {
	mu        sync.Mutex
	pending   map[stack.TransportEndpointID]tcpFailureKey
	forwarder *tcp.Forwarder
	stack     *stack.Stack
	limit     int
}

func newTrackedTCPForwarder(s *stack.Stack, window, limit int, handler func(*tcp.ForwarderRequest, tcpFailureKey, func(bool))) *trackedTCPForwarder {
	f := &trackedTCPForwarder{pending: make(map[stack.TransportEndpointID]tcpFailureKey), stack: s, limit: limit}
	f.forwarder = tcp.NewForwarder(s, window, limit, func(r *tcp.ForwarderRequest) {
		id := r.ID()
		f.mu.Lock()
		key := f.pending[id]
		f.mu.Unlock()
		handler(r, key, func(reset bool) {
			// Complete may synchronously write a RST. Do not hold the global
			// admission lock while output is blocked. Our pending record keeps
			// this tuple reserved after gVisor removes its own in-flight entry.
			r.Complete(reset)
			f.mu.Lock()
			delete(f.pending, id)
			f.mu.Unlock()
		})
	})
	return f
}

func (f *trackedTCPForwarder) HandlePacket(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
	segment := pkt.TransportHeader().Slice()
	if len(segment) < header.TCPMinimumSize {
		return false
	}
	h := header.TCP(segment)
	if !h.Flags().Contains(header.TCPFlagSyn) || h.Flags().Contains(header.TCPFlagAck) {
		return false
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	_, pending := f.pending[id]
	if pending {
		return true // Still dialing or completing; do not admit this tuple again.
	}
	if len(f.pending) >= f.limit {
		f.stack.Stats().TCP.ForwardMaxInFlightDrop.Increment()
		return true
	}
	f.pending[id] = tcpFailureKey{id: id, isn: h.SequenceNumber()}
	// gVisor validates each newly admitted SYN and its checksum.
	handled := f.forwarder.HandlePacket(id, pkt)
	if !handled {
		delete(f.pending, id)
	}
	return handled
}
