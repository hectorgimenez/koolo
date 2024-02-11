package memory

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/lxn/win"
	"github.com/winlabs/gowin32"
	"golang.org/x/sys/windows"
	"strings"
	"syscall"
)

const fullAccess = windows.PROCESS_VM_OPERATION | windows.PROCESS_VM_WRITE | windows.PROCESS_VM_READ

var handle windows.Handle
var HWND win.HWND
var getCursorPosAddr uintptr
var getCursorPosOrigBytes [32]byte
var getKeyStateAddr uintptr
var getKeyStateOrigBytes [18]byte

func ASMInjectorInit(pid uint32) error {
	pHandle, err := windows.OpenProcess(fullAccess, false, pid)
	if err != nil {
		return fmt.Errorf("error opening process: %w", err)
	}
	handle = pHandle

	modules, err := gowin32.GetProcessModules(pid)
	if err != nil {
		return fmt.Errorf("error getting process modules: %w", err)
	}

	_, err = syscall.LoadLibrary("USER32.dll")
	if err != nil {
		return fmt.Errorf("error loading USER32.dll: %w", err)
	}

	for _, module := range modules {
		// GetCursorPos
		if strings.EqualFold(module.ModuleName, "USER32.dll") {
			getCursorPosAddr, err = syscall.GetProcAddress(module.ModuleHandle, "GetCursorPos")
			getKeyStateAddr, _ = syscall.GetProcAddress(module.ModuleHandle, "GetKeyState")

			err = windows.ReadProcessMemory(handle, getCursorPosAddr, &getCursorPosOrigBytes[0], uintptr(len(getCursorPosOrigBytes)), nil)
			if err != nil {
				return fmt.Errorf("error reading memory: %w", err)
			}

			err = windows.ReadProcessMemory(handle, getKeyStateAddr, &getKeyStateOrigBytes[0], uintptr(len(getKeyStateOrigBytes)), nil)
			if err != nil {
				return fmt.Errorf("error reading memory: %w", err)
			}
		}
	}
	if getCursorPosAddr == 0 || getKeyStateAddr == 0 {
		return errors.New("could not find GetCursorPos address")
	}

	return nil
}

func ASMInjectorUnload() error {
	err := RestoreGetCursorPosAddr()
	if err != nil {
		return fmt.Errorf("error writing to memory: %w", err)
	}

	err = RestoreGetKeyState()
	if err != nil {
		return err
	}

	return nil
}

func InjectCursorPos(x, y int) error {
	/*
		push rax
		mov rax, rcx
		mov dword ptr [rax], 1 // X
		mov dword ptr [rax+4], 2 // Y
		pop rax
		mov al, 1
		ret
	*/
	bytes := []byte{0x50, 0x48, 0x89, 0xC8, 0xC7, 0x00, 0x01, 0x00, 0x00, 0x00, 0xC7, 0x40, 0x04, 0x02, 0x00, 0x00, 0x00, 0x58, 0xB0, 0x01, 0xC3}

	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, uint32(x))
	copy(bytes[6:], buff)

	binary.LittleEndian.PutUint32(buff, uint32(y))
	copy(bytes[13:], buff)

	return windows.WriteProcessMemory(handle, getCursorPosAddr, &bytes[0], uintptr(len(bytes)), nil)
}

func OverrideGetKeyState(key int) error {
	/*
		cmp rcx, 0x12
		mov rax, 0x8000
		ret
	*/
	bytes := []byte{0x48, 0x81, 0xF9, byte(key), 0x00, 0x00, 0x00, 0x48, 0xB8, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC3}
	return windows.WriteProcessMemory(handle, getKeyStateAddr, &bytes[0], uintptr(len(bytes)), nil)
}

func RestoreGetKeyState() error {
	return windows.WriteProcessMemory(handle, getKeyStateAddr, &getKeyStateOrigBytes[0], uintptr(len(getKeyStateOrigBytes)), nil)
}

func RestoreGetCursorPosAddr() error {
	return windows.WriteProcessMemory(handle, getCursorPosAddr, &getCursorPosOrigBytes[0], uintptr(len(getCursorPosOrigBytes)), nil)
}
