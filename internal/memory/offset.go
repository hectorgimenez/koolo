package memory

import "encoding/binary"

type Offset struct {
	GameData  uintptr
	UnitTable uintptr
	UI        uintptr
	Hover     uintptr
}

func CalculateOffsets(process Process) Offset {
	// Get all the process memory
	memory := process.getProcessMemory()

	// GameReader
	pattern := process.FindPattern(memory, "\x44\x88\x25\x00\x00\x00\x00\x66\x44\x89\x25\x00\x00\x00\x00", "xxx????xxxx????")
	bytes := process.ReadBytesFromMemory(pattern+0x3, 4)
	offsetInt := uintptr(binary.LittleEndian.Uint32(bytes))
	gameDataOffset := (pattern - process.moduleBaseAddressPtr) - 0x121 + offsetInt

	// UnitTable
	pattern = process.FindPattern(memory, "\x48\x03\xC7\x49\x8B\x8C\xC6", "xxxxxxx")
	bytes = process.ReadBytesFromMemory(pattern+7, 4)
	unitTableOffset := uintptr(binary.LittleEndian.Uint32(bytes))

	// UI
	pattern = process.FindPattern(memory, "\x40\x84\xed\x0f\x94\x05", "xxxxxx")
	uiOffset := process.ReadUInt(pattern+6, IntTypeUInt32)
	uiOffsetPtr := (pattern - process.moduleBaseAddressPtr) + 10 + uintptr(uiOffset)

	// Hover
	pattern = process.FindPattern(memory, "\xc6\x84\xc2\x00\x00\x00\x00\x00\x48\x8b\x74\x24\x00", "xxx?????xxxx?")
	hoverOffset := process.ReadUInt(pattern+3, IntTypeUInt32) - 1

	return Offset{
		GameData:  gameDataOffset,
		UnitTable: unitTableOffset,
		UI:        uiOffsetPtr,
		Hover:     uintptr(hoverOffset),
	}
}
