// Package iobased provides the implementation of io.ReadWriter
// based data-link layer endpoints.
package iobased

import (
	"context"
	"errors"
	"io"
	"sync"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

const (
	// Queue length for outbound packet, arriving for read. Overflow
	// causes packet drops.
	defaultOutQueueLen = 1 << 10
)

// Endpoint implements the interface of stack.LinkEndpoint from io.ReadWriter.
type Endpoint struct {
	*channel.Endpoint

	// rw is the io.ReadWriter for reading and writing packets.
	rw io.ReadWriter

	// mtu (maximum transmission unit) is the maximum size of a packet.
	mtu uint32

	// once is used to perform the init action once when attaching.
	once sync.Once
}

// New returns stack.LinkEndpoint(.*Endpoint) and error.
func New(rw io.ReadWriter, mtu uint32) (*Endpoint, error) {
	if mtu == 0 {
		return nil, errors.New("MTU size is zero")
	}

	if rw == nil {
		return nil, errors.New("RW interface is nil")
	}

	return &Endpoint{
		Endpoint: channel.New(defaultOutQueueLen, mtu, ""),
		rw:       rw,
		mtu:      mtu,
	}, nil
}

// Attach launches the goroutine that reads packets from io.Reader and
// dispatches them via the provided dispatcher.
func (e *Endpoint) Attach(dispatcher stack.NetworkDispatcher) {
	e.once.Do(func() {
		go e.dispatchLoop()
		go e.outboundLoop()
	})
	e.Endpoint.Attach(dispatcher)
}

// dispatchLoop dispatches packets to upper layer.
func (e *Endpoint) dispatchLoop() {
	for {
		data := make([]byte, e.MTU())

		n, err := e.rw.Read(data)
		if err != nil {
			break
		}

		if !e.IsAttached() {
			continue /* unattached, drop packet */
		}

		pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
			Data: buffer.View(data[:n]).ToVectorisedView(),
		})

		switch header.IPVersion(data) {
		case header.IPv4Version:
			e.InjectInbound(header.IPv4ProtocolNumber, pkt)
		case header.IPv6Version:
			e.InjectInbound(header.IPv6ProtocolNumber, pkt)
		}
		pkt.DecRef() /* release */
	}
}

// outboundLoop reads outbound packets from channel, and then it calls
// writePacket to send those packets back to lower layer.
func (e *Endpoint) outboundLoop() {
	// TODO: support cancel() in the future.
	ctx := context.Background()

	for {
		pkt := e.ReadContext(ctx)
		if pkt == nil {
			break
		}
		e.writePacket(pkt)
	}
}

// writePacket writes outbound packets to the io.Writer.
func (e *Endpoint) writePacket(pkt *stack.PacketBuffer) tcpip.Error {
	vView := buffer.NewVectorisedView(
		pkt.Size(),
		pkt.Views(),
	)
	if _, err := e.rw.Write(vView.ToView()); err != nil {
		return &tcpip.ErrInvalidEndpointState{}
	}
	return nil
}
