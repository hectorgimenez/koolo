package run

import (
	"fmt"
	"log/slog"
	"math"
	"slices"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
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
	ctx        *context.Status
	vizLayout  int
	seisLayout int
	infLayout  int
	cleared    []data.Position
	paths      map[string][]data.Position
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
	if d.ctx.CharacterCfg.Game.Diablo.StartFromStar == false {
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

	if d.ctx.CharacterCfg.Game.Diablo.StartFromStar == false {
		if err := d.clearPath("entranceToStar", d.getMonsterFilter()); err != nil {
			return err
		}
	}

	for _, boss := range []string{"Vizier", "Seis", "Infector"} {
		if d.ctx.CharacterCfg.Game.Diablo.StartFromStar == false {
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
			// Re-enable item pickup if it was disabled
			if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
				context.Get().EnableItemPickup()
			}
			return err
		}

		// Re-enable item pickup if it was disabled
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
	d.ctx.Logger.Debug(fmt.Sprintf("Starting boss sequence for %s", boss))

	// Disable item pickup for boss seals if configured
	if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
		context.Get().DisableItemPickup()
	}
	defer func() {
		// Re-enable item pickup after boss seal is dead
		if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
			context.Get().EnableItemPickup()
		}
	}()

	sealNames := map[string][]object.Name{
		"Vizier":   {object.DiabloSeal4, object.DiabloSeal5},
		"Seis":     {object.DiabloSeal3},
		"Infector": {object.DiabloSeal1, object.DiabloSeal2},
	}[boss]

	for i, sealName := range sealNames {
		d.ctx.Logger.Debug(fmt.Sprintf("Processing seal %v for %s", sealName, boss))

		if err := d.clearAndActivateSeal(sealName); err != nil {
			return err
		}

		// For Infector, kill the boss after the first seal
		if boss == "Infector" && i == 0 {
			if err := d.moveToBossSpawn(boss); err != nil {
				return err
			}
			time.Sleep(1500 * time.Millisecond)
			if err := d.killSealElite(); err != nil {
				return err
			}
		}
	}

	// For Vizier and Seis, kill the boss after all seals are activated
	if boss != "Infector" {
		if err := d.moveToBossSpawn(boss); err != nil {
			return err
		}
		time.Sleep(1500 * time.Millisecond)

		if !d.isBossVisibleAndInRange(boss, 10) {
			d.ctx.Logger.Debug(fmt.Sprintf("%s not visible, moving closer", boss))
			if err := d.moveToExactBossSpawn(boss); err != nil {
				return err
			}
		}

		if err := d.killSealElite(); err != nil {
			return err
		}
	}

	return nil
}

func (d *Diablo) clearAndActivateSeal(sealName object.Name) error {
	seal, found := d.ctx.Data.Objects.FindOne(sealName)
	if !found {
		return fmt.Errorf("seal %v not found", sealName)
	}

	d.ctx.Logger.Debug(fmt.Sprintf("Moving to seal %v", sealName))
	if err := action.MoveToCoords(seal.Position); err != nil {
		d.ctx.Logger.Error(fmt.Sprintf("Failed to move to seal %v: %v", sealName, err))
		return err
	}

	d.ctx.Logger.Debug("Clearing monsters around the seal")
	action.ClearAreaAroundPlayer(10, func(monsters data.Monsters) []data.Monster {
		return slices.DeleteFunc(monsters, func(m data.Monster) bool {
			return !d.ctx.Data.AreaData.IsInside(m.Position)
		})
	})

	d.ctx.Logger.Debug(fmt.Sprintf("Activating seal %v", sealName))
	return d.activateSeal(sealName)
}

func (d *Diablo) moveToExactBossSpawn(boss string) error {
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

	return action.MoveToCoords(spawnPos)
}

func (d *Diablo) moveToBossSpawn(boss string) error {
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
	d.ctx.Logger.Debug(fmt.Sprintf("Moving towards %s spawn at X: %d, Y: %d - Layout %d", boss, spawnPos.X, spawnPos.Y, layout))

	// Define a safe distance (8 units is about 16 yards, which should be a good balance)
	safeDistance := 8

	// Calculate a safe position
	safePos := d.getSafePosition(spawnPos, safeDistance)

	// Move to the safe position
	if err := action.MoveToCoords(safePos); err != nil {
		d.ctx.Logger.Error(fmt.Sprintf("Failed to move to safe position for %s: %v", boss, err))
		return err
	}

	// Clear the area around the player
	action.ClearAreaAroundPlayer(safeDistance+2, d.getMonsterFilter())

	return nil
}

func (d *Diablo) getSafePosition(target data.Position, safeDistance int) data.Position {
	playerPos := d.ctx.Data.PlayerUnit.Position
	dx := float64(target.X - playerPos.X)
	dy := float64(target.Y - playerPos.Y)
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance <= float64(safeDistance) {
		return playerPos // Already at a safe distance
	}

	ratio := float64(distance-float64(safeDistance)) / distance
	return data.Position{
		X: playerPos.X + int(dx*ratio),
		Y: playerPos.Y + int(dy*ratio),
	}
}

func (d *Diablo) isBossVisibleAndInRange(boss string, maxRange int) bool {
	for _, m := range d.ctx.Data.Monsters.Enemies(data.MonsterEliteFilter()) {
		if action.IsMonsterSealElite(m) && d.ctx.PathFinder.DistanceFromMe(m.Position) <= maxRange {
			d.ctx.Logger.Debug(fmt.Sprintf("%s found at distance %d", boss, d.ctx.PathFinder.DistanceFromMe(m.Position)))
			return true
		}
	}
	return false
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

	// Clear a larger area around the seal
	action.ClearAreaAroundPlayer(10, d.sealActivationFilter())

	// Move closer to the seal if not already near it
	if d.ctx.PathFinder.DistanceFromMe(obj.Position) > 5 {
		if err := action.MoveToCoords(obj.Position); err != nil {
			return fmt.Errorf("failed to move to seal position: %w", err)
		}
	}

	return action.InteractObject(obj, func() bool {
		updatedObj, found := d.ctx.Data.Objects.FindOne(seal)
		return found && !updatedObj.Selectable
	})
}

func (d *Diablo) sealActivationFilter() func(data.Monsters) []data.Monster {
	return func(monsters data.Monsters) []data.Monster {
		return slices.DeleteFunc(monsters, func(m data.Monster) bool {
			return !d.ctx.Data.AreaData.IsInside(m.Position)
		})
	}
}

func (d *Diablo) killSealElite() error {
	d.ctx.Logger.Debug("Waiting for and killing seal elite")
	startTime := time.Now()

	for time.Since(startTime) < 5*time.Second {
		for _, m := range d.getMonsterFilter()(d.ctx.Data.Monsters.Enemies(data.MonsterEliteFilter())) {
			if action.IsMonsterSealElite(m) {
				d.ctx.Logger.Debug("Seal defender found!")
				action.ClearAreaAroundPlayer(20, func(monsters data.Monsters) []data.Monster {
					return slices.DeleteFunc(d.getMonsterFilter()(monsters), func(monster data.Monster) bool {
						return !action.IsMonsterSealElite(monster)
					})
				})

				return d.ctx.Char.KillMonsterSequence(func(dat game.Data) (data.UnitID, bool) {
					for _, monster := range d.getMonsterFilter()(dat.Monsters.Enemies(data.MonsterEliteFilter())) {
						if action.IsMonsterSealElite(monster) {
							_, _, found := d.ctx.PathFinder.GetPath(monster.Position)
							if found {
								d.ctx.Logger.Debug(fmt.Sprintf("Attempting to kill seal elite: %v", monster.Name))
								return monster.UnitID, true
							}
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
		walkablePos := d.findNearestWalkablePosition(pos)
		d.ctx.Logger.Debug("Moving to coords", slog.Any("original", pos), slog.Any("walkable", walkablePos))
		if err := action.MoveToCoords(walkablePos); err != nil {
			d.ctx.Logger.Error("Failed to move to coords", slog.Any("pos", walkablePos), slog.String("error", err.Error()))
			return err
		}

		action.ClearAreaAroundPlayer(35, d.getMonsterFilter())

		d.cleared = append(d.cleared, walkablePos)
	}

	return d.clearStrays(d.getMonsterFilter())
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

func (d *Diablo) getMonsterFilter() func(data.Monsters) []data.Monster {
	return func(monsters data.Monsters) []data.Monster {
		filteredMonsters := monsters
		if d.ctx.CharacterCfg.Game.Diablo.FocusOnElitePacks {
			filteredMonsters = data.MonsterEliteFilter()(filteredMonsters)
		}
		filteredMonsters = d.offGridFilter(filteredMonsters)
		return filteredMonsters
	}
}

func (d *Diablo) offGridFilter(monsters data.Monsters) []data.Monster {
	return slices.DeleteFunc(monsters, func(m data.Monster) bool {
		isOffGrid := !d.ctx.Data.AreaData.IsInside(m.Position)
		if isOffGrid {
			d.ctx.Logger.Debug("Skipping off-grid monster", slog.Any("monster", m.Name), slog.Any("position", m.Position))
		}
		return isOffGrid
	})
}

func (d *Diablo) findNearestWalkablePosition(pos data.Position) data.Position {
	if d.ctx.Data.AreaData.Grid.IsWalkable(pos) {
		return pos
	}

	for radius := 1; radius <= 10; radius++ {
		for x := pos.X - radius; x <= pos.X+radius; x++ {
			for y := pos.Y - radius; y <= pos.Y+radius; y++ {
				checkPos := data.Position{X: x, Y: y}
				if d.ctx.Data.AreaData.Grid.IsWalkable(checkPos) {
					return checkPos
				}
			}
		}
	}

	// If no walkable position found, return the original position
	return pos
}
