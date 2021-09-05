// Package socks5 provides SOCKS5 client functionalities.
package socks5

// Ref: github.com/Dreamacro/clash/component/socks5

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

// Version is the protocol version as defined in RFC 1928 section 4.
const Version = 0x05

// Command is request commands as defined in RFC 1928 section 4.
type Command = uint8

// SOCKS request commands as defined in RFC 1928 section 4.
const (
	CmdConnect      Command = 0x01
	CmdBind         Command = 0x02
	CmdUDPAssociate Command = 0x03
)

type Atyp = uint8

// SOCKS address types as defined in RFC 1928 section 5.
const (
	AtypIPv4       Atyp = 0x01
	AtypDomainName Atyp = 0x03
	AtypIPv6       Atyp = 0x04
)

// Reply field as defined in RFC 1928 section 6.
type Reply uint8

func (r Reply) String() string {
	switch r {
	case 0x00:
		return "succeeded"
	case 0x01:
		return "general SOCKS server failure"
	case 0x02:
		return "connection not allowed by ruleset"
	case 0x03:
		return "network unreachable"
	case 0x04:
		return "host unreachable"
	case 0x05:
		return "connection refused"
	case 0x06:
		return "TTL expired"
	case 0x07:
		return "command not supported"
	case 0x08:
		return "address type not supported"
	default:
		return "unassigned"
	}
}

// MaxAddrLen is the maximum size of SOCKS address in bytes.
const MaxAddrLen = 1 + 1 + 255 + 2

// MaxAuthLen is the maximum size of user/password field in SOCKS auth.
const MaxAuthLen = 255

// Addr represents a SOCKS address as defined in RFC 1928 section 5.
type Addr []byte

func (a Addr) Valid() bool {
	if len(a) < 1+1+2 /* minimum length */ {
		return false
	}

	switch a[0] {
	case AtypDomainName:
		if len(a) < 1+1+int(a[1])+2 {
			return false
		}
	case AtypIPv4:
		if len(a) < 1+net.IPv4len+2 {
			return false
		}
	case AtypIPv6:
		if len(a) < 1+net.IPv6len+2 {
			return false
		}
	}
	return true
}

// String returns string of socks5.Addr.
func (a Addr) String() string {
	if !a.Valid() {
		return ""
	}

	var host, port string
	switch a[0] {
	case AtypDomainName:
		hostLen := int(a[1])
		host = string(a[2 : 2+hostLen])
		port = strconv.Itoa(int(binary.BigEndian.Uint16(a[2+hostLen:])))
	case AtypIPv4:
		host = net.IP(a[1 : 1+net.IPv4len]).String()
		port = strconv.Itoa(int(binary.BigEndian.Uint16(a[1+net.IPv4len:])))
	case AtypIPv6:
		host = net.IP(a[1 : 1+net.IPv6len]).String()
		port = strconv.Itoa(int(binary.BigEndian.Uint16(a[1+net.IPv6len:])))
	}
	return net.JoinHostPort(host, port)
}

// UDPAddr converts a socks5.Addr to *net.UDPAddr.
func (a Addr) UDPAddr() *net.UDPAddr {
	if !a.Valid() {
		return nil
	}

	var ip []byte
	var port int
	switch a[0] {
	case AtypDomainName /* unsupported */ :
		return nil
	case AtypIPv4:
		ip = make([]byte, net.IPv4len)
		copy(ip, a[1:1+net.IPv4len])
		port = int(binary.BigEndian.Uint16(a[1+net.IPv4len:]))
	case AtypIPv6:
		ip = make([]byte, net.IPv6len)
		copy(ip, a[1:1+net.IPv6len])
		port = int(binary.BigEndian.Uint16(a[1+net.IPv6len:]))
	}
	return &net.UDPAddr{IP: ip, Port: port}
}

// User provides basic socks5 auth functionality.
type User struct {
	Username string
	Password string
}

// ClientHandshake fast-tracks SOCKS initialization to get target address to connect on client side.
func ClientHandshake(rw io.ReadWriter, addr Addr, command Command, user *User) (Addr, error) {
	buf := make([]byte, MaxAddrLen)

	var method uint8
	if user != nil {
		method = 0x02 /* USERNAME/PASSWORD */
	} else {
		method = 0x00 /* NO AUTHENTICATION REQUIRED */
	}

	// VER, NMETHODS, METHODS
	if _, err := rw.Write([]byte{Version, 0x01 /* NMETHODS */, method}); err != nil {
		return nil, err
	}

	// VER, METHOD
	if _, err := io.ReadFull(rw, buf[:2]); err != nil {
		return nil, err
	}

	if buf[0] != Version {
		return nil, errors.New("socks version mismatched")
	}

	if buf[1] == 0x02 /* USERNAME/PASSWORD */ {
		if user == nil {
			return nil, errors.New("auth required")
		}

		// password protocol version
		authMsg := &bytes.Buffer{}
		authMsg.WriteByte(0x01 /* VER */)
		authMsg.WriteByte(byte(len(user.Username)) /* ULEN */)
		authMsg.WriteString(user.Username /* UNAME */)
		authMsg.WriteByte(byte(len(user.Password)) /* PLEN */)
		authMsg.WriteString(user.Password /* PASSWD */)

		if len(authMsg.Bytes()) > MaxAuthLen {
			return nil, errors.New("auth message too long")
		}

		if _, err := rw.Write(authMsg.Bytes()); err != nil {
			return nil, err
		}

		if _, err := io.ReadFull(rw, buf[:2]); err != nil {
			return nil, err
		}

		if buf[1] != 0x00 /* STATUS of SUCCESS */ {
			return nil, errors.New("rejected username/password")
		}

	} else if buf[1] != 0x00 /* NO AUTHENTICATION REQUIRED */ {
		return nil, errors.New("unsupported method")
	}

	// VER, CMD, RSV, ADDR
	if _, err := rw.Write(bytes.Join([][]byte{{Version, command, 0x00 /* RSV */}, addr}, nil)); err != nil {
		return nil, err
	}

	// VER, REP, RSV
	if _, err := io.ReadFull(rw, buf[:3]); err != nil {
		return nil, err
	}

	if rep := Reply(buf[1]); rep != 0x00 /* SUCCEEDED */ {
		return nil, fmt.Errorf("%#02x: %s", uint8(rep), rep)
	}

	return ReadAddr(rw, buf)
}

func ReadAddr(r io.Reader, b []byte) (Addr, error) {
	if len(b) < MaxAddrLen {
		return nil, io.ErrShortBuffer
	}

	// read 1st byte for address type
	if _, err := io.ReadFull(r, b[:1]); err != nil {
		return nil, err
	}

	switch b[0] /* ATYP */ {
	case AtypDomainName:
		// read 2nd byte for domain length
		if _, err := io.ReadFull(r, b[1:2]); err != nil {
			return nil, err
		}
		domainLength := uint16(b[1])
		_, err := io.ReadFull(r, b[2:2+domainLength+2])
		return b[:1+1+domainLength+2], err
	case AtypIPv4:
		_, err := io.ReadFull(r, b[1:1+net.IPv4len+2])
		return b[:1+net.IPv4len+2], err
	case AtypIPv6:
		_, err := io.ReadFull(r, b[1:1+net.IPv6len+2])
		return b[:1+net.IPv6len+2], err
	default:
		return nil, errors.New("invalid address type")
	}
}

// SplitAddr slices a SOCKS address from beginning of b. Returns nil if failed.
func SplitAddr(b []byte) Addr {
	addrLen := 1
	if len(b) < addrLen {
		return nil
	}

	switch b[0] {
	case AtypDomainName:
		if len(b) < 2 {
			return nil
		}
		addrLen = 1 + 1 + int(b[1]) + 2
	case AtypIPv4:
		addrLen = 1 + net.IPv4len + 2
	case AtypIPv6:
		addrLen = 1 + net.IPv6len + 2
	default:
		return nil
	}

	if len(b) < addrLen {
		return nil
	}

	return b[:addrLen]
}

// ParseAddr parses the address in string s. Returns nil if failed.
func ParseAddr(s string) Addr {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return nil
	}

	var addr Addr
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			addr = make([]byte, 1+net.IPv4len+2)
			addr[0] = AtypIPv4
			copy(addr[1:], ip4)
		} else {
			addr = make([]byte, 1+net.IPv6len+2)
			addr[0] = AtypIPv6
			copy(addr[1:], ip)
		}
	} else {
		if len(host) > 255 {
			return nil
		}
		addr = make([]byte, 1+1+len(host)+2)
		addr[0] = AtypDomainName
		addr[1] = byte(len(host))
		copy(addr[2:], host)
	}

	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil
	}
	binary.BigEndian.PutUint16(addr[len(addr)-2:], uint16(p))

	return addr
}

// ParseAddrToSocksAddr parse a socks addr from net.addr
// This is a fast path of ParseAddr(addr.String())
func ParseAddrToSocksAddr(addr net.Addr) Addr {
	var ip net.IP
	var port int
	if udpAddr, ok := addr.(*net.UDPAddr); ok {
		ip = udpAddr.IP
		port = udpAddr.Port
	} else if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		ip = tcpAddr.IP
		port = tcpAddr.Port
	}

	// fallback parse
	if ip == nil {
		return ParseAddr(addr.String())
	}

	var parsed Addr
	if ip4 := ip.To4(); ip4 != nil {
		parsed = make([]byte, 1+net.IPv4len+2)
		parsed[0] = AtypIPv4
		copy(parsed[1:], ip4)
		binary.BigEndian.PutUint16(parsed[1+net.IPv4len:], uint16(port))
	} else {
		parsed = make([]byte, 1+net.IPv6len+2)
		parsed[0] = AtypIPv6
		copy(parsed[1:], ip)
		binary.BigEndian.PutUint16(parsed[1+net.IPv6len:], uint16(port))
	}
	return parsed
}

// DecodeUDPPacket split `packet` to addr payload, and this function is mutable with `packet`
func DecodeUDPPacket(packet []byte) (addr Addr, payload []byte, err error) {
	if len(packet) < 5 {
		err = errors.New("insufficient length of packet")
		return
	}

	// packet[0] and packet[1] are reserved
	if !bytes.Equal(packet[:2], []byte{0x00, 0x00}) {
		err = errors.New("reserved fields should be zero")
		return
	}

	if packet[2] != 0x00 /* fragments */ {
		err = errors.New("discarding fragmented payload")
		return
	}

	addr = SplitAddr(packet[3:])
	if addr == nil {
		err = errors.New("socks5 UDP addr is nil")
	}

	payload = packet[3+len(addr):]
	return
}

func EncodeUDPPacket(addr Addr, payload []byte) (packet []byte, err error) {
	if addr == nil {
		return nil, errors.New("address is invalid")
	}
	packet = bytes.Join([][]byte{{0x00, 0x00, 0x00}, addr, payload}, nil)
	return
}
