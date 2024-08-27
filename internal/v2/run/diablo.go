package run

import (
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/v2/action"
	"github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/pather"
)

var diabloSpawnPosition = data.Position{X: 7792, Y: 5294}
var chaosSanctuaryEntrancePosition = data.Position{X: 7790, Y: 5544}

type Diablo struct {
	ctx            *context.Status
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

func NewDiablo() *Diablo {
	return &Diablo{
		ctx: context.Get(),
	}
}

func (d Diablo) Name() string {
	return string(config.DiabloRun)
}

func (d Diablo) Run() error {
	d = d.initLayout()
	d = d.initPaths()

	err := action.WayPoint(area.RiverOfFlame)
	if err != nil {
		return err
	}
	action.Buff()

	if d.ctx.CharacterCfg.Game.Diablo.FullClear {
		err = action.MoveToCoords(chaosSanctuaryEntrancePosition)
		if err != nil {
			return err
		}
	} else {
		err = action.MoveToCoords(diabloSpawnPosition)
		if err != nil {
			return err
		}
	}

	if d.ctx.CharacterCfg.Companion.Leader {
		action.OpenTPIfLeader()
		action.Buff()
		action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())
	}

	if d.ctx.CharacterCfg.Game.Diablo.FullClear {
		d.entranceToStarClear()
	}

	if d.ctx.CharacterCfg.Game.Diablo.FullClear {
		d.starToVizClear()
	}

	err = d.killVizier()
	if err != nil {
		return err
	}

	if d.ctx.CharacterCfg.Game.Diablo.FullClear {
		d.starToSeisClear()
	}

	err = d.killSeis()
	if err != nil {
		return err
	}

	if d.ctx.CharacterCfg.Game.Diablo.FullClear {
		d.starToInfClear()
	}

	err = d.killInfector()
	if err != nil {
		return err
	}

	if d.ctx.CharacterCfg.Game.Diablo.KillDiablo {
		action.Buff()
		action.MoveToCoords(diabloSpawnPosition)
		err = d.ctx.Char.KillDiablo()
		if err != nil {
			return err
		}
	}

	return nil
}

func (d Diablo) initLayout() Diablo {
	d.vizLayout = d.getLayout(object.DiabloSeal4, 5275)
	d.seisLayout = d.getLayout(object.DiabloSeal3, 7773)
	d.infLayout = d.getLayout(object.DiabloSeal1, 7893)

	d.ctx.Logger.Debug(fmt.Sprintf("Layouts initialized - Vizier: %d, Seis: %d, Infector: %d", d.vizLayout, d.seisLayout, d.infLayout))
	return d
}

func (d Diablo) getLayout(seal object.Name, value int) int {
	mapData := d.ctx.GameReader.GetCachedMapData(false)
	origin := mapData.Origin(area.ChaosSanctuary)
	_, _, objects, _ := mapData.NPCsExitsAndObjects(origin, area.ChaosSanctuary)

	for _, obj := range objects {
		if obj.Name == seal {
			if obj.Position.Y == value || obj.Position.X == value {
				d.ctx.Logger.Debug(fmt.Sprintf("Layout 1 detected for seal %v: position matches value %d", seal, value))
				return 1
			}
			d.ctx.Logger.Debug(fmt.Sprintf("Layout 2 detected for seal %v: position does not match value %d", seal, value))
			return 2
		}
	}

	d.ctx.Logger.Error(fmt.Sprintf("Failed to find seal preset: %v", seal))
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

func (d Diablo) killVizier() error {
	d.ctx.Logger.Debug("Moving to Vizier seal")

	err := action.MoveTo(func() (data.Position, bool) {
		seal4, _ := d.ctx.Data.Objects.FindOne(object.DiabloSeal4)
		return seal4.Position, true
	})
	if err != nil {
		return err
	}

	seal4, _ := d.ctx.Data.Objects.FindOne(object.DiabloSeal4)
	bestCorner := d.getLessConcurredCornerAroundSeal(seal4.Position)

	action.MoveToCoords(bestCorner)
	action.ClearAreaAroundPlayer(20, data.MonsterAnyFilter())
	d.activateSeal(object.DiabloSeal4)

	err = action.MoveTo(func() (data.Position, bool) {
		seal5, _ := d.ctx.Data.Objects.FindOne(object.DiabloSeal5)
		return seal5.Position, true
	})
	if err != nil {
		return err
	}

	seal5, _ := d.ctx.Data.Objects.FindOne(object.DiabloSeal5)
	bestCorner = d.getLessConcurredCornerAroundSeal(seal5.Position)

	action.MoveToCoords(bestCorner)
	action.ClearAreaAroundPlayer(20, data.MonsterAnyFilter())
	d.activateSeal(object.DiabloSeal5)

	d.moveToVizierSpawn()
	time.Sleep(500)

	err = d.killSealElite()
	if err != nil {
		return err
	}

	return nil
}

func (d Diablo) killSeis() error {
	d.ctx.Logger.Debug("Moving to Seis seal")

	err := action.MoveTo(func() (data.Position, bool) {
		seal3, _ := d.ctx.Data.Objects.FindOne(object.DiabloSeal3)
		return seal3.Position, true
	})
	if err != nil {
		return err
	}

	seal3, _ := d.ctx.Data.Objects.FindOne(object.DiabloSeal3)
	bestCorner := d.getLessConcurredCornerAroundSeal(seal3.Position)

	action.MoveToCoords(bestCorner)
	action.ClearAreaAroundPlayer(20, data.MonsterAnyFilter())
	d.activateSeal(object.DiabloSeal3)

	d.moveToSeisSpawn()
	time.Sleep(500)

	err = d.killSealElite()
	if err != nil {
		return err
	}

	return nil
}

func (d Diablo) killInfector() error {
	d.ctx.Logger.Debug("Moving to Infector seal")

	err := action.MoveTo(func() (data.Position, bool) {
		seal1, _ := d.ctx.Data.Objects.FindOne(object.DiabloSeal1)
		return seal1.Position, true
	})
	if err != nil {
		return err
	}

	seal1, _ := d.ctx.Data.Objects.FindOne(object.DiabloSeal1)
	bestCorner := d.getLessConcurredCornerAroundSeal(seal1.Position)

	action.MoveToCoords(bestCorner)
	action.ClearAreaAroundPlayer(20, data.MonsterAnyFilter())
	d.activateSeal(object.DiabloSeal1)

	err = action.MoveTo(func() (data.Position, bool) {
		seal2, _ := d.ctx.Data.Objects.FindOne(object.DiabloSeal2)
		return seal2.Position, true
	})
	if err != nil {
		return err
	}

	seal2, _ := d.ctx.Data.Objects.FindOne(object.DiabloSeal2)
	bestCorner = d.getLessConcurredCornerAroundSeal(seal2.Position)

	action.MoveToCoords(bestCorner)
	action.ClearAreaAroundPlayer(20, data.MonsterAnyFilter())
	d.activateSeal(object.DiabloSeal2)

	d.moveToInfectorSpawn()
	time.Sleep(500)

	err = d.killSealElite()
	if err != nil {
		return err
	}

	return nil
}

func (d Diablo) killSealElite() error {
	d.ctx.Logger.Debug("Waiting for and killing seal elite")
	startTime := time.Now()

	for _, m := range d.ctx.Data.Monsters.Enemies(data.MonsterEliteFilter()) {
		if action.IsMonsterSealElite(m) {
			d.ctx.Logger.Debug("Seal defender found!")
			return nil // Exit the step chain when elite is found
		}
	}

	if time.Since(startTime) < time.Second*5 {
		time.Sleep(100 * time.Millisecond)
	}

	d.ctx.Logger.Debug("No seal elite found within 5 seconds")

	// After the wait, attempt to clear and kill
	for _, m := range d.ctx.Data.Monsters.Enemies(data.MonsterEliteFilter()) {
		if action.IsMonsterSealElite(m) {
			// Clear normal monsters around the elite
			action.ClearAreaAroundPlayer(20, func(m data.Monsters) []data.Monster {
				return m.Enemies(func(m data.Monsters) []data.Monster {
					return slices.DeleteFunc(m, func(monster data.Monster) bool {
						return action.IsMonsterSealElite(monster)
					})
				})
			})

			// Kill the seal elite
			d.ctx.Char.KillMonsterSequence(func(dat game.Data) (data.UnitID, bool) {
				for _, m := range dat.Monsters.Enemies(data.MonsterEliteFilter()) {
					if action.IsMonsterSealElite(m) {
						_, _, found := d.ctx.PathFinder.GetPath(m.Position)
						if found {
							d.ctx.Logger.Debug(fmt.Sprintf("Attempting to kill seal elite: %v", m.Name))
							return m.UnitID, true
						}
					}
				}
				d.ctx.Logger.Debug("Seal elite has been killed or is not found")
				return 0, false
			}, nil)
		}
	}

	d.ctx.Logger.Debug("No seal elite found after waiting")

	return nil
}

func (d Diablo) activateSeal(seal object.Name) error {
	obj, found := d.ctx.Data.Objects.FindOne(seal)
	if !found {
		return fmt.Errorf("seal %v not found", seal)
	}

	// Check for the bugged seal
	if seal == object.DiabloSeal3 && obj.Position.X == 7773 && obj.Position.Y == 5155 {
		if err := action.MoveToCoords(data.Position{X: 7768, Y: 5160}); err != nil {
			return fmt.Errorf("failed to move to bugged seal position: %w", err)
		}
	}

	// Interact with the seal (bugged or normal)
	if err := action.InteractObject(obj, func() bool {
		updatedObj, found := d.ctx.Data.Objects.FindOne(seal)
		if found {
			if !updatedObj.Selectable {
				d.ctx.Logger.Debug(fmt.Sprintf("Seal activated: %v", seal))
			}
			return !updatedObj.Selectable
		}
		return false
	}); err != nil {
		return fmt.Errorf("failed to interact with seal: %w", err)
	}

	return nil
}

func (d Diablo) moveToVizierSpawn() error {
	if d.vizLayout == 1 {
		d.ctx.Logger.Debug("Moving to X: 7664, Y: 5305 - vizLayout 1")
		action.MoveToCoords(data.Position{X: 7664, Y: 5305})
	} else {
		d.ctx.Logger.Debug("Moving to X: 7675, Y: 5284 - vizLayout 2")
		action.MoveToCoords(data.Position{X: 7675, Y: 5284})
	}

	// Check for nearby monsters after moving
	for _, m := range d.ctx.Data.Monsters.Enemies() {
		if dist := d.ctx.PathFinder.DistanceFromMe(m.Position); dist < 4 {
			d.ctx.Logger.Debug("Monster detected close to the player, clearing small radius")
			action.ClearAreaAroundPlayer(5, data.MonsterAnyFilter())
		}
	}
	// If no nearby monsters, do nothing
	return nil
}

func (d Diablo) moveToSeisSpawn() error {
	if d.seisLayout == 1 {
		d.ctx.Logger.Debug("Moving to X: 7795, Y: 5195 - seisLayout 1")
		action.MoveToCoords(data.Position{X: 7795, Y: 5195})
	} else {
		d.ctx.Logger.Debug("Moving to X: 7795, Y: 5155 - seisLayout 2")
		action.MoveToCoords(data.Position{X: 7795, Y: 5155})
	}

	// Check for nearby monsters after moving
	for _, m := range d.ctx.Data.Monsters.Enemies() {
		if dist := d.ctx.PathFinder.DistanceFromMe(m.Position); dist < 4 {
			d.ctx.Logger.Debug("Monster detected close to the player, clearing small radius")
			action.ClearAreaAroundPlayer(5, data.MonsterAnyFilter())
		}
	}
	// If no nearby monsters, do nothing
	return nil
}

func (d Diablo) moveToInfectorSpawn() error {
	if d.infLayout == 1 {
		d.ctx.Logger.Debug("Moving to X: 7894, Y: 5294 - infLayout 1")
		action.MoveToCoords(data.Position{X: 7894, Y: 5294})
	} else {
		d.ctx.Logger.Debug("Moving to X: 7928, Y: 5296 - infLayout 2")
		action.MoveToCoords(data.Position{X: 7928, Y: 5296})
	}

	// Check for nearby monsters after moving
	for _, m := range d.ctx.Data.Monsters.Enemies() {
		if dist := d.ctx.PathFinder.DistanceFromMe(m.Position); dist < 4 {
			d.ctx.Logger.Debug("Monster detected close to the player, clearing small radius")
			action.ClearAreaAroundPlayer(5, data.MonsterAnyFilter())
		}
	}
	// If no nearby monsters, do nothing
	return nil
}

func (d Diablo) entranceToStarClear() error {
	onlyElites := d.ctx.CharacterCfg.Game.Diablo.FocusOnElitePacks

	monsterFilter := func(monsters data.Monsters) []data.Monster {
		filteredMonsters := skipStormCasterFilter(monsters)
		if onlyElites {
			return data.MonsterEliteFilter()(filteredMonsters)
		}
		return filteredMonsters
	}

	d.ctx.Logger.Debug("Clearing path from entrance to star")
	return d.clearPath(d.entranceToStar, monsterFilter)
}

func (d Diablo) starToVizClear() error {
	onlyElites := d.ctx.CharacterCfg.Game.Diablo.FocusOnElitePacks

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
	d.ctx.Logger.Debug("Clearing path from star to Vizier")
	return d.clearPath(path, monsterFilter)
}

func (d Diablo) starToSeisClear() error {
	onlyElites := d.ctx.CharacterCfg.Game.Diablo.FocusOnElitePacks

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
	d.ctx.Logger.Debug("Clearing path from star to Seis")
	return d.clearPath(path, monsterFilter)
}

func (d Diablo) starToInfClear() error {
	onlyElites := d.ctx.CharacterCfg.Game.Diablo.FocusOnElitePacks

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
	d.ctx.Logger.Debug("Clearing path from star to Infector")
	return d.clearPath(path, monsterFilter)
}

func (d Diablo) clearPath(path []data.Position, monsterFilter func(data.Monsters) []data.Monster) error {
	maxPosDiff := 20 // Keep this as a constant for now

	// Create a local random number generator
	localRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, pos := range path {
		if pather.IsWalkable(pos, d.ctx.Data.AreaOrigin, d.ctx.Data.CollisionGrid) {
			action.MoveToCoords(pos)
		}

		// Directional zigzag: alternate between diagonal, horizontal, and vertical
		directions := []struct{ dx, dy int }{{1, 1}, {1, 0}, {0, 1}}

		for _, dir := range directions {
			multiplier := 1
			for range 2 {
				for i := 1; i < maxPosDiff; i++ {
					// Random offset: add a small random variation to each point
					offsetX := localRand.Intn(3) - 1 // Random offset of -1, 0, or 1
					offsetY := localRand.Intn(3) - 1 // Random offset of -1, 0, or 1

					newPos := data.Position{
						X: pos.X + (i * multiplier * dir.dx) + offsetX,
						Y: pos.Y + (i * multiplier * dir.dy) + offsetY,
					}

					if pather.IsWalkable(newPos, d.ctx.Data.AreaOrigin, d.ctx.Data.CollisionGrid) {
						action.MoveToCoords(newPos)
						action.ClearAreaAroundPlayer(35, monsterFilter)
					}
				}
				multiplier *= -1
			}
		}

		d.cleared = append(d.cleared, pos)
	}

	d.clearStrays(monsterFilter)

	return nil
}

func (d Diablo) clearStrays(monsterFilter data.MonsterFilter) error {
	d.ctx.Logger.Debug("Clearing potential stray monsters")
	oldPos := d.ctx.Data.PlayerUnit.Position

	// Apply both filters at once
	monsterfilter := func(m data.Monsters) []data.Monster {
		return skipStormCasterFilter(monsterFilter(m))
	}

	monsters := monsterfilter(d.ctx.Data.Monsters)
	d.ctx.Logger.Debug(fmt.Sprintf("Stray monsters to clear after filtering: %d", len(monsters)))

	actionPerformed := false
	for _, monster := range monsters {
		for _, clearedPos := range d.cleared {
			if pather.DistanceFromPoint(monster.Position, clearedPos) < 30 {
				action.MoveToCoords(monster.Position)
				action.ClearAreaAroundPlayer(15, monsterfilter)
				actionPerformed = true
				break
			}
		}
		if actionPerformed {
			break
		}
	}

	if actionPerformed {
		action.MoveToCoords(oldPos)
	}

	return nil
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

func (d Diablo) getLessConcurredCornerAroundSeal(sealPosition data.Position) data.Position {
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
		for _, m := range d.ctx.Data.Monsters.Enemies() {
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
			d.ctx.Logger.Debug("Moving to corner", slog.Int("corner", i), slog.Int("monsters", monstersFound))
			return corners[i]
		}
		d.ctx.Logger.Debug("Corner", slog.Int("corner", i), slog.Int("monsters", monstersFound), slog.Int("distance", averageDistance))
	}
	d.ctx.Logger.Debug("Moving to corner", slog.Int("corner", bestCorner), slog.Int("monsters", bestCornerDistance))
	return corners[bestCorner]
}
