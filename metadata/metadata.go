package metadata

import (
	"net"
	"strconv"
)

// Metadata contains metadata of transport protocol sessions.
type Metadata struct {
	Network Network `json:"network"`
	SrcIP   net.IP  `json:"sourceIP"`
	MidIP   net.IP  `json:"dialerIP"`
	DstIP   net.IP  `json:"destinationIP"`
	SrcPort uint16  `json:"sourcePort"`
	MidPort uint16  `json:"dialerPort"`
	DstPort uint16  `json:"destinationPort"`
}

func (m *Metadata) DestinationAddress() string {
	return net.JoinHostPort(m.DstIP.String(), strconv.FormatUint(uint64(m.DstPort), 10))
}

func (m *Metadata) SourceAddress() string {
	return net.JoinHostPort(m.SrcIP.String(), strconv.FormatUint(uint64(m.SrcPort), 10))
}

func (m *Metadata) Addr() net.Addr {
	return &Addr{metadata: m}
}

func (m *Metadata) TCPAddr() *net.TCPAddr {
	if m.Network != TCP || m.DstIP == nil {
		return nil
	}
	return &net.TCPAddr{
		IP:   m.DstIP,
		Port: int(m.DstPort),
	}
}

func (m *Metadata) UDPAddr() *net.UDPAddr {
	if m.Network != UDP || m.DstIP == nil {
		return nil
	}
	return &net.UDPAddr{
		IP:   m.DstIP,
		Port: int(m.DstPort),
	}
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
