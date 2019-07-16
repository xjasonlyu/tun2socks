// +build windows

package windows

import (
	"syscall"
	"unsafe"
)

var (
	iphlpapi = syscall.NewLazyDLL("iphlpapi.dll")

	procGetTcpStatistics     = iphlpapi.NewProc("GetTcpStatistics")
	procGetExtendedTcpTable  = iphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable  = iphlpapi.NewProc("GetExtendedUdpTable")
	procGetBestRoute         = iphlpapi.NewProc("GetBestRoute")
	procGetIpForwardTable    = iphlpapi.NewProc("GetIpForwardTable")
	procGetInterfaceInfo     = iphlpapi.NewProc("GetInterfaceInfo")
	procGetIfTable           = iphlpapi.NewProc("GetIfTable")
	procDeleteIpForwardEntry = iphlpapi.NewProc("DeleteIpForwardEntry")
	procCreateIpForwardEntry = iphlpapi.NewProc("CreateIpForwardEntry")
)

func GetTcpStatistics(statistics *MIB_TCPSTATS) int {
	ret, _, _ := procGetTcpStatistics.Call(
		uintptr(unsafe.Pointer(statistics)),
	)
	return int(ret)
}

func GetExtendedTcpTable(tcpTable uintptr, size *uint32, order int32, af uint32, tableClass TCP_TABLE_CLASS) int {
	ret, _, _ := procGetExtendedTcpTable.Call(
		tcpTable,
		uintptr(unsafe.Pointer(size)),
		uintptr(order),
		uintptr(af),
		uintptr(tableClass),
		0,
	)
	return int(ret)
}

func GetExtendedUdpTable(udpTable uintptr, size *uint32, order int32, af uint32, tableClass UDP_TABLE_CLASS) int {
	ret, _, _ := procGetExtendedUdpTable.Call(
		udpTable,
		uintptr(unsafe.Pointer(size)),
		uintptr(order),
		uintptr(af),
		uintptr(tableClass),
		0,
	)
	return int(ret)
}

func GetBestRoute(destAddr, sourceAddr uint32, bestRoute *MIB_IPFORWARDROW) int {
	ret, _, _ := procGetBestRoute.Call(
		uintptr(destAddr),
		uintptr(sourceAddr),
		uintptr(unsafe.Pointer(bestRoute)),
	)
	return int(ret)
}

func GetIpForwardTable(table *MIB_IPFORWARDTABLE, size *uint32, order int32) int {
	ret, _, _ := procGetIpForwardTable.Call(
		uintptr(unsafe.Pointer(table)),
		uintptr(unsafe.Pointer(size)),
		uintptr(order),
	)
	return int(ret)
}

func GetInterfaceInfo(ifTable *IP_INTERFACE_INFO, outBufLen *uint32) int {
	ret, _, _ := procGetInterfaceInfo.Call(
		uintptr(unsafe.Pointer(ifTable)),
		uintptr(unsafe.Pointer(outBufLen)),
	)
	return int(ret)
}

func GetIfTable(table *MIB_IFTABLE, size *uint32, order int32) int {
	ret, _, _ := procGetIfTable.Call(
		uintptr(unsafe.Pointer(table)),
		uintptr(unsafe.Pointer(size)),
		uintptr(order),
	)
	return int(ret)
}

func DeleteIpForwardEntry(route *MIB_IPFORWARDROW) uint32 {
	ret, _, _ := procDeleteIpForwardEntry.Call(
		uintptr(unsafe.Pointer(route)),
	)
	return uint32(ret)
}

func CreateIpForwardEntry(route *MIB_IPFORWARDROW) uint32 {
	ret, _, _ := procCreateIpForwardEntry.Call(
		uintptr(unsafe.Pointer(route)),
	)
	return uint32(ret)
}
