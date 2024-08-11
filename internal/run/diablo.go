package run

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/pather"
)

var diabloSpawnPosition = data.Position{X: 7792, Y: 5294}
var chaosSanctuaryEntrancePosition = data.Position{X: 7790, Y: 5544}

type Diablo struct {
	baseRun
	bm             health.BeltManager
	vizLayout      int
	seisLayout     int
	infLayout      int
	cleared        []data.Position
	entranceToStar []data.Position
	starToVizA     []data.Position
	starToVizB     []data.Position
	starToSeisA    []data.Position
	starToSeisB    []data.Position
	starToInfA     []data.Position
	starToInfB     []data.Position
}

func (d Diablo) Name() string {
	return string(config.DiabloRun)
}

func (d Diablo) BuildActions() []action.Action {
	d = d.initLayout()
	d = d.initPaths()

	var actions []action.Action

	actions = append(actions,
		d.builder.WayPoint(area.RiverOfFlame),
		d.builder.Buff(),
	)

	if d.CharacterCfg.Game.Diablo.FullClear {
		actions = append(actions, d.builder.MoveToCoords(chaosSanctuaryEntrancePosition))
	} else {
		actions = append(actions, d.builder.MoveToCoords(diabloSpawnPosition))
	}

	if d.CharacterCfg.Game.Diablo.FullClear {
		actions = append(actions, d.entranceToStarClear()...)
	}

	if d.CharacterCfg.Game.Diablo.FullClear {
		actions = append(actions, d.starToVizClear()...)
	}

	actions = append(actions, d.killVizier())

	if d.CharacterCfg.Game.Diablo.FullClear {
		actions = append(actions, d.starToSeisClear()...)
	}

	actions = append(actions, d.killSeis())

	if d.CharacterCfg.Game.Diablo.FullClear {
		actions = append(actions, d.starToInfClear()...)
	}

	actions = append(actions, d.killInfector())

	if d.CharacterCfg.Game.Diablo.KillDiablo {
		actions = append(actions,
			d.builder.Buff(),
			d.builder.MoveToCoords(diabloSpawnPosition),
			d.char.KillDiablo(),
		)
	}

	actions = append(actions, d.builder.ItemPickup(true, 40))

	return actions
}

func (d Diablo) initLayout() Diablo {
	d.vizLayout = d.getLayout(object.DiabloSeal4, 5275)
	d.seisLayout = d.getLayout(object.DiabloSeal3, 7773)
	d.infLayout = d.getLayout(object.DiabloSeal1, 7893)

	d.logger.Debug(fmt.Sprintf("Layouts initialized - Vizier: %d, Seis: %d, Infector: %d", d.vizLayout, d.seisLayout, d.infLayout))
	return d
}

func (d Diablo) getLayout(seal object.Name, value int) int {
	mapData := d.Reader.GetCachedMapData(false)
	origin := mapData.Origin(area.ChaosSanctuary)
	_, _, objects, _ := mapData.NPCsExitsAndObjects(origin, area.ChaosSanctuary)

	for _, obj := range objects {
		if obj.Name == seal {
			if obj.Position.Y == value || obj.Position.X == value {
				d.logger.Debug(fmt.Sprintf("Layout 1 detected for seal %v: position matches value %d", seal, value))
				return 1
			}
			d.logger.Debug(fmt.Sprintf("Layout 2 detected for seal %v: position does not match value %d", seal, value))
			return 2
		}
	}

	d.logger.Error(fmt.Sprintf("Failed to find seal preset: %v", seal))
	return 1 // Default to 1 if we can't determine the layout
}

func (d Diablo) initPaths() Diablo {
	d.entranceToStar = []data.Position{{X: 7794, Y: 5517}, {X: 7791, Y: 5491}, {X: 7768, Y: 5459}, {X: 7775, Y: 5424}, {X: 7817, Y: 5458}, {X: 7777, Y: 5408}, {X: 7769, Y: 5379}, {X: 7777, Y: 5357}, {X: 7809, Y: 5359}, {X: 7805, Y: 5330}, {X: 7780, Y: 5317}, {X: 7791, Y: 5293}}
	d.starToVizA = []data.Position{{X: 7759, Y: 5295}, {X: 7734, Y: 5295}, {X: 7716, Y: 5295}, {X: 7718, Y: 5276}, {X: 7697, Y: 5292}, {X: 7678, Y: 5293}, {X: 7665, Y: 5276}, {X: 7662, Y: 5314}}
	d.starToVizB = []data.Position{{X: 7759, Y: 5295}, {X: 7734, Y: 5295}, {X: 7716, Y: 5295}, {X: 7701, Y: 5315}, {X: 7666, Y: 5313}, {X: 7653, Y: 5284}}
	d.starToSeisA = []data.Position{{X: 7781, Y: 5259}, {X: 7805, Y: 5258}, {X: 7802, Y: 5237}, {X: 7776, Y: 5228}, {X: 7775, Y: 5205}, {X: 7804, Y: 5193}, {X: 7814, Y: 5169}, {X: 7788, Y: 5153}}
	d.starToSeisB = []data.Position{{X: 7781, Y: 5259}, {X: 7805, Y: 5258}, {X: 7802, Y: 5237}, {X: 7776, Y: 5228}, {X: 7811, Y: 5218}, {X: 7807, Y: 5194}, {X: 7779, Y: 5193}, {X: 7774, Y: 5160}, {X: 7803, Y: 5154}}
	d.starToInfA = []data.Position{{X: 7809, Y: 5268}, {X: 7834, Y: 5306}, {X: 7852, Y: 5280}, {X: 7852, Y: 5310}, {X: 7869, Y: 5294}, {X: 7895, Y: 5295}, {X: 7919, Y: 5290}}
	d.starToInfB = []data.Position{{X: 7809, Y: 5268}, {X: 7834, Y: 5306}, {X: 7852, Y: 5280}, {X: 7852, Y: 5310}, {X: 7869, Y: 5294}, {X: 7895, Y: 5274}, {X: 7927, Y: 5275}, {X: 7932, Y: 5297}, {X: 7923, Y: 5313}}
	return d
}

func (d Diablo) killVizier() action.Action {
	return action.NewChain(func(gameData game.Data) []action.Action {
		d.logger.Debug("Moving to Vizier seal")
		return []action.Action{
			d.builder.MoveTo(func(gameData game.Data) (data.Position, bool) {
				seal4, _ := gameData.Objects.FindOne(object.DiabloSeal4)
				return seal4.Position, true
			}, step.StopAtDistance(20)),
			action.NewChain(func(gameData game.Data) []action.Action {
				seal4, _ := gameData.Objects.FindOne(object.DiabloSeal4)
				bestCorner := d.getLessConcurredCornerAroundSeal(gameData, seal4.Position)
				return []action.Action{
					d.builder.MoveToCoords(bestCorner),
					d.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
					d.builder.ItemPickup(false, 30),
				}
			}),
			d.activateSeal(object.DiabloSeal4),

			d.builder.MoveTo(func(gameData game.Data) (data.Position, bool) {
				seal5, _ := gameData.Objects.FindOne(object.DiabloSeal5)
				return seal5.Position, true
			}, step.StopAtDistance(20)),
			action.NewChain(func(gameData game.Data) []action.Action {
				seal5, _ := gameData.Objects.FindOne(object.DiabloSeal5)
				bestCorner := d.getLessConcurredCornerAroundSeal(gameData, seal5.Position)
				return []action.Action{
					d.builder.MoveToCoords(bestCorner),
					d.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
					d.builder.ItemPickup(false, 30),
				}
			}),
			d.activateSeal(object.DiabloSeal5),
			d.moveToVizierSpawn(),
			d.builder.Wait(time.Millisecond * 500),
			d.killSealElite(),
			d.builder.ItemPickup(false, 40),
		}
	})
}

func (d Diablo) killSeis() action.Action {
	return action.NewChain(func(gameData game.Data) []action.Action {
		d.logger.Debug("Moving to Seis seal")
		return []action.Action{
			d.builder.MoveTo(func(gameData game.Data) (data.Position, bool) {
				seal3, _ := gameData.Objects.FindOne(object.DiabloSeal3)
				return seal3.Position, true
			}, step.StopAtDistance(20)),
			action.NewChain(func(gameData game.Data) []action.Action {
				seal3, _ := gameData.Objects.FindOne(object.DiabloSeal3)
				bestCorner := d.getLessConcurredCornerAroundSeal(gameData, seal3.Position)
				return []action.Action{
					d.builder.MoveToCoords(bestCorner),
					d.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
					d.builder.ItemPickup(false, 30),
				}
			}),
			d.activateSeal(object.DiabloSeal3),
			d.moveToSeisSpawn(),
			d.builder.Wait(time.Millisecond * 500),
			d.killSealElite(),
			d.builder.ItemPickup(false, 40),
		}
	})
}

func (d Diablo) killInfector() action.Action {
	return action.NewChain(func(gameData game.Data) []action.Action {
		d.logger.Debug("Moving to Infector seal")
		return []action.Action{
			d.builder.MoveTo(func(gameData game.Data) (data.Position, bool) {
				seal1, _ := gameData.Objects.FindOne(object.DiabloSeal1)
				return seal1.Position, true
			}, step.StopAtDistance(20)),
			action.NewChain(func(gameData game.Data) []action.Action {
				seal1, _ := gameData.Objects.FindOne(object.DiabloSeal1)
				bestCorner := d.getLessConcurredCornerAroundSeal(gameData, seal1.Position)
				return []action.Action{
					d.builder.MoveToCoords(bestCorner),
					d.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
					d.builder.ItemPickup(false, 30),
				}
			}),
			d.activateSeal(object.DiabloSeal1),
			d.moveToInfectorSpawn(),
			d.builder.Wait(time.Millisecond * 500),
			d.killSealElite(),
			d.builder.ItemPickup(false, 40),

			d.builder.MoveTo(func(gameData game.Data) (data.Position, bool) {
				seal2, _ := gameData.Objects.FindOne(object.DiabloSeal2)
				return seal2.Position, true
			}, step.StopAtDistance(20)),
			action.NewChain(func(gameData game.Data) []action.Action {
				seal2, _ := gameData.Objects.FindOne(object.DiabloSeal2)
				bestCorner := d.getLessConcurredCornerAroundSeal(gameData, seal2.Position)
				return []action.Action{
					d.builder.MoveToCoords(bestCorner),
					d.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()),
					d.builder.ItemPickup(false, 30),
				}
			}),
			d.activateSeal(object.DiabloSeal2),
			d.builder.ItemPickup(false, 30),
		}
	})
}

func (d Diablo) killSealElite() action.Action {
	return action.NewChain(func(gameData game.Data) []action.Action {
		d.logger.Debug("Waiting for and killing seal elite")
		startTime := time.Now()
		var actions []action.Action

		actions = append(actions, action.NewStepChain(func(gameData game.Data) []step.Step {
			for _, m := range gameData.Monsters.Enemies(data.MonsterEliteFilter()) {
				if d.builder.IsMonsterSealElite(m) {
					d.logger.Debug("Seal defender found!")
					return nil // Exit the step chain when elite is found
				}
			}
			if time.Since(startTime) < time.Second*5 {
				return []step.Step{step.Wait(time.Millisecond * 100)}
			}
			d.logger.Debug("No seal elite found within 5 seconds")
			return nil
		}, action.RepeatUntilNoSteps()))

		// After the wait, attempt to clear and kill
		actions = append(actions, action.NewChain(func(gameData game.Data) []action.Action {
			for _, m := range gameData.Monsters.Enemies(data.MonsterEliteFilter()) {
				if d.builder.IsMonsterSealElite(m) {
					return []action.Action{
						// Clear normal monsters around the elite
						d.builder.ClearAreaAroundPlayer(20, func(m data.Monsters) []data.Monster {
							return m.Enemies(func(m data.Monsters) []data.Monster {
								return slices.DeleteFunc(m, func(monster data.Monster) bool {
									return d.builder.IsMonsterSealElite(monster)
								})
							})
						}),
						// Kill the seal elite
						d.char.KillMonsterSequence(func(dat game.Data) (data.UnitID, bool) {
							for _, m := range dat.Monsters.Enemies(data.MonsterEliteFilter()) {
								if d.builder.IsMonsterSealElite(m) {
									_, _, found := d.PathFinder.GetPath(dat, m.Position)
									if found {
										d.logger.Debug(fmt.Sprintf("Attempting to kill seal elite: %v", m.Name))
										return m.UnitID, true
									}
								}
							}
							d.logger.Debug("Seal elite has been killed or is not found")
							return 0, false
						}, nil),
					}
				}
			}
			d.logger.Debug("No seal elite found after waiting")
			return nil
		}))

		// Always perform item pickup at the end
		actions = append(actions, d.builder.ItemPickup(false, 40))

		return actions
	})
}

func (d Diablo) activateSeal(seal object.Name) action.Action {
	return action.NewChain(func(gameData game.Data) []action.Action {
		obj, _ := gameData.Objects.FindOne(seal)

		// Check for the bugged seal
		if seal == object.DiabloSeal3 && obj.Position.X == 7773 && obj.Position.Y == 5155 {
			return []action.Action{
				d.builder.MoveToCoords(data.Position{
					X: 7768,
					Y: 5160,
				}),
				d.builder.InteractObject(seal, func(gameData game.Data) bool {
					obj, found := gameData.Objects.FindOne(seal)
					if found {
						if !obj.Selectable {
							d.logger.Debug(fmt.Sprintf("Seal activated: %v", seal))
						}
						return !obj.Selectable
					}
					return false
				}),
			}
		}

		// Normal seal activation
		return []action.Action{
			d.builder.InteractObject(seal, func(gameData game.Data) bool {
				obj, found := gameData.Objects.FindOne(seal)
				if found {
					if !obj.Selectable {
						d.logger.Debug(fmt.Sprintf("Seal activated: %v", seal))
					}
					return !obj.Selectable
				}
				return false
			}),
		}
	})
}

func (d Diablo) moveToVizierSpawn() action.Action {
	return action.NewChain(func(gameData game.Data) []action.Action {
		var actions []action.Action

		if d.vizLayout == 1 {
			d.logger.Debug("Moving to X: 7664, Y: 5305 - vizLayout 1")
			actions = append(actions, d.builder.MoveToCoords(data.Position{X: 7664, Y: 5305}))
		} else {
			d.logger.Debug("Moving to X: 7675, Y: 5284 - vizLayout 2")
			actions = append(actions, d.builder.MoveToCoords(data.Position{X: 7675, Y: 5284}))
		}

		// Check for nearby monsters after moving
		actions = append(actions, action.NewChain(func(gameData game.Data) []action.Action {
			for _, m := range gameData.Monsters.Enemies() {
				if dist := pather.DistanceFromMe(gameData, m.Position); dist < 4 {
					d.logger.Debug("Monster detected close to the player, clearing small radius")
					return []action.Action{d.builder.ClearAreaAroundPlayer(5, data.MonsterAnyFilter())}
				}
			}
			// If no nearby monsters, do nothing
			return nil
		}))

		return actions
	})
}

func (d Diablo) moveToSeisSpawn() action.Action {
	return action.NewChain(func(gameData game.Data) []action.Action {
		var actions []action.Action

		if d.seisLayout == 1 {
			d.logger.Debug("Moving to X: 7795, Y: 5195 - seisLayout 1")
			actions = append(actions, d.builder.MoveToCoords(data.Position{X: 7795, Y: 5195}))
		} else {
			d.logger.Debug("Moving to X: 7795, Y: 5155 - seisLayout 2")
			actions = append(actions, d.builder.MoveToCoords(data.Position{X: 7795, Y: 5155}))
		}

		// Check for nearby monsters after moving
		actions = append(actions, action.NewChain(func(gameData game.Data) []action.Action {
			for _, m := range gameData.Monsters.Enemies() {
				if dist := pather.DistanceFromMe(gameData, m.Position); dist < 4 {
					d.logger.Debug("Monster detected close to the player, clearing small radius")
					return []action.Action{d.builder.ClearAreaAroundPlayer(5, data.MonsterAnyFilter())}
				}
			}
			// If no nearby monsters, do nothing
			return nil
		}))

		return actions
	})
}

func (d Diablo) moveToInfectorSpawn() action.Action {
	return action.NewChain(func(gameData game.Data) []action.Action {
		var actions []action.Action

		if d.infLayout == 1 {
			d.logger.Debug("Moving to X: 7894, Y: 5294 - infLayout 1")
			actions = append(actions, d.builder.MoveToCoords(data.Position{X: 7894, Y: 5294}))
		} else {
			d.logger.Debug("Moving to X: 7928, Y: 5296 - infLayout 2")
			actions = append(actions, d.builder.MoveToCoords(data.Position{X: 7928, Y: 5296}))
		}

		// Check for nearby monsters after moving
		actions = append(actions, action.NewChain(func(gameData game.Data) []action.Action {
			for _, m := range gameData.Monsters.Enemies() {
				if dist := pather.DistanceFromMe(gameData, m.Position); dist < 4 {
					d.logger.Debug("Monster detected close to the player, clearing small radius")
					return []action.Action{d.builder.ClearAreaAroundPlayer(5, data.MonsterAnyFilter())}
				}
			}
			// If no nearby monsters, do nothing
			return nil
		}))

		return actions
	})
}

func (d Diablo) entranceToStarClear() []action.Action {
	onlyElites := d.CharacterCfg.Game.Diablo.FocusOnElitePacks

	monsterFilter := func(monsters data.Monsters) []data.Monster {
		filteredMonsters := skipStormCasterFilter(monsters)
		if onlyElites {
			return data.MonsterEliteFilter()(filteredMonsters)
		}
		return filteredMonsters
	}

	d.logger.Debug("Clearing path from entrance to star")
	return d.clearPath(d.entranceToStar, monsterFilter)
}

func (d Diablo) starToVizClear() []action.Action {
	onlyElites := d.CharacterCfg.Game.Diablo.FocusOnElitePacks

	monsterFilter := func(monsters data.Monsters) []data.Monster {
		filteredMonsters := skipStormCasterFilter(monsters)
		if onlyElites {
			return data.MonsterEliteFilter()(filteredMonsters)
		}
		return filteredMonsters
	}

	path := d.starToVizA
	if d.vizLayout == 2 {
		path = d.starToVizB
	}
	d.logger.Debug("Clearing path from star to Vizier")
	return d.clearPath(path, monsterFilter)
}

func (d Diablo) starToSeisClear() []action.Action {
	onlyElites := d.CharacterCfg.Game.Diablo.FocusOnElitePacks

	monsterFilter := func(monsters data.Monsters) []data.Monster {
		filteredMonsters := skipStormCasterFilter(monsters)
		if onlyElites {
			return data.MonsterEliteFilter()(filteredMonsters)
		}
		return filteredMonsters
	}

	path := d.starToSeisA
	if d.seisLayout == 2 {
		path = d.starToSeisB
	}
	d.logger.Debug("Clearing path from star to Seis")
	return d.clearPath(path, monsterFilter)
}

func (d Diablo) starToInfClear() []action.Action {
	onlyElites := d.CharacterCfg.Game.Diablo.FocusOnElitePacks

	monsterFilter := func(monsters data.Monsters) []data.Monster {
		filteredMonsters := skipStormCasterFilter(monsters)
		if onlyElites {
			return data.MonsterEliteFilter()(filteredMonsters)
		}
		return filteredMonsters
	}

	path := d.starToInfA
	if d.infLayout == 2 {
		path = d.starToInfB
	}
	d.logger.Debug("Clearing path from star to Infector")
	return d.clearPath(path, monsterFilter)
}

func (d Diablo) clearPath(path []data.Position, monsterFilter func(data.Monsters) []data.Monster) []action.Action {
	var actions []action.Action
	var maxPosDiff = 20

	actions = append(actions, d.builder.Buff())

	for _, pos := range path {
		actions = append(actions,
			action.NewChain(func(gameData game.Data) []action.Action {
				multiplier := 1
				if pather.IsWalkable(pos, gameData.AreaOrigin, gameData.CollisionGrid) {
					return []action.Action{d.builder.MoveToCoords(pos)}
				}
				for range 2 {
					for i := 1; i < maxPosDiff; i++ {
						newPos := data.Position{X: pos.X + (i * multiplier), Y: pos.Y + (i * multiplier)}
						if pather.IsWalkable(newPos, gameData.AreaOrigin, gameData.CollisionGrid) {
							return []action.Action{d.builder.MoveToCoords(newPos)}
						}
					}
					multiplier *= -1
				}
				return []action.Action{d.builder.MoveToCoords(pos)}
			}),
			d.builder.ClearAreaAroundPlayer(35, func(m data.Monsters) []data.Monster {
				monsters := monsterFilter(m)
				return skipStormCasterFilter(monsters)
			}),
			d.builder.ItemPickup(false, 35),
		)
		d.cleared = append(d.cleared, pos)
	}

	actions = append(actions, d.clearStrays(monsterFilter)...)

	return actions
}

func (d Diablo) clearStrays(monsterFilter data.MonsterFilter) []action.Action {
	d.logger.Debug("Clearing potential stray monsters")
	return []action.Action{
		action.NewChain(func(gameData game.Data) []action.Action {
			var actions []action.Action
			oldPos := gameData.PlayerUnit.Position

			monsters := monsterFilter(gameData.Monsters)
			monsters = skipStormCasterFilter(monsters)
			for _, monster := range monsters {
				for _, clearedPos := range d.cleared {
					if pather.DistanceFromPoint(monster.Position, clearedPos) < 30 {
						actions = append(actions,
							d.builder.MoveToCoords(monster.Position),
							d.builder.ClearAreaAroundPlayer(15, func(m data.Monsters) []data.Monster {
								return skipStormCasterFilter(monsterFilter(m))
							}),
						)
						break
					}
				}
			}

			if len(actions) > 0 {
				actions = append(actions, d.builder.MoveToCoords(oldPos))
			}

			return actions
		}),
	}
}

func skipStormCasterFilter(monsters data.Monsters) []data.Monster {
	var stormCasterIds = []npc.ID{npc.StormCaster, npc.StormCaster2}
	var filteredMonsters []data.Monster

	for _, m := range monsters {
		if !slices.Contains(stormCasterIds, m.Name) {
			filteredMonsters = append(filteredMonsters, m)
		}
	}

	return filteredMonsters
}

func (d Diablo) getLessConcurredCornerAroundSeal(gameData game.Data, sealPosition data.Position) data.Position {
	corners := [4]data.Position{
		{X: sealPosition.X + 7, Y: sealPosition.Y + 7},
		{X: sealPosition.X - 7, Y: sealPosition.Y + 7},
		{X: sealPosition.X - 7, Y: sealPosition.Y - 7},
		{X: sealPosition.X + 7, Y: sealPosition.Y - 7},
	}
	bestCorner := 0
	bestCornerDistance := 0
	for i, c := range corners {
		averageDistance := 0
		monstersFound := 0
		for _, m := range gameData.Monsters.Enemies() {
			distance := pather.DistanceFromPoint(c, m.Position)
			if distance < 5 {
				monstersFound++
				averageDistance += pather.DistanceFromPoint(c, m.Position)
			}
		}
		if averageDistance > bestCornerDistance {
			bestCorner = i
			bestCornerDistance = averageDistance
		}
		if monstersFound == 0 {
			d.logger.Debug("Moving to corner", slog.Int("corner", i), slog.Int("monsters", monstersFound))
			return corners[i]
		}
		d.logger.Debug("Corner", slog.Int("corner", i), slog.Int("monsters", monstersFound), slog.Int("distance", averageDistance))
	}
	d.logger.Debug("Moving to corner", slog.Int("corner", bestCorner), slog.Int("monsters", bestCornerDistance))
	return corners[bestCorner]
}
