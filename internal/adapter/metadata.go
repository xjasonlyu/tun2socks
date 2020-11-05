package adapter

import (
	"bytes"
	"encoding/json"
	"net"
	"strconv"

	"github.com/xjasonlyu/clash/component/socks5"
)

const (
	TCP Network = iota
	UDP
)

type Network int

func (n Network) String() string {
	if n == TCP {
		return "tcp"
	}
	return "udp"
}

func (n Network) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
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
	Host    string  `json:"host"`
}

func (m *Metadata) DestinationAddress() string {
	if m.Host != "" {
		return net.JoinHostPort(m.Host, strconv.FormatUint(uint64(m.DstPort), 10))
	} else if m.DstIP != nil {
		return net.JoinHostPort(m.DstIP.String(), strconv.FormatUint(uint64(m.DstPort), 10))
	}
	return ""
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

func (m *Metadata) SerializesSocksAddr() socks5.Addr {
	var buf [][]byte
	port := []byte{uint8(m.DstPort >> 8), uint8(m.DstPort & 0xff)}
	switch {
	case m.Host != "": /* Domain */
		aType := socks5.AtypDomainName
		length := uint8(len(m.Host))
		host := []byte(m.Host)
		buf = [][]byte{{uint8(aType), length}, host, port}
	case m.DstIP.To4() != nil: /* IPv4 */
		aType := socks5.AtypIPv4
		host := m.DstIP.To4()
		buf = [][]byte{{uint8(aType)}, host, port}
	default: /* IPv6 */
		aType := socks5.AtypIPv6
		host := m.DstIP.To16()
		buf = [][]byte{{uint8(aType)}, host, port}
	}
	return bytes.Join(buf, nil)
}

func (m *Metadata) Network() string {
	return m.Net.String()
}

func (m *Metadata) String() string {
	return m.DestinationAddress()
}

func (m *Metadata) Valid() bool {
	return m.Host != "" || m.DstIP != nil
}
