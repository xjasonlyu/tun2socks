// Package socks4 provides SOCKS4/SOCKS4A client functionalities.
package socks4

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/netip"
	"strconv"

	"github.com/xjasonlyu/tun2socks/v2/transport/internal/bufferpool"
)

const Version = 0x04

type Command = uint8

const (
	CmdConnect Command = 0x01
	CmdBind    Command = 0x02
)

type Code = uint8

const (
	RequestGranted          Code = 90
	RequestRejected         Code = 91
	RequestIdentdFailed     Code = 92
	RequestIdentdMismatched Code = 93
)

var (
	errVersionMismatched = errors.New("version code mismatched")
	errIPv6NotSupported  = errors.New("IPv6 not supported")
	errCmdNotSupported   = errors.New("command not supported")

	ErrRequestRejected         = errors.New("request rejected or failed")
	ErrRequestIdentdFailed     = errors.New("request rejected because SOCKS server cannot connect to identd on the client")
	ErrRequestIdentdMismatched = errors.New("request rejected because the client program and identd report different user-ids")
	ErrRequestUnknownCode      = errors.New("request failed with unknown code")
)

func ClientHandshake(rw io.ReadWriter, addr string, command Command, userID string) (err error) {
	if command == CmdBind {
		return errCmdNotSupported
	}

	var (
		host string
		port uint16
	)
	if host, port, err = splitHostPort(addr); err != nil {
		return err
	}

	ip, _ := netip.ParseAddr(host)
	switch {
	case !ip.IsValid(): /* HOST */
		ip = netip.AddrFrom4([4]byte{0, 0, 0, 1})
	case ip.Is4In6(): /* IPv4-mapped IPv6 */
		ip = netip.AddrFrom4(ip.As4())
	case ip.Is4(): /* IPv4 */
	case ip.Is6(): /* IPv6 */
		return errIPv6NotSupported
	}

	req := bufferpool.Get()
	defer bufferpool.Put(req)
	req.WriteByte(Version)
	req.WriteByte(command)
	_ = binary.Write(req, binary.BigEndian, port)
	req.Write(ip.AsSlice())
	req.WriteString(userID)
	req.WriteByte(0x00) /* NULL */

	if isReservedIP(ip) /* SOCKS4A */ {
		req.WriteString(host)
		req.WriteByte(0) /* NULL */
	}

	if _, err = rw.Write(req.Bytes()); err != nil {
		return err
	}

	var resp [8]byte
	if _, err = io.ReadFull(rw, resp[:]); err != nil {
		return err
	}

	if resp[0] != 0x00 {
		return errVersionMismatched
	}

	switch resp[1] {
	case RequestGranted:
		return nil
	case RequestRejected:
		return ErrRequestRejected
	case RequestIdentdFailed:
		return ErrRequestIdentdFailed
	case RequestIdentdMismatched:
		return ErrRequestIdentdMismatched
	default:
		return ErrRequestUnknownCode
	}
}

// For version 4A, if the client cannot resolve the destination host's
// domain name to find its IP address, it should set the first three bytes
// of DSTIP to NULL and the last byte to a non-zero value. (This corresponds
// to IP address 0.0.0.x, with x nonzero. As decreed by IANA  -- The
// Internet Assigned Numbers Authority -- such an address is inadmissible
// as a destination IP address and thus should never occur if the client
// can resolve the domain name.)
func isReservedIP(ip netip.Addr) bool {
	prefix := netip.PrefixFrom(netip.IPv4Unspecified(), 24)
	return !ip.IsUnspecified() && prefix.Contains(ip)
}

func splitHostPort(addr string) (string, uint16, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}

	portInt, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return "", 0, err
	}

	return host, uint16(portInt), nil
}
