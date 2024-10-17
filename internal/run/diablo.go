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
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

var diabloSpawnPosition = data.Position{X: 7792, Y: 5294}
var chaosSanctuaryEntrancePosition = data.Position{X: 7790, Y: 5544}

type Diablo struct {
	ctx             *context.Status
	vizLayout       int
	seisLayout      int
	infLayout       int
	cleared         []data.Position
	paths           map[string][]data.Position
	isKillingVizier bool
}

func NewDiablo() *Diablo {
	return &Diablo{
		ctx:   context.Get(),
		paths: make(map[string][]data.Position),
	}
}

func (d *Diablo) Name() string {
	return string(config.DiabloRun)
}

func (d *Diablo) Run() error {
	if err := action.WayPoint(area.RiverOfFlame); err != nil {
		return err
	}

	targetPosition := diabloSpawnPosition
	if d.ctx.CharacterCfg.Game.Diablo.FullClear {
		targetPosition = chaosSanctuaryEntrancePosition
	}

	if err := action.MoveToCoords(targetPosition); err != nil {
		return err
	}

	d.initLayout()
	d.initPaths()

	if d.ctx.CharacterCfg.Companion.Leader {
		action.OpenTPIfLeader()
		action.Buff()
		action.ClearAreaAroundPlayer(30, d.getMonsterFilter())
	}

	if d.ctx.CharacterCfg.Game.Diablo.FullClear {
		if err := d.clearPath("entranceToStar", d.getMonsterFilter()); err != nil {
			return err
		}
	}

	for _, boss := range []string{"Vizier", "Seis", "Infector"} {
		if d.ctx.CharacterCfg.Game.Diablo.FullClear {
			if err := d.clearPath(fmt.Sprintf("starTo%s", boss), d.getMonsterFilter()); err != nil {
				return err
			}
		}

		if err := d.killBoss(boss); err != nil {
			return err
		}
	}

	if d.ctx.CharacterCfg.Game.Diablo.KillDiablo {
		action.Buff()
		action.MoveToCoords(diabloSpawnPosition)

		// Check if we should disable item pickup for Diablo
		if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
			context.Get().DisableItemPickup()
		}

		if err := d.ctx.Char.KillDiablo(); err != nil {
			if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
				context.Get().EnableItemPickup() // Re-enable pickup if killing Diablo fails
			}
			return err
		}
		// Re-enable item pickup if it was disabled after diablo is defeated
		if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
			context.Get().EnableItemPickup()
		}

		// Now that it's safe, attempt to pick up items
		_ = action.ItemPickup(20)
	}

	return nil
}

func (d *Diablo) initLayout() {
	d.vizLayout = d.getLayout(object.DiabloSeal4, 5275)
	d.seisLayout = d.getLayout(object.DiabloSeal3, 7773)
	d.infLayout = d.getLayout(object.DiabloSeal1, 7893)
	d.ctx.Logger.Debug(fmt.Sprintf("Layouts initialized - Vizier: %d, Seis: %d, Infector: %d", d.vizLayout, d.seisLayout, d.infLayout))
}

func (d *Diablo) getLayout(seal object.Name, value int) int {
	for _, obj := range d.ctx.Data.AreaData.Objects {
		if obj.Name == seal {
			if obj.Position.Y == value || obj.Position.X == value {
				return 1
			}
			return 2
		}
	}
	d.ctx.Logger.Error(fmt.Sprintf("Failed to find seal preset: %v", seal))
	return 1
}

func (d *Diablo) initPaths() {
	d.paths["entranceToStar"] = []data.Position{{X: 7794, Y: 5517}, {X: 7791, Y: 5491}, {X: 7768, Y: 5459}, {X: 7775, Y: 5424}, {X: 7817, Y: 5458}, {X: 7777, Y: 5408}, {X: 7769, Y: 5379}, {X: 7777, Y: 5357}, {X: 7809, Y: 5359}, {X: 7805, Y: 5330}, {X: 7780, Y: 5317}, {X: 7791, Y: 5293}}
	d.paths["starToVizierA"] = []data.Position{{X: 7759, Y: 5295}, {X: 7734, Y: 5295}, {X: 7716, Y: 5295}, {X: 7718, Y: 5276}, {X: 7697, Y: 5292}, {X: 7678, Y: 5293}, {X: 7665, Y: 5276}, {X: 7662, Y: 5314}}
	d.paths["starToVizierB"] = []data.Position{{X: 7759, Y: 5295}, {X: 7734, Y: 5295}, {X: 7716, Y: 5295}, {X: 7701, Y: 5315}, {X: 7666, Y: 5313}, {X: 7653, Y: 5284}}
	d.paths["starToSeisA"] = []data.Position{{X: 7781, Y: 5259}, {X: 7805, Y: 5258}, {X: 7802, Y: 5237}, {X: 7776, Y: 5228}, {X: 7775, Y: 5205}, {X: 7804, Y: 5193}, {X: 7814, Y: 5169}, {X: 7788, Y: 5153}}
	d.paths["starToSeisB"] = []data.Position{{X: 7781, Y: 5259}, {X: 7805, Y: 5258}, {X: 7802, Y: 5237}, {X: 7776, Y: 5228}, {X: 7811, Y: 5218}, {X: 7807, Y: 5194}, {X: 7779, Y: 5193}, {X: 7774, Y: 5160}, {X: 7803, Y: 5154}}
	d.paths["starToInfectorA"] = []data.Position{{X: 7809, Y: 5268}, {X: 7834, Y: 5306}, {X: 7852, Y: 5280}, {X: 7852, Y: 5310}, {X: 7869, Y: 5294}, {X: 7895, Y: 5295}, {X: 7919, Y: 5290}}
	d.paths["starToInfectorB"] = []data.Position{{X: 7809, Y: 5268}, {X: 7834, Y: 5306}, {X: 7852, Y: 5280}, {X: 7852, Y: 5310}, {X: 7869, Y: 5294}, {X: 7895, Y: 5274}, {X: 7927, Y: 5275}, {X: 7932, Y: 5297}, {X: 7923, Y: 5313}}
}

func (d *Diablo) killBoss(boss string) error {
	d.ctx.Logger.Debug(fmt.Sprintf("Moving to %s seal", boss))

	// Set the current seal boss
	sealName := map[string]object.Name{
		"Vizier":   object.DiabloSeal4,
		"Seis":     object.DiabloSeal3,
		"Infector": object.DiabloSeal1,
	}[boss]

	d.isKillingVizier = (boss == "Vizier")

	if err := d.moveToBossArea(boss); err != nil {
		return err
	}

	seal, _ := d.ctx.Data.Objects.FindOne(sealName)
	bestCorner := d.getLessConcurredCornerAroundSeal(seal.Position)

	action.MoveToCoords(bestCorner)
	action.ClearAreaAroundPlayer(20, d.getMonsterFilter())

	// Check if we should disable item pickup while killing boss seals
	if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
		context.Get().DisableItemPickup()
	}

	if err := d.activateSeal(sealName); err != nil {
		if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
			context.Get().EnableItemPickup() // Re-enable pickup if seal activation fails
		}
		return err
	}

	if boss == "Vizier" || boss == "Infector" {
		secondSeal := map[string]object.Name{
			"Vizier":   object.DiabloSeal5,
			"Infector": object.DiabloSeal2,
		}[boss]
		if err := d.moveToBossArea(boss); err != nil {
			context.Get().EnableItemPickup() // Re-enable pickup if movement fails
			return err
		}
		seal, _ = d.ctx.Data.Objects.FindOne(secondSeal)
		bestCorner = d.getLessConcurredCornerAroundSeal(seal.Position)
		action.MoveToCoords(bestCorner)
		action.ClearAreaAroundPlayer(20, d.getMonsterFilter())
		if err := d.activateSeal(secondSeal); err != nil {
			context.Get().EnableItemPickup() // Re-enable pickup if seal activation fails
			return err
		}
	}

	d.moveToBossSpawn(boss)
	time.Sleep(1500 * time.Millisecond)

	err := d.killSealElite()
	// Re-enable item pickup if it was disabled
	if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
		context.Get().EnableItemPickup()
	}

	if err != nil {
		return err
	}

	// Now that it's safe, attempt to pick up items
	_ = action.ItemPickup(30)

	return nil
}

func (d *Diablo) moveToBossArea(boss string) error {
	sealName := map[string]object.Name{
		"Vizier":   object.DiabloSeal4,
		"Seis":     object.DiabloSeal3,
		"Infector": object.DiabloSeal1,
	}[boss]

	return action.MoveTo(func() (data.Position, bool) {
		seal, _ := d.ctx.Data.Objects.FindOne(sealName)
		return seal.Position, true
	})
}

func (d *Diablo) moveToBossSpawn(boss string) {
	spawnPositions := map[string]map[int]data.Position{
		"Vizier": {
			1: {X: 7664, Y: 5305},
			2: {X: 7675, Y: 5284},
		},
		"Seis": {
			1: {X: 7795, Y: 5195},
			2: {X: 7795, Y: 5155},
		},
		"Infector": {
			1: {X: 7894, Y: 5294},
			2: {X: 7928, Y: 5296},
		},
	}

	layout := map[string]int{
		"Vizier":   d.vizLayout,
		"Seis":     d.seisLayout,
		"Infector": d.infLayout,
	}[boss]

	spawnPos := spawnPositions[boss][layout]

	// Add a safety offset to avoid moving directly to the boss spawn position
	safetyOffset := data.Position{X: 10, Y: 10}
	safeSpawnPos := data.Position{
		X: spawnPos.X + safetyOffset.X,
		Y: spawnPos.Y + safetyOffset.Y,
	}

	d.ctx.Logger.Debug(fmt.Sprintf("Moving to safe position near boss spawn: X: %d, Y: %d - %sLayout %d", safeSpawnPos.X, safeSpawnPos.Y, boss, layout))
	action.MoveToCoords(safeSpawnPos)

	// Clear the area around the safe position before approaching the actual spawn
	action.ClearAreaAroundPlayer(20, d.getMonsterFilter())

	// Now move to the actual spawn position if it's safe
	if d.isPositionSafe(spawnPos) {
		d.ctx.Logger.Debug(fmt.Sprintf("Moving to actual boss spawn: X: %d, Y: %d", spawnPos.X, spawnPos.Y))
		action.MoveToCoords(spawnPos)
	} else {
		d.ctx.Logger.Debug("Staying at safe position, actual spawn is not safe")
	}
}

func (d *Diablo) isPositionSafe(pos data.Position) bool {
	for _, m := range d.ctx.Data.Monsters.Enemies() {
		if pather.DistanceFromPoint(pos, m.Position) < 10 {
			return false
		}
	}
	return true
}

func (d *Diablo) activateSeal(seal object.Name) error {
	obj, found := d.ctx.Data.Objects.FindOne(seal)
	if !found {
		return fmt.Errorf("seal %v not found", seal)
	}

	if seal == object.DiabloSeal3 && obj.Position.X == 7773 && obj.Position.Y == 5155 {
		if err := action.MoveToCoords(data.Position{X: 7768, Y: 5160}); err != nil {
			return fmt.Errorf("failed to move to bugged seal position: %w", err)
		}
	}

	return action.InteractObject(obj, func() bool {
		updatedObj, found := d.ctx.Data.Objects.FindOne(seal)
		return found && !updatedObj.Selectable
	})
}

func (d *Diablo) killSealElite() error {
	d.ctx.Logger.Debug("Waiting for and killing seal elite")
	startTime := time.Now()

	monsterFilter := d.getMonsterFilter()

	for time.Since(startTime) < 5*time.Second {
		for _, m := range monsterFilter(d.ctx.Data.Monsters.Enemies()) {
			if action.IsMonsterSealElite(m) {
				d.ctx.Logger.Debug("Seal defender found!", slog.Int("monsterId", int(m.Name)))

				action.ClearAreaAroundPlayer(20, monsterFilter)

				return d.ctx.Char.KillMonsterSequence(func(dat game.Data) (data.UnitID, bool) {
					filteredMonsters := monsterFilter(dat.Monsters.Enemies())
					for _, monster := range filteredMonsters {
						if action.IsMonsterSealElite(monster) {
							d.ctx.Logger.Debug("Attempting to kill seal elite", slog.Int("monsterId", int(monster.Name)), slog.Any("position", monster.Position))
							return monster.UnitID, true
						}
					}
					d.ctx.Logger.Debug("Seal elite has been killed or is not found")
					return 0, false
				}, nil)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	d.ctx.Logger.Debug("No seal elite found within 5 seconds")
	return nil
}
func (d *Diablo) clearPath(pathName string, monsterFilter func(data.Monsters) []data.Monster) error {
	action.Buff()

	path := d.paths[pathName]
	if pathName == "starToVizier" {
		path = d.paths[fmt.Sprintf("starToVizier%s", string('A'+d.vizLayout-1))]
	} else if pathName == "starToSeis" {
		path = d.paths[fmt.Sprintf("starToSeis%s", string('A'+d.seisLayout-1))]
	} else if pathName == "starToInfector" {
		path = d.paths[fmt.Sprintf("starToInfector%s", string('A'+d.infLayout-1))]
	}

	for _, pos := range path {
		if d.ctx.Data.AreaData.Grid.IsWalkable(pos) {
			d.ctx.Logger.Debug("Moving to coords", slog.Any("coords", pos))
			action.MoveToCoords(pos)
		}

		action.ClearAreaAroundPlayer(35, monsterFilter)

		d.cleared = append(d.cleared, pos)
	}

	return d.clearStrays(monsterFilter)
}

func (d *Diablo) clearStrays(monsterFilter data.MonsterFilter) error {
	d.ctx.Logger.Debug("Clearing potential stray monsters")
	oldPos := d.ctx.Data.PlayerUnit.Position

	monsters := monsterFilter(d.ctx.Data.Monsters)

	d.ctx.Logger.Debug(fmt.Sprintf("Stray monsters to clear after filtering: %d", len(monsters)))

	actionPerformed := false
	for _, monster := range monsters {
		for _, clearedPos := range d.cleared {
			if pather.DistanceFromPoint(monster.Position, clearedPos) < 30 {
				action.MoveToCoords(monster.Position)
				action.ClearAreaAroundPlayer(15, monsterFilter)
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

func (d *Diablo) getLessConcurredCornerAroundSeal(sealPosition data.Position) data.Position {
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
				averageDistance += distance
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

func (d *Diablo) getMonsterFilter() func(data.Monsters) []data.Monster {
	return func(monsters data.Monsters) []data.Monster {
		filteredMonsters := monsters
		if d.ctx.CharacterCfg.Game.Diablo.FocusOnElitePacks {
			filteredMonsters = data.MonsterEliteFilter()(filteredMonsters)
		}
		filteredMonsters = d.offGridFilter(filteredMonsters)
		if d.ctx.CharacterCfg.Game.Diablo.SkipStormcasters {
			filteredMonsters = d.skipStormCasterFilter(filteredMonsters)
		}
		return filteredMonsters
	}
}

func (d *Diablo) skipStormCasterFilter(monsters data.Monsters) []data.Monster {
	return slices.DeleteFunc(monsters, func(m data.Monster) bool {
		isStormcaster := m.Name == npc.StormCaster || m.Name == npc.StormCaster2
		isSuperUnique := m.Type == data.MonsterTypeSuperUnique
		if isStormcaster && !isSuperUnique {
			return true
		}
		return false
	})
}

func (d *Diablo) offGridFilter(monsters data.Monsters) []data.Monster {
	return slices.DeleteFunc(monsters, func(m data.Monster) bool {
		isOffGrid := !d.ctx.Data.AreaData.IsInside(m.Position)
		if isOffGrid {
		}
		return isOffGrid
	})
}
