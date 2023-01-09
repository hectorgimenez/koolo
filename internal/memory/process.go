package memory

import (
	"encoding/binary"
	"github.com/winlabs/gowin32"
	"go.uber.org/zap"
	"golang.org/x/sys/windows"
	"unsafe"
)

const moduleName = "D2R.exe"

type Process struct {
	handler              windows.Handle
	pid                  uint
	moduleBaseAddress    unsafe.Pointer
	moduleBaseAddressPtr uintptr
	moduleBaseSize       uint
	logger               *zap.Logger
}

func NewProcess(logger *zap.Logger) (Process, error) {
	module, err := getGameModule()
	if err != nil {
		return Process{}, err
	}

	h, err := windows.OpenProcess(0x0010, false, uint32(module.ProcessID))
	if err != nil {
		return Process{}, err
	}

	return Process{
		handler:              h,
		pid:                  module.ProcessID,
		moduleBaseAddress:    unsafe.Pointer(module.ModuleBaseAddress),
		moduleBaseAddressPtr: uintptr(unsafe.Pointer(module.ModuleBaseAddress)),
		moduleBaseSize:       module.ModuleBaseSize,
		logger:               logger,
	}, nil
}

func getGameModule() (gowin32.ModuleInfo, error) {
	processes := make([]uint32, 2048)
	length := uint32(0)
	err := windows.EnumProcesses(processes, &length)
	if err != nil {
		return gowin32.ModuleInfo{}, err
	}

	for _, process := range processes {
		module, _ := getMainModule(process)

		if module.ExePath == "C:\\Program Files (x86)\\Diablo II Resurrected\\D2R.exe" {
			return module, nil
		}
	}

	return gowin32.ModuleInfo{}, err
}

func getMainModule(pid uint32) (gowin32.ModuleInfo, error) {
	mi, err := gowin32.GetProcessModules(pid)
	if err != nil {
		return gowin32.ModuleInfo{}, err
	}
	for _, m := range mi {
		if m.ModuleName == moduleName {
			return m, nil
		}
	}

	return gowin32.ModuleInfo{}, err
}

func (p Process) getProcessMemory() []byte {
	var data = make([]byte, p.moduleBaseSize)
	err := windows.ReadProcessMemory(p.handler, uintptr(p.moduleBaseAddress), &data[0], uintptr(p.moduleBaseSize), nil)
	if err != nil {
		p.logger.Debug("Error reading memory position", zap.Error(err))
	}

	return data
}

func (p Process) ReadBytesFromMemory(address uintptr, size uint) []byte {
	var data = make([]byte, size)
	err := windows.ReadProcessMemory(p.handler, address, &data[0], uintptr(size), nil)
	if err != nil {
		p.logger.Debug("Error reading memory position", zap.Error(err))
	}

	return data
}

type IntType uint

const (
	IntTypeUInt8  = 1
	IntTypeUInt16 = 2
	IntTypeUInt32 = 4
	IntTypeUInt64 = 8
)

func (p Process) ReadUInt(address uintptr, size IntType) uint {
	bytes := p.ReadBytesFromMemory(address, uint(size))

	return bytesToUint(bytes, size)
}

func ReadUIntFromBuffer(bytes []byte, offset uint, size IntType) uint {
	return bytesToUint(bytes[offset:offset+uint(size)], size)
}

func bytesToUint(bytes []byte, size IntType) uint {
	switch size {
	case IntTypeUInt8:
		return uint(bytes[0])
	case IntTypeUInt16:
		return uint(binary.LittleEndian.Uint16(bytes))
	case IntTypeUInt32:
		return uint(binary.LittleEndian.Uint32(bytes))
	case IntTypeUInt64:
		return uint(binary.LittleEndian.Uint64(bytes))
	}

	return 0
}

func (p Process) ReadStringFromMemory(address uintptr, size uint) string {
	if size == 0 {
		for i := 1; true; i++ {
			data := p.ReadBytesFromMemory(address, uint(i))
			if data[i-1] == 0 {
				return string(data)
			}
		}
	}

	return string(p.ReadBytesFromMemory(address, size))
}

func (p Process) findPattern(memory []byte, pattern, mask string) int {
	patternLength := len(pattern)
	for i := 0; i < int(p.moduleBaseSize)-patternLength; i++ {
		found := true
		for j := 0; j < patternLength; j++ {
			if string(mask[j]) != "?" && string(pattern[j]) != string(memory[i+j]) {
				found = false
				break
			}
		}

		if found {
			return i
		}
	}

	return 0
}

func (p Process) FindPattern(memory []byte, pattern, mask string) uintptr {
	if offset := p.findPattern(memory, pattern, mask); offset != 0 {
		return p.moduleBaseAddressPtr + uintptr(offset)
	}

	return 0
}

func (p Process) GetPID() uint {
	return p.pid
}
