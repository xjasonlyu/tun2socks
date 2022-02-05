package metadata

import (
	"bytes"
	"encoding/binary"
	"net"
	"strconv"

	"github.com/xjasonlyu/tun2socks/v2/transport/socks5"
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
	if udpAddr := m.UDPAddr(); udpAddr != nil {
		return udpAddr
	}
	return &Addr{metadata: m}
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

func (m *Metadata) SerializeSocksAddr() socks5.Addr {
	var (
		buf  [][]byte
		port [2]byte
	)
	binary.BigEndian.PutUint16(port[:], m.DstPort)

	if m.DstIP.To4() != nil /* IPv4 */ {
		aType := socks5.AtypIPv4
		buf = [][]byte{{aType}, m.DstIP.To4(), port[:]}
	} else /* IPv6 */ {
		aType := socks5.AtypIPv6
		buf = [][]byte{{aType}, m.DstIP.To16(), port[:]}
	}
	return bytes.Join(buf, nil)
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
