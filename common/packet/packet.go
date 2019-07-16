package packet

import (
	"encoding/binary"
	"net"
)

const (
	IPVERSION_4 = 4
	IPVERSION_6 = 6

	PROTOCOL_ICMP = 1
	PROTOCOL_TCP  = 6
	PROTOCOL_UDP  = 17
)

func PeekIPVersion(data []byte) uint8 {
	return uint8((data[0] & 0xf0) >> 4)
}

func PeekProtocol(data []byte) string {
	switch uint8(data[9]) {
	case PROTOCOL_ICMP:
		return "icmp"
	case PROTOCOL_TCP:
		return "tcp"
	case PROTOCOL_UDP:
		return "udp"
	default:
		return "unknown"
	}
}

func PeekSourceAddress(data []byte) net.IP {
	return net.IP(data[12:16])
}

func PeekSourcePort(data []byte) uint16 {
	ihl := uint8(data[0] & 0x0f)
	return binary.BigEndian.Uint16(data[ihl*4 : ihl*4+2])
}

func PeekDestinationAddress(data []byte) net.IP {
	return net.IP(data[16:20])
}

func PeekDestinationPort(data []byte) uint16 {
	ihl := uint8(data[0] & 0x0f)
	return binary.BigEndian.Uint16(data[ihl*4+2 : ihl*4+4])
}

func IsSYNSegment(data []byte) bool {
	ihl := uint8(data[0] & 0x0f)
	if uint8(data[ihl*4+13]&(1<<1)) == 0 {
		return false
	} else {
		return true
	}
}
