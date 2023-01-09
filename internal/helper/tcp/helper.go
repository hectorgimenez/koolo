package tcp

import (
	"syscall"
	"unsafe"
)

var (
	lib                 = syscall.MustLoadDLL("iphlpapi.dll")
	setTCPEntry         = lib.MustFindProc("SetTcpEntry").Addr()
	getExtendedTCPTable = lib.MustFindProc("GetExtendedTcpTable").Addr()
)

func SetTCPEntry(ptr uintptr) error {
	_, _, err := syscall.SyscallN(setTCPEntry, ptr)

	return err
}

func GetExtendedTCPTable(tcpTablePtr uintptr, pdwSize *DWORD, bOrder bool, ulAf ULONG, tableClass TCP_TABLE_CLASS, reserved ULONG) error {
	_, _, err := syscall.SyscallN(getExtendedTCPTable,
		tcpTablePtr,
		uintptr(unsafe.Pointer(pdwSize)),
		getUintptrFromBool(bOrder),
		uintptr(ulAf),
		uintptr(tableClass),
		uintptr(reserved))

	return err
}

func getUintptrFromBool(b bool) uintptr {
	if b {
		return 1
	} else {
		return 0
	}
}
