// +build windows

package lsof

import (
	"fmt"
	"syscall"
	"unsafe"

	win "github.com/xjasonlyu/tun2socks/common/lsof/windows"
)

func GetCommandNameBySocket(network string, addr string, port uint16) (string, error) {
	switch network {
	case "tcp":
		tcpTable, err := getTcpTable()
		if err != nil {
			return "", fmt.Errorf("failed to get TCP table: %v", err)
		}
		for i := 0; i < int(tcpTable.NumEntries); i++ {
			row := tcpTable.Table[i]
			if win.NTOHS(uint16(row.LocalPort)) == port /* && win.IPAddrNTOA(uint32(row.LocalAddr)) == addr */ {
				return getNameByPid(uint32(row.OwningPid))
			}
		}
		return "", ErrNotFound
	case "udp":
		var udpTable win.MIB_UDPTABLE_OWNER_PID
		err := getUdpTable(
			uintptr(unsafe.Pointer(&udpTable)),
			win.AF_INET,
		)
		if err != nil {
			return "", fmt.Errorf("failed to get UDP table: %v", err)
		}
		for i := 0; i < int(udpTable.NumEntries); i++ {
			row := udpTable.Table[i]
			if win.NTOHS(uint16(row.LocalPort)) == port /* && win.IPAddrNTOA(uint32(row.LocalAddr)) == addr */ {
				return getNameByPid(uint32(row.OwningPid))
			}
		}

		// var udp6Table win.MIB_UDP6TABLE_OWNER_PID
		// err = getUdpTable(
		// 	uintptr(unsafe.Pointer(&udp6Table)),
		// 	win.AF_INET6,
		// )
		// if err != nil {
		// 	return "", fmt.Errorf("failed to get UDP table: %v", err)
		// }
		// for i := 0; i < int(udp6Table.NumEntries); i++ {
		// 	row := udp6Table.Table[i]
		// 	if win.NTOHS(uint16(row.LocalPort)) == port /* && win.IPAddrNTOA(uint32(row.LocalAddr)) == addr */ {
		// 		return getNameByPid(uint32(row.OwningPid))
		// 	}
		// }

		return "", ErrNotFound
	default:
		return "", ErrNotFound
	}
}

func getNameByPid(pid uint32) (string, error) {
	handle := win.CreateToolhelp32Snapshot(
		win.TH32CS_SNAPMODULE,
		pid,
	)
	if handle <= 0 {
		return "", fmt.Errorf("failed to create snapshot: %v", handle)
	}
	defer win.CloseHandle(handle)

	var me win.MODULEENTRY32
	me.Size = uint32(unsafe.Sizeof(me))
	success := win.Module32First(handle, &me)
	if success {
		return win.UTF16PtrToString(&me.Module[0]), nil
	} else {
		return "", fmt.Errorf("failed to get process entry: %v", syscall.GetLastError())
	}
}

func getTcpTable() (win.MIB_TCPTABLE_OWNER_PID, error) {
	var tcpTable win.MIB_TCPTABLE_OWNER_PID
	var size uint32 = 64 * 1024
	var order int32 = 0
	ret := win.GetExtendedTcpTable(
		uintptr(unsafe.Pointer(&tcpTable)),
		&size,
		order,
		win.AF_INET,
		win.TCP_TABLE_OWNER_PID_ALL,
	)
	if ret != 0 {
		return tcpTable, fmt.Errorf("ret: %d", int(ret))
	}
	return tcpTable, nil
}

func getUdpTable(table uintptr, af uint32) error {
	var size uint32 = 64 * 1024
	var order int32 = 0
	ret := win.GetExtendedUdpTable(
		table,
		&size,
		order,
		af,
		win.UDP_TABLE_OWNER_PID,
	)
	if ret != 0 {
		return fmt.Errorf("ret: %d", int(ret))
	}
	return nil
}
