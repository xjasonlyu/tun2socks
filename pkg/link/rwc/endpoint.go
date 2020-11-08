package rwc

import (
	"errors"
	"io"
	"sync"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// Endpoint wraps io.ReadWriter to stack.LinkEndpoint.
type Endpoint struct {
	mtu uint32
	rwc io.ReadWriteCloser

	wg sync.WaitGroup

	dispatcher         stack.NetworkDispatcher
	LinkEPCapabilities stack.LinkEndpointCapabilities
}

// New returns stack.LinkEndpoint(.*Endpoint) and error.
func New(rwc io.ReadWriteCloser, mtu uint32) (*Endpoint, error) {
	switch {
	case mtu == 0:
		return nil, errors.New("MTU size is zero")
	case rwc == nil:
		return nil, errors.New("RWC interface is nil")
	}

	return &Endpoint{
		rwc: rwc,
		mtu: mtu,
	}, nil
}

// Attach launches the goroutine that reads packets from io.ReadWriter and
// dispatches them via the provided dispatcher.
func (e *Endpoint) Attach(dispatcher stack.NetworkDispatcher) {
	go e.dispatchLoop()
	e.dispatcher = dispatcher
}

// IsAttached implements stack.LinkEndpoint.IsAttached.
func (e *Endpoint) IsAttached() bool {
	return e.dispatcher != nil
}

// dispatchLoop dispatches packets to upper layer.
func (e *Endpoint) dispatchLoop() {
	e.wg.Add(1)
	defer e.wg.Done()

	for {
		packet := make([]byte, e.mtu)
		n, err := e.rwc.Read(packet)
		if err != nil {
			break
		}

		if !e.IsAttached() {
			continue
		}

		var p tcpip.NetworkProtocolNumber
		switch header.IPVersion(packet) {
		case header.IPv4Version:
			p = header.IPv4ProtocolNumber
		case header.IPv6Version:
			p = header.IPv6ProtocolNumber
		}

		e.dispatcher.DeliverNetworkPacket("", "", p, &stack.PacketBuffer{
			Data: buffer.View(packet[:n]).ToVectorisedView(),
		})
	}
}

// writePacket writes packets back into io.ReadWriter.
func (e *Endpoint) writePacket(pkt *stack.PacketBuffer) *tcpip.Error {
	networkHdr := pkt.NetworkHeader().View()
	transportHdr := pkt.TransportHeader().View()
	payload := pkt.Data.ToView()

	buf := buffer.NewVectorisedView(
		len(networkHdr)+len(transportHdr)+len(payload),
		[]buffer.View{networkHdr, transportHdr, payload},
	)

	if _, err := e.rwc.Write(buf.ToView()); err != nil {
		return tcpip.ErrInvalidEndpointState
	}

	return nil
}

// WritePacket writes packets back into io.ReadWriter.
func (e *Endpoint) WritePacket(_ *stack.Route, _ *stack.GSO, _ tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) *tcpip.Error {
	return e.writePacket(pkt)
}

// WritePackets writes packets back into io.ReadWriter.
func (e *Endpoint) WritePackets(_ *stack.Route, _ *stack.GSO, pkts stack.PacketBufferList, _ tcpip.NetworkProtocolNumber) (int, *tcpip.Error) {
	n := 0
	for pkt := pkts.Front(); pkt != nil; pkt = pkt.Next() {
		if err := e.writePacket(pkt); err != nil {
			break
		}
		n++
	}
	return n, nil
}

// WriteRawPacket implements stack.LinkEndpoint.WriteRawPacket.
func (e *Endpoint) WriteRawPacket(vv buffer.VectorisedView) *tcpip.Error {
	pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
		Data: vv,
	})
	return e.writePacket(pkt)
}

// MTU implements stack.LinkEndpoint.MTU.
func (e *Endpoint) MTU() uint32 {
	return e.mtu
}

// Capabilities implements stack.LinkEndpoint.Capabilities.
func (e *Endpoint) Capabilities() stack.LinkEndpointCapabilities {
	return e.LinkEPCapabilities
}

// GSOMaxSize returns the maximum GSO packet size.
func (*Endpoint) GSOMaxSize() uint32 {
	return 1 << 15
}

// MaxHeaderLength returns the maximum size of the link layer header. Given it
// doesn't have a header, it just returns 0.
func (*Endpoint) MaxHeaderLength() uint16 {
	return 0
}

// LinkAddress returns the link address of this endpoint.
func (*Endpoint) LinkAddress() tcpip.LinkAddress {
	return ""
}

// ARPHardwareType implements stack.LinkEndpoint.ARPHardwareType.
func (*Endpoint) ARPHardwareType() header.ARPHardwareType {
	return header.ARPHardwareNone
}

// AddHeader implements stack.LinkEndpoint.AddHeader.
func (e *Endpoint) AddHeader(tcpip.LinkAddress, tcpip.LinkAddress, tcpip.NetworkProtocolNumber, *stack.PacketBuffer) {
}

// Wait implements stack.LinkEndpoint.Wait.
func (e *Endpoint) Wait() {
	e.wg.Wait()
}

// Close closes io.ReadWriteCloser.
func (e *Endpoint) Close() error {
	return e.rwc.Close()
}
