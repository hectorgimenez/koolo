package memory

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/skill"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/game/state"
	"golang.org/x/sys/windows"
)

var dllSeed = windows.MustLoadDLL("rustdecrypt.dll")

func (gd *GameReader) getPlayerUnitPtr() uintptr {
	for i := 1; i < 128; i++ {
		unitOffset := gd.offset.UnitTable + uintptr(i*8)
		playerUnitAddr := gd.process.moduleBaseAddressPtr + unitOffset
		playerUnit := gd.process.ReadUInt(playerUnitAddr, IntTypeUInt64)
		if playerUnit > 0 {
			pInventory := uintptr(playerUnit) + 0x90
			inventoryAddr := uintptr(gd.process.ReadUInt(pInventory, IntTypeUInt64))

			pPath := uintptr(playerUnit) + 0x38
			pathAddress := uintptr(gd.process.ReadUInt(pPath, IntTypeUInt64))
			xPos := gd.process.ReadUInt(pathAddress+0x02, IntTypeUInt16)
			yPos := gd.process.ReadUInt(pathAddress+0x06, IntTypeUInt16)

			// Only current player has inventory
			if inventoryAddr > 0 && xPos > 0 && yPos > 0 {
				return uintptr(playerUnit)
			}
		}
	}

	return 0
}

func (gd *GameReader) GetPlayerUnit(playerUnit uintptr) game.PlayerUnit {
	// Read X and Y Positions
	pPath := playerUnit + 0x38
	pathAddress := uintptr(gd.process.ReadUInt(pPath, IntTypeUInt64))
	xPos := gd.process.ReadUInt(pathAddress+0x02, IntTypeUInt16)
	yPos := gd.process.ReadUInt(pathAddress+0x06, IntTypeUInt16)

	// Player name
	pUnitData := playerUnit + 0x10
	playerNameAddr := uintptr(gd.process.ReadUInt(pUnitData, IntTypeUInt64))
	name := gd.process.ReadStringFromMemory(playerNameAddr, 0)

	// Get Stats
	statsListExPtr := uintptr(gd.process.ReadUInt(playerUnit+0x88, IntTypeUInt64))
	statPtr := gd.process.ReadUInt(statsListExPtr+0x30, IntTypeUInt64)
	statCount := gd.process.ReadUInt(statsListExPtr+0x38, IntTypeUInt64)

	stats := map[stat.Stat]int{}
	for j := 0; j < int(statCount); j++ {
		statOffset := uintptr(statPtr) + 0x2 + uintptr(j*8)
		statNumber := gd.process.ReadUInt(statOffset, IntTypeUInt16)
		statValue := gd.process.ReadUInt(statOffset+0x02, IntTypeUInt32)

		switch stat.Stat(statNumber) {
		case stat.Life,
			stat.MaxLife,
			stat.Mana,
			stat.MaxMana:
			stats[stat.Stat(statNumber)] = int(uint32(statValue) >> 8)
		default:
			stats[stat.Stat(statNumber)] = int(statValue)
		}
	}

	// States (Buff, Debuff, Auras)
	states := gd.getStates(statsListExPtr)

	// Skills
	skills := gd.getSkills(playerUnit + 0x100)

	// Level number
	pathPtr := uintptr(gd.process.ReadUInt(playerUnit+0x38, IntTypeUInt64))
	room1Ptr := uintptr(gd.process.ReadUInt(pathPtr+0x20, IntTypeUInt64))
	room2Ptr := uintptr(gd.process.ReadUInt(room1Ptr+0x18, IntTypeUInt64))
	levelPtr := uintptr(gd.process.ReadUInt(room2Ptr+0x90, IntTypeUInt64))
	levelNo := gd.process.ReadUInt(levelPtr+0x1F8, IntTypeUInt32)

	return game.PlayerUnit{
		Name: name,
		Area: area.Area(levelNo),
		Position: game.Position{
			X: int(xPos),
			Y: int(yPos),
		},
		Stats:  stats,
		Skills: skills,
		States: states,
	}
}

func (gd *GameReader) getMapSeed(playerUnit uintptr) (uintptr, error) {
	actPtr := uintptr(gd.process.ReadUInt(playerUnit+0x20, IntTypeUInt64))
	actMiscPtr := uintptr(gd.process.ReadUInt(actPtr+0x78, IntTypeUInt64))

	dwInitSeedHash1 := uintptr(gd.process.ReadUInt(actMiscPtr+0x840, IntTypeUInt32))
	dwInitSeedHash2 := uintptr(gd.process.ReadUInt(actMiscPtr+0x844, IntTypeUInt32))
	dwEndSeedHash1 := uintptr(gd.process.ReadUInt(actMiscPtr+0x868, IntTypeUInt32))

	p, err := dllSeed.FindProc("get_seed")
	if err != nil {
		return 0, err
	}

	mapSeed, _, err := p.Call(dwInitSeedHash1, dwInitSeedHash2, dwEndSeedHash1)
	if err != windows.ERROR_SUCCESS {
		return 0, err
	}

	return mapSeed, nil
}

func (gd *GameReader) getSkills(skillsPtr uintptr) map[skill.Skill]int {
	skills := make(map[skill.Skill]int)
	skillListPtr := uintptr(gd.process.ReadUInt(skillsPtr, IntTypeUInt64))

	skillPtr := uintptr(gd.process.ReadUInt(skillListPtr, IntTypeUInt64))

	for skillPtr != 0 {
		skillTxtPtr := uintptr(gd.process.ReadUInt(skillPtr, IntTypeUInt64))
		skillTxt := uintptr(gd.process.ReadUInt(skillTxtPtr, IntTypeUInt16))
		skillLvl := gd.process.ReadUInt(skillTxtPtr+0x34, IntTypeUInt16)

		skills[skill.Skill(skillTxt)] = int(skillLvl)

		skillPtr = uintptr(gd.process.ReadUInt(skillPtr+0x08, IntTypeUInt64))
	}

	return skills
}

func (gd *GameReader) getStates(statsListExPtr uintptr) []state.State {
	var states []state.State
	for i := 0; i < 6; i++ {
		offset := i * 4
		stateByte := gd.process.ReadUInt(statsListExPtr+0xAD0+uintptr(offset), IntTypeUInt32)

		offset = (32 * i) - 1
		states = append(states, calculateStates(stateByte, uint(offset))...)
	}

	return states
}

func calculateStates(stateFlag uint, offset uint) []state.State {
	var states []state.State
	if 0x00000001&stateFlag != 0 {
		states = append(states, state.State(1+offset))
	}
	if 0x00000002&stateFlag != 0 {
		states = append(states, state.State(2+offset))
	}
	if 0x00000004&stateFlag != 0 {
		states = append(states, state.State(3+offset))
	}
	if 0x00000008&stateFlag != 0 {
		states = append(states, state.State(4+offset))
	}
	if 0x00000010&stateFlag != 0 {
		states = append(states, state.State(5+offset))
	}
	if 0x00000020&stateFlag != 0 {
		states = append(states, state.State(6+offset))
	}
	if 0x00000040&stateFlag != 0 {
		states = append(states, state.State(7+offset))
	}
	if 0x00000080&stateFlag != 0 {
		states = append(states, state.State(8+offset))
	}
	if 0x00000100&stateFlag != 0 {
		states = append(states, state.State(9+offset))
	}
	if 0x00000200&stateFlag != 0 {
		states = append(states, state.State(10+offset))
	}
	if 0x00000400&stateFlag != 0 {
		states = append(states, state.State(11+offset))
	}
	if 0x00000800&stateFlag != 0 {
		states = append(states, state.State(12+offset))
	}
	if 0x00001000&stateFlag != 0 {
		states = append(states, state.State(13+offset))
	}
	if 0x00002000&stateFlag != 0 {
		states = append(states, state.State(14+offset))
	}
	if 0x00004000&stateFlag != 0 {
		states = append(states, state.State(15+offset))
	}
	if 0x00008000&stateFlag != 0 {
		states = append(states, state.State(16+offset))
	}
	if 0x00010000&stateFlag != 0 {
		states = append(states, state.State(17+offset))
	}
	if 0x00020000&stateFlag != 0 {
		states = append(states, state.State(18+offset))
	}
	if 0x00040000&stateFlag != 0 {
		states = append(states, state.State(19+offset))
	}
	if 0x00080000&stateFlag != 0 {
		states = append(states, state.State(20+offset))
	}
	if 0x00100000&stateFlag != 0 {
		states = append(states, state.State(21+offset))
	}
	if 0x00200000&stateFlag != 0 {
		states = append(states, state.State(22+offset))
	}
	if 0x00400000&stateFlag != 0 {
		states = append(states, state.State(23+offset))
	}
	if 0x00800000&stateFlag != 0 {
		states = append(states, state.State(24+offset))
	}
	if 0x01000000&stateFlag != 0 {
		states = append(states, state.State(25+offset))
	}
	if 0x02000000&stateFlag != 0 {
		states = append(states, state.State(26+offset))
	}
	if 0x04000000&stateFlag != 0 {
		states = append(states, state.State(27+offset))
	}
	if 0x08000000&stateFlag != 0 {
		states = append(states, state.State(28+offset))
	}
	if 0x10000000&stateFlag != 0 {
		states = append(states, state.State(29+offset))
	}
	if 0x20000000&stateFlag != 0 {
		states = append(states, state.State(30+offset))
	}
	if 0x40000000&stateFlag != 0 {
		states = append(states, state.State(31+offset))
	}
	if 0x80000000&stateFlag != 0 {
		states = append(states, state.State(32+offset))
	}

	return states
}
