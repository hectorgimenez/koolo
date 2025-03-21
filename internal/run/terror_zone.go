package run

import (
	"fmt"
	"slices"

	"github.com/hectorgimenez/koolo/internal/utils"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/context"
)

type TerrorZone struct {
	ctx *context.Status
}

func NewTerrorZone() *TerrorZone {
	return &TerrorZone{
		ctx: context.Get(),
	}
}

func (tz TerrorZone) Name() string {
	tzNames := make([]string, 0)
	for _, tzArea := range tz.AvailableTZs() {
		tzNames = append(tzNames, tzArea.Area().Name)
	}

	return fmt.Sprintf("TerrorZone Run: %v", tzNames)
}

func (tz TerrorZone) Run() error {

	availableTzs := tz.AvailableTZs()
	if len(availableTzs) == 0 {
		return nil
	}

	switch availableTzs[0] {
	case area.PitLevel1:
		return NewPit().Run()
	case area.Tristram:
		return NewTristram().Run()
	case area.MooMooFarm:
		return NewCows().Run()
	case area.TalRashasTomb1:
		return NewTalRashaTombs().Run()
	case area.AncientTunnels:
		return NewAncientTunnels().Run()
	case area.RockyWaste:
		return NewStonyTomb().Run()
	case area.Travincal:
		return NewTravincal().Run()
	case area.DuranceOfHateLevel1:
		return NewMephisto(tz.customTZEnemyFilter()).Run()
	case area.ChaosSanctuary:
		return NewDiablo().Run()
	case area.NihlathaksTemple:
		return NewNihlathak().Run()
	case area.TheWorldStoneKeepLevel1:
		return NewBaal(tz.customTZEnemyFilter()).Run()
	}

	tzAreaGroups := tz.tzAreaGroups(tz.ctx.Data.TerrorZones[0])
	if len(tzAreaGroups) == 0 {
		return nil
	}

	for _, tzAreaGroup := range tzAreaGroups {
		matchingTzInGroup := false
		for _, tzArea := range tzAreaGroup {
			if slices.Contains(availableTzs, tzArea) {
				matchingTzInGroup = true
				break
			}
		}

		if !matchingTzInGroup {
			continue
		}

		for k, tzArea := range tzAreaGroup {
			if k == 0 {
				err := action.WayPoint(tzArea)
				if err != nil {
					return err
				}
				_ = action.OpenTPIfLeader()
			} else {
				err := action.MoveToArea(tzArea)
				if err != nil {
					return err
				}
				_ = action.OpenTPIfLeader()
			}
			if slices.Contains(availableTzs, tzArea) {
				if tz.ctx.CharacterCfg.Companion.Leader {
					action.OpenTPIfLeader()
					utils.Sleep(5000)
					action.Buff()
				}
				action.ClearCurrentLevel(tz.ctx.CharacterCfg.Game.TerrorZone.OpenChests, tz.customTZEnemyFilter())
			} else {
				tz.ctx.Logger.Debug("Skipping area %v", tzArea.Area().Name)
			}
		}
	}

	return nil
}

func (tz TerrorZone) AvailableTZs() []area.ID {
	tz.ctx.RefreshGameData()
	var availableTZs []area.ID
	for _, tzone := range tz.ctx.Data.TerrorZones {
		for _, tzArea := range tz.ctx.CharacterCfg.Game.TerrorZone.Areas {
			if tzone == tzArea {
				availableTZs = append(availableTZs, tzone)
			}
		}
	}

	return availableTZs
}

func (tz TerrorZone) tzAreaGroups(firstTZ area.ID) [][]area.ID {
	switch firstTZ {
	// Act 1
	case area.BloodMoor:
		return [][]area.ID{{area.RogueEncampment, area.BloodMoor, area.DenOfEvil}}
	case area.ColdPlains:
		return [][]area.ID{{area.ColdPlains, area.CaveLevel1, area.CaveLevel2}}
	case area.BurialGrounds:
		return [][]area.ID{{area.ColdPlains, area.BurialGrounds, area.Crypt}, {area.ColdPlains, area.BurialGrounds, area.Mausoleum}}
	case area.StonyField:
		return [][]area.ID{{area.StonyField}}
	case area.DarkWood:
		return [][]area.ID{{area.DarkWood, area.UndergroundPassageLevel1, area.UndergroundPassageLevel2}}
	case area.BlackMarsh:
		return [][]area.ID{{area.BlackMarsh, area.HoleLevel1, area.HoleLevel2}}
	case area.ForgottenTower:
		return [][]area.ID{{area.BlackMarsh, area.ForgottenTower, area.TowerCellarLevel1, area.TowerCellarLevel2, area.TowerCellarLevel3, area.TowerCellarLevel4, area.TowerCellarLevel5}}
	case area.Barracks:
		return [][]area.ID{{area.JailLevel1, area.Barracks}, {area.JailLevel1, area.JailLevel2, area.JailLevel3}}
	case area.Cathedral:
		return [][]area.ID{{area.InnerCloister, area.Cathedral, area.CatacombsLevel1, area.CatacombsLevel2, area.CatacombsLevel3, area.CatacombsLevel4}}
	// Act 2
	case area.SewersLevel1Act2:
		return [][]area.ID{{area.LutGholein, area.SewersLevel1Act2, area.SewersLevel2Act2, area.SewersLevel3Act2}}
	case area.DryHills:
		return [][]area.ID{{area.DryHills, area.HallsOfTheDeadLevel1, area.HallsOfTheDeadLevel2, area.HallsOfTheDeadLevel3}}
	case area.FarOasis:
		return [][]area.ID{{area.FarOasis}}
	case area.LostCity:
		return [][]area.ID{{area.LostCity, area.ValleyOfSnakes, area.ClawViperTempleLevel1, area.ClawViperTempleLevel2}}
	case area.ArcaneSanctuary:
		return [][]area.ID{{area.ArcaneSanctuary}}
	// Act 3
	case area.SpiderForest:
		return [][]area.ID{{area.SpiderForest, area.SpiderCavern}}
	case area.GreatMarsh:
		return [][]area.ID{{area.GreatMarsh}}
	case area.FlayerJungle:
		return [][]area.ID{{area.FlayerJungle, area.FlayerDungeonLevel1, area.FlayerDungeonLevel2, area.FlayerDungeonLevel3}}
	case area.KurastBazaar:
		return [][]area.ID{{area.KurastBazaar, area.RuinedTemple, area.DisusedFane}}
	// Act 4
	case area.OuterSteppes:
		return [][]area.ID{{area.ThePandemoniumFortress, area.OuterSteppes, area.PlainsOfDespair}}
	case area.RiverOfFlame:
		return [][]area.ID{{area.CityOfTheDamned, area.RiverOfFlame}, {area.RiverOfFlame}}
	// Act 5
	case area.BloodyFoothills:
		return [][]area.ID{{area.Harrogath, area.BloodyFoothills, area.FrigidHighlands, area.Abaddon}}
	case area.GlacialTrail:
		return [][]area.ID{{area.GlacialTrail, area.DrifterCavern}}
	case area.CrystallinePassage:
		return [][]area.ID{{area.CrystallinePassage, area.FrozenRiver}}
	case area.ArreatPlateau:
		return [][]area.ID{{area.ArreatPlateau, area.PitOfAcheron}}
	case area.TheAncientsWay:
		return [][]area.ID{{area.TheAncientsWay, area.IcyCellar}}
	}

	return [][]area.ID{}
}

func (tz TerrorZone) customTZEnemyFilter() data.MonsterFilter {

	return func(m data.Monsters) []data.Monster {
		var filteredMonsters []data.Monster
		monsterFilter := data.MonsterAnyFilter()
		if tz.ctx.CharacterCfg.Game.TerrorZone.FocusOnElitePacks {
			monsterFilter = data.MonsterEliteFilter()
		}

		for _, mo := range m.Enemies(monsterFilter) {
			isImmune := false
			for _, resist := range tz.ctx.CharacterCfg.Game.TerrorZone.SkipOnImmunities {
				if mo.IsImmune(resist) {
					isImmune = true
				}
			}
			if !isImmune {
				filteredMonsters = append(filteredMonsters, mo)
			}
		}

		return filteredMonsters
	}
}
