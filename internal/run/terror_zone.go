package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"go.uber.org/zap"
)

type TerrorZone struct {
	baseRun
	currentTZ string
}

func (a TerrorZone) Name() string {
	return a.currentTZ
}

func (a TerrorZone) BuildActions() (actions []action.Action) {
	act := action.NewChain(func(d data.Data) (actions []action.Action) {
		if len(d.TerrorZones) == 0 {
			a.logger.Info("No TerrorZones detected, skipping TerrorZone run")
			return
		}

		// Try to match terror zones with an existing predefined run
		for _, tz := range d.TerrorZones {
			switch tz {
			case area.PitLevel1:
				return Pit{baseRun: a.baseRun}.BuildActions()
			case area.Tristram:
				return Tristram{baseRun: a.baseRun}.BuildActions()
			case area.MooMooFarm:
				return Cows{baseRun: a.baseRun}.BuildActions()
			case area.TalRashasTomb1:
				return TalRashaTombs{baseRun: a.baseRun}.BuildActions()
			case area.AncientTunnels:
				return AncientTunnels{baseRun: a.baseRun}.BuildActions()
			case area.RockyWaste:
				return StonyTomb{baseRun: a.baseRun}.BuildActions()
			case area.Travincal:
				return Council{baseRun: a.baseRun}.BuildActions()
			case area.DuranceOfHateLevel1:
				return Mephisto{baseRun: a.baseRun}.BuildActions()
			case area.ChaosSanctuary:
				return Diablo{baseRun: a.baseRun}.BuildActions()
			case area.NihlathaksTemple:
				return Nihlathak{baseRun: a.baseRun}.BuildActions()
			case area.TheWorldStoneKeepLevel1:
				return Baal{baseRun: a.baseRun}.BuildActions()
			}
		}

		// If no predefined run is found, we build a custom one
		areasGroups := a.tzAreaChain(d.TerrorZones[0])
		for _, areaGroup := range areasGroups {
			for _, ar := range areaGroup {
				actions = append(actions, a.buildTZAction(ar))
			}
		}

		return
	})

	return []action.Action{act}
}

func (a TerrorZone) AvailableTZs(d data.Data) []area.Area {
	var availableTZs []area.Area
	for _, tz := range d.TerrorZones {
		for _, tzArea := range config.Config.Game.TerrorZone.Areas {
			if tz == tzArea {
				availableTZs = append(availableTZs, tz)
			}
		}
	}

	return availableTZs
}

func (a TerrorZone) buildTZAction(dstArea area.Area) action.Action {
	return action.NewChain(func(d data.Data) (actions []action.Action) {
		if d.PlayerUnit.Area != dstArea && d.PlayerUnit.Area.IsTown() {
			actions = append(actions, a.builder.WayPoint(dstArea))
		}

		actions = append(actions,
			a.builder.MoveToArea(dstArea),
		)

		// Clear only TZ areas, skip traversing areas
		clearArea := false
		for _, terrorizedArea := range d.TerrorZones {
			if terrorizedArea == dstArea {
				// Skip areas that are not selected in the configuration
				for _, tz := range config.Config.Game.TerrorZone.Areas {
					if tz == dstArea {
						clearArea = true
					}
				}
			}
		}

		if clearArea {
			a.logger.Debug("Clearing TZ area", zap.Any("area", dstArea))
			actions = append(actions, a.builder.ClearArea(true, customTZEnemyFilter(config.Config.Game.TerrorZone.SkipOnImmunities...)))
		} else {
			a.logger.Debug("TZ area skipped", zap.Any("area", dstArea))
		}

		return actions
	})
}

func (a TerrorZone) tzAreaChain(firstTZ area.Area) [][]area.Area {
	switch firstTZ {
	// Act 1
	case area.BloodMoor:
		return [][]area.Area{{area.RogueEncampment, area.BloodMoor, area.DenOfEvil}}
	case area.ColdPlains:
		return [][]area.Area{{area.ColdPlains, area.CaveLevel1, area.CaveLevel2}}
	case area.BurialGrounds:
		return [][]area.Area{{area.ColdPlains, area.BurialGrounds, area.Crypt}, {area.ColdPlains, area.BurialGrounds, area.Mausoleum}}
	case area.StonyField:
		return [][]area.Area{{area.StonyField}}
	case area.DarkWood:
		return [][]area.Area{{area.DarkWood, area.UndergroundPassageLevel1, area.UndergroundPassageLevel2}}
	case area.BlackMarsh:
		return [][]area.Area{{area.BlackMarsh, area.HoleLevel1, area.HoleLevel2}}
	case area.ForgottenTower:
		return [][]area.Area{{area.BlackMarsh, area.ForgottenTower, area.TowerCellarLevel1, area.TowerCellarLevel2, area.TowerCellarLevel3, area.TowerCellarLevel4, area.TowerCellarLevel5}}
	case area.JailLevel1:
		return [][]area.Area{{area.JailLevel1, area.JailLevel2, area.JailLevel3}}
	case area.Cathedral:
		return [][]area.Area{{area.InnerCloister, area.Cathedral, area.CatacombsLevel1, area.CatacombsLevel2, area.CatacombsLevel3}}
	// Act 2
	case area.SewersLevel1Act2:
		return [][]area.Area{{area.LutGholein, area.SewersLevel1Act2, area.SewersLevel2Act2, area.SewersLevel3Act2}}
	case area.DryHills:
		return [][]area.Area{{area.DryHills, area.HallsOfTheDeadLevel1, area.HallsOfTheDeadLevel2, area.HallsOfTheDeadLevel3}}
	case area.FarOasis:
		return [][]area.Area{{area.FarOasis}}
	case area.LostCity:
		return [][]area.Area{{area.LostCity, area.ValleyOfSnakes, area.ClawViperTempleLevel1, area.ClawViperTempleLevel2}}
	case area.ArcaneSanctuary:
		return [][]area.Area{{area.ArcaneSanctuary}}
	// Act 3
	case area.SpiderForest:
		return [][]area.Area{{area.SpiderForest, area.SpiderCavern}}
	case area.GreatMarsh:
		return [][]area.Area{{area.GreatMarsh}}
	case area.FlayerJungle:
		return [][]area.Area{{area.FlayerJungle, area.FlayerDungeonLevel1, area.FlayerDungeonLevel2, area.FlayerDungeonLevel3}}
	case area.KurastBazaar:
		return [][]area.Area{{area.KurastBazaar, area.RuinedTemple, area.DisusedFane}}
	// Act 4
	case area.OuterSteppes:
		return [][]area.Area{{area.ThePandemoniumFortress, area.OuterSteppes, area.PlainsOfDespair}}
	case area.RiverOfFlame:
		return [][]area.Area{{area.CityOfTheDamned, area.RiverOfFlame}}
	// Act 5
	case area.BloodyFoothills:
		return [][]area.Area{{area.Harrogath, area.BloodyFoothills, area.FrigidHighlands, area.Abaddon}}
	case area.GlacialTrail:
		return [][]area.Area{{area.GlacialTrail, area.DrifterCavern}}
	case area.CrystallinePassage:
		return [][]area.Area{{area.CrystallinePassage, area.FrozenRiver}}
	case area.ArreatPlateau:
		return [][]area.Area{{area.ArreatPlateau, area.PitOfAcheron}}
	case area.TheAncientsWay:
		return [][]area.Area{{area.TheAncientsWay, area.IcyCellar}}
	}

	return [][]area.Area{}
}

func customTZEnemyFilter(resists ...stat.Resist) data.MonsterFilter {
	return func(m data.Monsters) []data.Monster {
		var filteredMonsters []data.Monster
		monsterFilter := data.MonsterAnyFilter()
		if config.Config.Game.TerrorZone.FocusOnElitePacks {
			monsterFilter = data.MonsterEliteFilter()
		}

		for _, mo := range m.Enemies(monsterFilter) {
			isImmune := false
			for _, resist := range resists {
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
