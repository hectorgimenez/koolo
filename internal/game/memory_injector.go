package game

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"syscall"

	"github.com/hectorgimenez/d2go/pkg/memory"
	"golang.org/x/sys/windows"
)

const fullAccess = windows.PROCESS_VM_OPERATION | windows.PROCESS_VM_WRITE | windows.PROCESS_VM_READ

type MemoryInjector struct {
	isLoaded              bool
	pid                   uint32
	handle                windows.Handle
	getCursorPosAddr      uintptr
	getCursorPosOrigBytes [32]byte
	trackMouseEventAddr   uintptr
	trackMouseEventBytes  [32]byte
	getKeyStateAddr       uintptr
	getKeyStateOrigBytes  [18]byte
	setCursorPosAddr      uintptr
	logger                *slog.Logger
}

func InjectorInit(logger *slog.Logger, pid uint32) (*MemoryInjector, error) {
	i := &MemoryInjector{pid: pid, logger: logger}
	pHandle, err := windows.OpenProcess(fullAccess, false, pid)
	if err != nil {
		return nil, fmt.Errorf("error opening process: %w", err)
	}
	i.handle = pHandle

	return i, nil
}

func (i *MemoryInjector) Load() error {
	if i.isLoaded {
		return nil
	}

	modules, err := memory.GetProcessModules(i.pid)
	if err != nil {
		return fmt.Errorf("error getting process modules: %w", err)
	}

	syscall.MustLoadDLL("USER32.dll")

	for _, module := range modules {
		// GetCursorPos
		if strings.Contains(strings.ToLower(module.ModuleName), "user32.dll") {
			i.getCursorPosAddr, err = syscall.GetProcAddress(module.ModuleHandle, "GetCursorPos")
			i.getKeyStateAddr, _ = syscall.GetProcAddress(module.ModuleHandle, "GetKeyState")
			i.trackMouseEventAddr, _ = syscall.GetProcAddress(module.ModuleHandle, "TrackMouseEvent")
			i.setCursorPosAddr, _ = syscall.GetProcAddress(module.ModuleHandle, "SetCursorPos")

			err = windows.ReadProcessMemory(i.handle, i.getCursorPosAddr, &i.getCursorPosOrigBytes[0], uintptr(len(i.getCursorPosOrigBytes)), nil)
			if err != nil {
				return fmt.Errorf("error reading memory: %w", err)
			}

			err = i.stopTrackingMouseLeaveEvents()
			if err != nil {
				return err
			}

			err = i.OverrideSetCursorPos()
			if err != nil {
				return err
			}

			err = windows.ReadProcessMemory(i.handle, i.getKeyStateAddr, &i.getKeyStateOrigBytes[0], uintptr(len(i.getKeyStateOrigBytes)), nil)
			if err != nil {
				return fmt.Errorf("error reading memory: %w", err)
			}
		}
	}
	if i.getCursorPosAddr == 0 || i.getKeyStateAddr == 0 {
		return errors.New("could not find GetCursorPos address")
	}

	i.isLoaded = true
	return nil
}

func (i *MemoryInjector) Unload() error {
	if err := i.RestoreMemory(); err != nil {
		i.logger.Error(fmt.Sprintf("error restoring memory: %v", err))
	}

	return windows.CloseHandle(i.handle)
}

func (i *MemoryInjector) RestoreMemory() error {
	if !i.isLoaded {
		return nil
	}

	i.isLoaded = false
	err := i.RestoreGetCursorPosAddr()
	if err != nil {
		return fmt.Errorf("error restoring memory: %v", err)
	}

	return i.RestoreGetKeyState()
}

func (i *MemoryInjector) CursorPos(x, y int) error {
	if !i.isLoaded {
		return nil
	}

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

	return windows.WriteProcessMemory(i.handle, i.getCursorPosAddr, &bytes[0], uintptr(len(bytes)), nil)
}

func (i *MemoryInjector) OverrideGetKeyState(key byte) error {
	if !i.isLoaded {
		return nil
	}
	/*
		Assembly: Compare key byte, set al if match, shift left to create 0x8000 if matched (4 cycles)
		cmp cl, key    -> Compare key byte
		sete al        -> Set al to 1 if equal
		shl ax, 15     -> Shift left by 15 to create 0x8000 if was 1, else 0
		ret            -> Return with result in ax
	*/

	bytes := []byte{0x80, 0xF9, key, 0x0F, 0x94, 0xC0, 0x66, 0xC1, 0xE0, 0x0F, 0xC3}

	return windows.WriteProcessMemory(i.handle, i.getKeyStateAddr, &bytes[0], uintptr(len(bytes)), nil)
}
func (i *MemoryInjector) OverrideSetCursorPos() error {
	/*
		Just do nothing, this prevents the game from moving our cursor, for example when opening inventory or wp list
		mov eax, 1
		ret
	*/

	blob := []byte{0xB8, 0x01, 0x00, 0x00, 0x00, 0xC3}
	return windows.WriteProcessMemory(i.handle, i.setCursorPosAddr, &blob[0], uintptr(len(blob)), nil)
}

func (i *MemoryInjector) RestoreGetKeyState() error {
	return windows.WriteProcessMemory(i.handle, i.getKeyStateAddr, &i.getKeyStateOrigBytes[0], uintptr(len(i.getKeyStateOrigBytes)), nil)
}

func (i *MemoryInjector) RestoreGetCursorPosAddr() error {
	return windows.WriteProcessMemory(i.handle, i.getCursorPosAddr, &i.getCursorPosOrigBytes[0], uintptr(len(i.getCursorPosOrigBytes)), nil)
}

// This is needed in order to let the game keep processing mouse events even if the mouse is not over the window
func (i *MemoryInjector) stopTrackingMouseLeaveEvents() error {
	err := windows.ReadProcessMemory(i.handle, i.trackMouseEventAddr, &i.trackMouseEventBytes[0], uintptr(len(i.trackMouseEventBytes)), nil)
	if err != nil {
		return err
	}

	// and dword ptr [rcx+4], 0xFFFFFFFD
	// Modify TRACKMOUSEEVENT struct to disable mouse leave events, since we are injecting our events even if the mouse is not over the window
	disableMouseLeaveRequest := []byte{0x81, 0x61, 0x04, 0xFD, 0xFF, 0xFF, 0xFF}

	// Already hooked
	if bytes.Contains(i.trackMouseEventBytes[:], disableMouseLeaveRequest) {
		return nil
	}

	// We need to move back the pointer 7 bytes to get the correct position, since we are injecting 7 bytes in front of it
	num := int32(binary.LittleEndian.Uint32(i.trackMouseEventBytes[2:6]))
	num -= 7
	numberBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(numberBytes, uint32(num))
	injectBytes := append(i.trackMouseEventBytes[0:2], numberBytes...)

	hook := append(disableMouseLeaveRequest, injectBytes...)

	return windows.WriteProcessMemory(i.handle, i.trackMouseEventAddr, &hook[0], uintptr(len(hook)), nil)
}
