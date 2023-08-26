package tcp

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

func CloseCurrentGameSocket(gamePID uint) error {
	var lastOpenSocket []byte
	var buffSize DWORD
	_ = GetExtendedTCPTable(uintptr(0), &buffSize, true, 2, TCP_TABLE_OWNER_PID_ALL, 0)
	var buffTable = make([]byte, int(buffSize))
	err := GetExtendedTCPTable(uintptr(unsafe.Pointer(&buffTable[0])), &buffSize, true, 2, TCP_TABLE_OWNER_PID_ALL, 0)
	if !errors.Is(err, windows.DNS_ERROR_RCODE_NO_ERROR) {
		return CloseGameSocketErr
	}

	count := *(*uint32)(unsafe.Pointer(&buffTable[0]))
	const structLen = 24
	for n, pos := uint32(0), 4; n < count && pos+structLen <= len(buffTable); n, pos = n+1, pos+structLen {
		state := *(*uint32)(unsafe.Pointer(&buffTable[pos]))
		if state < 1 || state > 12 {
			return CloseGameSocketErr
		}
		pid := *(*uint32)(unsafe.Pointer(&buffTable[pos+20]))

		if uint(pid) == gamePID && state == uint32(MIB_TCP_STATE_ESTAB) {
			buffTable[pos] = byte(MIB_TCP_STATE_DELETE_TCB)
			lastOpenSocket = buffTable[pos : pos+24]
		}
	}

	if len(lastOpenSocket) == 0 {
		return CloseGameSocketErr
	}

	err = SetTCPEntry(uintptr(unsafe.Pointer(&lastOpenSocket[0])))
	if errors.Is(err, windows.DNS_ERROR_RCODE_NO_ERROR) {
		return nil
	}

	return CloseGameSocketErr
}
