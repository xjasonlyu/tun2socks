package metadata

import (
	"net"
	"net/netip"
)

// Metadata contains metadata of transport protocol sessions.
type Metadata struct {
	Network Network    `json:"network"`
	SrcIP   netip.Addr `json:"sourceIP"`
	MidIP   netip.Addr `json:"dialerIP"`
	DstIP   netip.Addr `json:"destinationIP"`
	SrcPort uint16     `json:"sourcePort"`
	MidPort uint16     `json:"dialerPort"`
	DstPort uint16     `json:"destinationPort"`
}

func (m *Metadata) DestinationAddrPort() netip.AddrPort {
	return netip.AddrPortFrom(m.DstIP, m.DstPort)
}

func (m *Metadata) DestinationAddress() string {
	return m.DestinationAddrPort().String()
}

func (m *Metadata) SourceAddrPort() netip.AddrPort {
	return netip.AddrPortFrom(m.SrcIP, m.SrcPort)
}

func (m *Metadata) SourceAddress() string {
	return m.SourceAddrPort().String()
}

func (m *Metadata) Addr() net.Addr {
	return &Addr{metadata: m}
}

func (m *Metadata) TCPAddr() *net.TCPAddr {
	if m.Network != TCP || !m.DstIP.IsValid() {
		return nil
	}
	return net.TCPAddrFromAddrPort(m.DestinationAddrPort())
}

func (m *Metadata) UDPAddr() *net.UDPAddr {
	if m.Network != UDP || !m.DstIP.IsValid() {
		return nil
	}
	return net.UDPAddrFromAddrPort(m.DestinationAddrPort())
}

// Addr implements the net.Addr interface.
type Addr struct {
	metadata *Metadata
}

func (a *Addr) Metadata() *Metadata {
	return a.metadata
}

func (a *Addr) Network() string {
	return a.metadata.Network.String()
}

func (a *Addr) String() string {
	return a.metadata.DestinationAddress()
}
