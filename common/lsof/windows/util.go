// +build windows

package windows

import (
	"encoding/binary"
	"fmt"
	"net"
	"unicode/utf16"
	"unsafe"
)

func UTF16PtrToString(cstr *uint16) string {
	if cstr != nil {
		us := make([]uint16, 0, 256)
		for p := uintptr(unsafe.Pointer(cstr)); ; p += 2 {
			u := *(*uint16)(unsafe.Pointer(p))
			if u == 0 {
				return string(utf16.Decode(us))
			}
			us = append(us, u)
		}
	}

	return ""
}

func NTOHS(port uint16) uint16 {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, port)
	return binary.LittleEndian.Uint16(buf)
}

// FIXME IPv6
func IPAddrNTOA(addr uint32) string {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, addr)
	return fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3])
}

// FIXME IPv6
func IPAddrATON(addr string) uint32 {
	ip := net.ParseIP(addr)
	if ip == nil {
		panic("invalid IP")
	}
	return binary.BigEndian.Uint32([]byte(ip)[net.IPv6len-net.IPv4len:])
}
