package metadata

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"

	"github.com/xjasonlyu/tun2socks/v2/transport/socks5"
)

const (
	TCP Network = iota
	UDP
)

type Network uint8

func (n Network) String() string {
	switch n {
	case TCP:
		return "tcp"
	case UDP:
		return "udp"
	default:
		return fmt.Sprintf("network(%d)", n)
	}
}

func (n Network) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

// Metadata implements the net.Addr interface.
type Metadata struct {
	Net     Network `json:"network"`
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

func (m *Metadata) UDPAddr() *net.UDPAddr {
	if m.Net != UDP || m.DstIP == nil {
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

func (m *Metadata) Network() string {
	return m.Net.String()
}

// String returns destination address of this metadata.
// Also, it implements net.Addr interface.
func (m *Metadata) String() string {
	return m.DestinationAddress()
}
