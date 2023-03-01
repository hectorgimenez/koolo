package memory

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
)

func (gd *GameReader) getRoster() (roster []game.RosterMember) {
	partyStruct := uintptr(gd.Process.ReadUInt(gd.Process.moduleBaseAddressPtr+gd.offset.RosterOffset, Uint64))

	for partyStruct > 0 {
		name := gd.Process.ReadStringFromMemory(partyStruct, 16)
		a := gd.Process.ReadUInt(partyStruct+0x5C, Uint32)

		xPos := gd.Process.ReadUInt(partyStruct+0x60, Uint32)
		yPos := gd.Process.ReadUInt(partyStruct+0x64, Uint32)

		roster = append(roster, game.RosterMember{
			Name:     name,
			Area:     area.Area(a),
			Position: game.Position{X: int(xPos), Y: int(yPos)},
		})
		partyStruct = uintptr(gd.Process.ReadUInt(partyStruct+0x148, Uint64))
	}

	return
}
