package memory

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/npc"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/pather"
	"sort"
)

func (gd *GameReader) Monsters(playerPositionX, playerPositionY int) game.Monsters {
	hoveredUnitID, hoveredType, isHovered := gd.hoveredData()

	baseAddr := gd.Process.moduleBaseAddressPtr + gd.offset.UnitTable + 1024
	unitTableBuffer := gd.Process.ReadBytesFromMemory(baseAddr, 128*8)

	monsters := game.Monsters{}
	for i := 0; i < 128; i++ {
		monsterOffset := 8 * i
		monsterUnitPtr := uintptr(ReadUIntFromBuffer(unitTableBuffer, uint(monsterOffset), IntTypeUInt64))
		for monsterUnitPtr > 0 {
			monsterDataBuffer := gd.Process.ReadBytesFromMemory(monsterUnitPtr, 144)

			//monsterType := ReadUIntFromBuffer(monsterDataBuffer, 0x00, IntTypeUInt32)
			txtFileNo := ReadUIntFromBuffer(monsterDataBuffer, 0x04, IntTypeUInt32)
			if !gd.shouldBeIgnored(txtFileNo) {
				unitID := ReadUIntFromBuffer(monsterDataBuffer, 0x08, IntTypeUInt32)

				//mode := ReadUIntFromBuffer(monsterDataBuffer, 0x0C, IntTypeUInt32)

				unitDataPtr := uintptr(ReadUIntFromBuffer(monsterDataBuffer, 0x10, IntTypeUInt64))
				//isUnique := gd.Process.ReadUInt(unitDataPtr+0x18, IntTypeUInt16)
				flag := gd.Process.ReadBytesFromMemory(unitDataPtr+0x1A, IntTypeUInt8)[0]
				isCorpse := gd.Process.ReadUInt(monsterUnitPtr+0x1A6, IntTypeUInt8)

				//unitDataBuffer := gd.Process.ReadBytesFromMemory(unitDataPtr, 144)

				// Coordinates (X, Y)
				pathPtr := uintptr(gd.Process.ReadUInt(monsterUnitPtr+0x38, IntTypeUInt64))
				posX := gd.Process.ReadUInt(pathPtr+0x02, IntTypeUInt16)
				posY := gd.Process.ReadUInt(pathPtr+0x06, IntTypeUInt16)

				hovered := false
				if isHovered && hoveredType == 1 && hoveredUnitID == unitID {
					hovered = true
				}

				statsListExPtr := uintptr(ReadUIntFromBuffer(monsterDataBuffer, 0x88, IntTypeUInt64))
				statPtr := uintptr(gd.Process.ReadUInt(statsListExPtr+0x30, IntTypeUInt64))
				statCount := gd.Process.ReadUInt(statsListExPtr+0x38, IntTypeUInt64)

				stats := gd.getMonsterStats(statCount, statPtr)

				m := game.Monster{
					UnitID:    game.UnitID(unitID),
					Name:      npc.ID(int(txtFileNo)),
					IsHovered: hovered,
					Position: game.Position{
						X: int(posX),
						Y: int(posY),
					},
					Stats: stats,
					Type:  getMonsterType(flag),
				}

				//fmt.Println(monsterType, mode, flag, isUnique)
				if isCorpse == 0 {
					monsters = append(monsters, m)
				}
			}

			monsterUnitPtr = uintptr(gd.Process.ReadUInt(monsterUnitPtr+0x150, IntTypeUInt64))
		}
	}

	sort.SliceStable(monsters, func(i, j int) bool {
		distanceI := pather.DistanceFromPoint(playerPositionX, playerPositionY, monsters[i].Position.X, monsters[i].Position.Y)
		distanceJ := pather.DistanceFromPoint(playerPositionX, playerPositionY, monsters[j].Position.X, monsters[j].Position.Y)

		return distanceI < distanceJ
	})

	return monsters
}

func getMonsterType(typeFlag byte) game.MonsterType {
	switch typeFlag {
	case 10:
		return game.MonsterTypeSuperUnique
	case 1 << 2:
		return game.MonsterTypeChampion
	case 1 << 3:
		return game.MonsterTypeUnique
	case 1 << 4:
		return game.MonsterTypeMinion
	}

	return game.MonsterTypeNone
}

func (gd *GameReader) getMonsterStats(statCount uint, statPtr uintptr) map[stat.Stat]int {
	stats := map[stat.Stat]int{}

	if statCount > 0 {
		statBuffer := gd.Process.ReadBytesFromMemory(statPtr+0x2, statCount*8)
		for i := 0; i < int(statCount); i++ {
			offset := uint(i * 8)
			statEnum := ReadUIntFromBuffer(statBuffer, offset, IntTypeUInt16)
			statValue := ReadUIntFromBuffer(statBuffer, offset+0x2, IntTypeUInt32)
			stats[stat.Stat(statEnum)] = int(statValue)
		}
	}

	return stats
}

func (gd *GameReader) shouldBeIgnored(txtNo uint) bool {
	switch txtNo {
	case 149, //Chicken
		151, //Rat
		152, //Rogue
		153, //HellMeteor
		157, //Bird
		158, //Bird2
		159, //Bat
		195, //Act2Male
		196, //Act2Female
		197, //Act2Child
		179, //Cow

		185, //Camel
		203, //Act2Guard
		204, //Act2Vendor
		205, //Act2Vendor2
		227, //Maggot
		268, //Bug
		269, //Scorpion
		271, //Rogue2
		272, //Rogue3
		283, //Larva
		293, //Familiar
		294, //Act3Male
		289, //ClayGolem
		290, //BloodGolem
		291, //IronGolem
		292, //FireGolem
		296, //Act3Female
		318, //Snake
		319, //Parrot
		320, //Fish
		321, //EvilHole
		322, //EvilHole2
		323, //EvilHole3
		324, //EvilHole4
		325, //EvilHole5
		326, //FireboltTrap
		327, //HorzMissileTrap
		328, //VertMissileTrap
		329, //PoisonCloudTrap
		330, //LightningTrap
		332, //InvisoSpawner
		//338, //Guard
		339, //MiniSpider
		344, //BoneWall
		351, //Hydra
		352, //Hydra2
		353, //Hydra3
		355, //SevenTombs
		357, //Valkyrie
		359, //IronWolf
		363, //NecroSkeleton
		364, //NecroMage
		366, //CompellingOrb
		370, //SpiritMummy
		377, //Act2Guard4
		378, //Act2Guard5
		392, //Window
		393, //Window2
		401, //MephistoSpirit
		410, //WakeOfDestruction
		411, //ChargedBoltSentry
		412, //LightningSentry
		414, //InvisiblePet
		415, //InfernoSentry
		416, //DeathSentry
		417, //ShadowWarrior
		418, //ShadowMaster
		419, //DruidHawk
		420, //DruidSpiritWolf
		421, //DruidFenris
		423, //HeartOfWolverine
		424, //OakSage
		428, //DruidBear
		543, //BaalThrone
		567, //InjuredBarbarian
		568, //InjuredBarbarian2
		569, //InjuredBarbarian3
		711: //DemonHole
		return true
	}

	return false
}
