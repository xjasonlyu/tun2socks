// Package iobased provides the implementation of io.ReadWriter
// based data-link layer endpoints.
package iobased

import (
	"context"
	"errors"
	"io"
	"sync"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

const (
	// Queue length for outbound packet, arriving for read. Overflow
	// causes packet drops. Reduced from 1024 to limit queued packet
	// buffer memory under bursty load.
	defaultOutQueueLen = 256
)

// Endpoint implements the interface of stack.LinkEndpoint from io.ReadWriter.
type Endpoint struct {
	*channel.Endpoint

	// rw is the io.ReadWriter for reading and writing packets.
	rw io.ReadWriter

	// mtu (maximum transmission unit) is the maximum size of a packet.
	mtu uint32

	// offset can be useful when perform TUN device I/O with TUN_PI enabled.
	offset int

	// once is used to perform the init action once when attaching.
	once sync.Once

	// wg keeps track of running goroutines.
	wg sync.WaitGroup

	// readPool recycles read buffers to avoid per-packet heap allocation.
	readPool sync.Pool

	// writePool recycles write buffers to avoid per-packet heap allocation.
	writePool sync.Pool
}

// New returns stack.LinkEndpoint(.*Endpoint) and error.
func New(rw io.ReadWriter, mtu uint32, offset int) (*Endpoint, error) {
	if mtu == 0 {
		return nil, errors.New("MTU size is zero")
	}

	if rw == nil {
		return nil, errors.New("RW interface is nil")
	}

	if offset < 0 {
		return nil, errors.New("offset must be non-negative")
	}

	bufSize := offset + int(mtu)
	return &Endpoint{
		Endpoint: channel.New(defaultOutQueueLen, mtu, ""),
		rw:       rw,
		mtu:      mtu,
		offset:   offset,
		readPool: sync.Pool{New: func() any {
			b := make([]byte, bufSize)
			return &b
		}},
		writePool: sync.Pool{New: func() any {
			b := make([]byte, 0, bufSize)
			return &b
		}},
	}, nil
}

// Attach launches the goroutine that reads packets from io.Reader and
// dispatches them via the provided dispatcher.
func (e *Endpoint) Attach(dispatcher stack.NetworkDispatcher) {
	e.Endpoint.Attach(dispatcher)
	e.once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		e.wg.Add(2)
		go func() {
			e.outboundLoop(ctx)
			e.wg.Done()
		}()
		go func() {
			e.dispatchLoop(cancel)
			e.wg.Done()
		}()
	})
}

func (e *Endpoint) Wait() {
	e.wg.Wait()
}

// dispatchLoop dispatches packets to upper layer.
func (e *Endpoint) dispatchLoop(cancel context.CancelFunc) {
	defer cancel()

	offset, mtu := e.offset, int(e.mtu)

	for {
		bufp := e.readPool.Get().(*[]byte)
		data := *bufp

		n, err := e.rw.Read(data)
		if err != nil {
			e.readPool.Put(bufp)
			break
		}

		if n == 0 || n > mtu {
			e.readPool.Put(bufp)
			continue
		}

		if !e.IsAttached() {
			e.readPool.Put(bufp)
			continue
		}

		// Determine IP version before returning the buffer to the pool.
		ipVer := header.IPVersion(data[offset:])

		// MakeWithData copies the slice content, so we can return the
		// buffer to the pool immediately after this call.
		pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
			Payload: buffer.MakeWithData(data[offset : offset+n]),
		})
		e.readPool.Put(bufp)

		switch ipVer {
		case header.IPv4Version:
			e.InjectInbound(header.IPv4ProtocolNumber, pkt)
		case header.IPv6Version:
			e.InjectInbound(header.IPv6ProtocolNumber, pkt)
		}
		pkt.DecRef()
	}
}

// outboundLoop reads outbound packets from channel, and then it calls
// writePacket to send those packets back to lower layer.
func (e *Endpoint) outboundLoop(ctx context.Context) {
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
	defer pkt.DecRef()

	pktBuf := pkt.ToBuffer()
	defer pktBuf.Release()

	size := e.offset + int(pktBuf.Size())

	bufp := e.writePool.Get().(*[]byte)
	writeBuf := (*bufp)[:0]

	if cap(writeBuf) < size {
		writeBuf = make([]byte, 0, size)
	}

	// Prepend offset zero-bytes (TUN_PI header) if needed.
	for i := 0; i < e.offset; i++ {
		writeBuf = append(writeBuf, 0)
	}

	views := pktBuf.AsViewList()
	for v := views.Front(); v != nil; v = v.Next() {
		writeBuf = append(writeBuf, v.AsSlice()...)
	}

	_, err := e.rw.Write(writeBuf)

	*bufp = writeBuf[:0]
	e.writePool.Put(bufp)

	if err != nil {
		return &tcpip.ErrInvalidEndpointState{}
	}
	return nil
}
