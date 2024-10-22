package run

import (
	"errors"
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"log/slog"
	"slices"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

var diabloSpawnPosition = data.Position{X: 7792, Y: 5294}
var chaosSanctuaryEntrancePosition = data.Position{X: 7790, Y: 5544}
var ErrBossNotFound = errors.New("seal elite not found")

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
	d.initPaths()

	// Clear the path from River of Flame to Chaos Sanctuary  if can't Teleport
	if !d.ctx.CharacterCfg.Character.UseTeleport {
		if err := d.clearPath("riverToChaos", ""); err != nil {
			return err
		}
	}

	targetPosition := diabloSpawnPosition
	if !d.ctx.CharacterCfg.Game.Diablo.StartFromStar {
		targetPosition = chaosSanctuaryEntrancePosition
	}

	if err := action.MoveToCoords(targetPosition); err != nil {
		return err
	}

	d.initLayout()

	if d.ctx.CharacterCfg.Companion.Leader {
		action.OpenTPIfLeader()
		action.Buff()
		action.ClearAreaAroundPlayer(30, d.getMonsterFilter("")) // Use empty string for general clearing
	}

	// Clear the path from entrance to star if not starting from star
	if !d.ctx.CharacterCfg.Game.Diablo.StartFromStar {
		if err := d.clearPath("entranceToStar", ""); err != nil {
			return err
		}
	}

	bosses := []string{"Vizier", "Seis", "Infector"}

	for _, boss := range bosses {
		pathName := fmt.Sprintf("starTo%s", boss)
		if err := d.clearPath(pathName, boss); err != nil {
			return err
		}

		if err := d.killBoss(boss); err != nil {
			return err
		}
	}

	if d.ctx.CharacterCfg.Game.Diablo.KillDiablo {
		action.Buff()

		safePos := action.FindNearestWalkablePosition(diabloSpawnPosition)
		action.MoveToCoords(safePos)

		// Check if we should disable item pickup for Diablo
		if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
			context.Get().DisableItemPickup()
		}
		// Re-enable item pickup if it was disabled
		if err := d.ctx.Char.KillDiablo(); err != nil {
			if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
				context.Get().EnableItemPickup()
			}
			return err
		}

		if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
			context.Get().EnableItemPickup()
		}
	}

	return nil
}

func (d *Diablo) initLayout() {
	d.vizLayout = d.getLayout(object.DiabloSeal4, 5275)
	d.seisLayout = d.getLayout(object.DiabloSeal3, 7773)
	d.infLayout = d.getLayout(object.DiabloSeal1, 7893)
	d.ctx.Logger.Debug(fmt.Sprintf("Layouts initialized - Vizier: %d, Seis: %d, Infector: %d", d.vizLayout, d.seisLayout, d.infLayout))
}

func (d *Diablo) initPaths() {
	d.paths["riverToChaos"] = []data.Position{{X: 7792, Y: 5925}, {X: 7790, Y: 5890}, {X: 7793, Y: 5853}, {X: 7794, Y: 5819}, {X: 7795, Y: 5782}, {X: 7794, Y: 5744}, {X: 7795, Y: 5711}, {X: 7791, Y: 5672}, {X: 7792, Y: 5632}, {X: 7795, Y: 5601}, {X: 7794, Y: 5563}, {X: 7790, Y: 5544}}
	d.paths["entranceToStar"] = []data.Position{{X: 7794, Y: 5517}, {X: 7791, Y: 5491}, {X: 7768, Y: 5459}, {X: 7775, Y: 5424}, {X: 7817, Y: 5458}, {X: 7777, Y: 5408}, {X: 7769, Y: 5379}, {X: 7777, Y: 5357}, {X: 7809, Y: 5359}, {X: 7805, Y: 5330}, {X: 7780, Y: 5317}, {X: 7791, Y: 5293}}
	d.paths["starToVizierA"] = []data.Position{{X: 7759, Y: 5295}, {X: 7734, Y: 5295}, {X: 7716, Y: 5295}, {X: 7718, Y: 5276}, {X: 7697, Y: 5292}, {X: 7678, Y: 5293}, {X: 7665, Y: 5276}, {X: 7662, Y: 5314}}
	d.paths["starToVizierB"] = []data.Position{{X: 7759, Y: 5295}, {X: 7734, Y: 5295}, {X: 7716, Y: 5295}, {X: 7701, Y: 5315}, {X: 7666, Y: 5313}, {X: 7653, Y: 5284}}
	d.paths["starToSeisA"] = []data.Position{{X: 7781, Y: 5259}, {X: 7805, Y: 5258}, {X: 7802, Y: 5237}, {X: 7776, Y: 5228}, {X: 7775, Y: 5205}, {X: 7804, Y: 5193}, {X: 7814, Y: 5169}, {X: 7788, Y: 5153}}
	d.paths["starToSeisB"] = []data.Position{{X: 7781, Y: 5259}, {X: 7805, Y: 5258}, {X: 7802, Y: 5237}, {X: 7776, Y: 5228}, {X: 7811, Y: 5218}, {X: 7807, Y: 5194}, {X: 7779, Y: 5193}, {X: 7774, Y: 5160}, {X: 7803, Y: 5154}}
	d.paths["starToInfectorA"] = []data.Position{{X: 7809, Y: 5268}, {X: 7834, Y: 5306}, {X: 7852, Y: 5280}, {X: 7852, Y: 5310}, {X: 7869, Y: 5294}, {X: 7895, Y: 5295}, {X: 7919, Y: 5290}}
	d.paths["starToInfectorB"] = []data.Position{{X: 7809, Y: 5268}, {X: 7834, Y: 5306}, {X: 7852, Y: 5280}, {X: 7852, Y: 5310}, {X: 7869, Y: 5294}, {X: 7895, Y: 5274}, {X: 7927, Y: 5275}, {X: 7932, Y: 5297}, {X: 7923, Y: 5313}}
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

func (d *Diablo) killBoss(boss string) error {
	d.ctx.Logger.Debug(fmt.Sprintf("Starting boss sequence for %s", boss))

	if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
		context.Get().DisableItemPickup()
	}
	defer func() {
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

		if err := d.clearAndActivateSeal(sealName, d.getMonsterFilter(boss)); err != nil {
			return err
		}

		// For Infector, kill the boss after the first seal
		if boss == "Infector" && i == 0 {
			if err := d.killSealElite(boss); err != nil {
				return err
			}
		}
	}

	// For Vizier and Seis, kill the boss after all seals are activated
	if boss != "Infector" {
		if boss == "Seis" {
			if err := d.moveToDeSeisSpawn(); err != nil {
				return err
			}
		}
		if err := d.killSealElite(boss); err != nil {
			if errors.Is(err, ErrBossNotFound) {
				d.ctx.Logger.Debug(fmt.Sprintf("%s was already killed during seal clearing", boss))
				return nil
			}
			return err
		}
	}

	return nil
}

func (d *Diablo) clearAndActivateSeal(sealName object.Name, _ func(data.Monsters) []data.Monster) error {
	seal, found := d.ctx.Data.Objects.FindOne(sealName)
	if !found {
		return fmt.Errorf("seal %v not found", sealName)
	}

	d.ctx.Logger.Debug(fmt.Sprintf("Moving to seal %v", sealName))
	if err := action.MoveToCoords(seal.Position); err != nil {
		d.ctx.Logger.Error(fmt.Sprintf("Failed to move to seal %v: %v", sealName, err))
		return err
	}

	// Handle the special case for DiabloSeal3
	if sealName == object.DiabloSeal3 && seal.Position.X == 7773 && seal.Position.Y == 5155 {
		if err := action.MoveToCoords(data.Position{X: 7768, Y: 5160}); err != nil {
			return fmt.Errorf("failed to move to bugged seal position: %w", err)
		}
	}

	// Clear monsters immediately around the seal without moving away
	d.ctx.Logger.Debug("Clearing monsters around the seal")
	d.clearImmediateArea(seal.Position, 10, d.sealActivationFilter())

	d.ctx.Logger.Debug(fmt.Sprintf("Activating seal %v", sealName))
	return action.InteractObject(seal, func() bool {
		updatedObj, found := d.ctx.Data.Objects.FindOne(sealName)
		return found && !updatedObj.Selectable
	})
}

func (d *Diablo) clearImmediateArea(center data.Position, radius int, monsterFilter func(data.Monsters) []data.Monster) {
	monsters := monsterFilter(d.ctx.Data.Monsters.Enemies())
	for _, monster := range monsters {
		if d.ctx.PathFinder.DistanceFromMe(monster.Position) <= radius {
			_ = d.ctx.Char.KillMonsterSequence(func(dat game.Data) (data.UnitID, bool) {
				return monster.UnitID, true
			}, nil)
		}
	}
}
func (d *Diablo) killSealElite(boss string) error {
	d.ctx.Logger.Debug(fmt.Sprintf("Starting kill sequence for %s", boss))
	startTime := time.Now()
	timeout := 10 * time.Second

	monsterFilter := d.getMonsterFilter(boss)

	for time.Since(startTime) < timeout {
		monsters := monsterFilter(d.ctx.Data.Monsters.Enemies())
		for _, m := range monsters {
			if action.IsMonsterSealElite(m) {
				d.ctx.Logger.Debug(fmt.Sprintf("Seal elite found: %s at position X: %d, Y: %d", m.Name, m.Position.X, m.Position.Y))

				safeDistance := d.ctx.CharacterCfg.Game.Diablo.AttackFromDistance

				err := d.ctx.Char.KillMonsterSequence(func(dat game.Data) (data.UnitID, bool) {
					monster, found := dat.Monsters.FindByID(m.UnitID)
					if !found || monster.Stats[stat.Life] <= 0 {
						return 0, false
					}
					currentDist := d.ctx.PathFinder.DistanceFromMe(monster.Position)
					if currentDist < safeDistance-5 || currentDist > safeDistance+5 {
						var newSafePos data.Position
						if currentDist < safeDistance {
							newSafePos = action.GetSafePositionAwayFromMonster(d.ctx.Data.PlayerUnit.Position, monster.Position, safeDistance)
						} else {
							newSafePos = action.GetSafePositionTowardsMonster(d.ctx.Data.PlayerUnit.Position, monster.Position, safeDistance)
						}
						_ = action.MoveToCoords(newSafePos)
					}
					return monster.UnitID, true
				}, nil)

				if err != nil {
					d.ctx.Logger.Warn(fmt.Sprintf("Failed to kill seal elite: %v", err))
					return err
				} else {
					d.ctx.Logger.Debug("Successfully killed seal elite")
					return nil
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	d.ctx.Logger.Warn(fmt.Sprintf("No seal elite found for %s within %v seconds", boss, timeout.Seconds()))
	return ErrBossNotFound
}

func (d *Diablo) moveToDeSeisSpawn() error {
	spawnPositions := map[int][]data.Position{
		1: {{X: 7789, Y: 5202}, {X: 7785, Y: 5193}, {X: 7775, Y: 5183}},
		2: {{X: 7795, Y: 5155}, {X: 7785, Y: 5145}, {X: 7775, Y: 5135}, {X: 7765, Y: 5125}},
	}

	positions := spawnPositions[d.seisLayout]
	d.ctx.Logger.Debug(fmt.Sprintf("Attempting to move near De Seis spawn for layout %d", d.seisLayout))

	for _, pos := range positions {
		d.ctx.Logger.Debug(fmt.Sprintf("Trying position X: %d, Y: %d", pos.X, pos.Y))

		walkablePos := action.FindNearestWalkablePosition(pos)
		if !d.isSafePositionForSeis(walkablePos) {
			d.ctx.Logger.Debug(fmt.Sprintf("Skipping unsafe position X: %d, Y: %d", walkablePos.X, walkablePos.Y))
			continue
		}

		d.ctx.Logger.Debug(fmt.Sprintf("Moving to safe position X: %d, Y: %d", walkablePos.X, walkablePos.Y))
		if err := action.MoveToCoords(walkablePos); err != nil {
			d.ctx.Logger.Warn(fmt.Sprintf("Failed to move to position: %v", err))
			continue
		}

		currentPos := d.ctx.Data.PlayerUnit.Position
		d.ctx.Logger.Debug(fmt.Sprintf("Moved to position X: %d, Y: %d", currentPos.X, currentPos.Y))

		// Check if De Seis is already dead
		if !d.isSealEliteAlive("Seis") {
			d.ctx.Logger.Debug("De Seis was already killed during seal clearing")
			return nil
		}

		// Clear the area
		action.ClearAreaAroundPlayer(15, d.getMonsterFilter("Seis"))

		// Check if we're in an acceptable position
		if d.ctx.PathFinder.DistanceFromMe(pos) <= 20 && d.isSafePositionForSeis(currentPos) {
			d.ctx.Logger.Debug("Successfully positioned for De Seis encounter")
			return nil
		}
	}

	return errors.New("failed to move to an acceptable position for De Seis")
}
func (d *Diablo) isSafePositionForSeis(pos data.Position) bool {
	if d.seisLayout != 1 {
		return true
	}
	safeX, safeY := 7789, 5202
	return pos.X <= safeX && pos.Y <= safeY
}

func (d *Diablo) clearPath(pathName string, boss string) error {
	action.Buff()

	path := d.paths[pathName]
	if pathName == "starToVizier" {
		path = d.paths[fmt.Sprintf("starToVizier%s", string('A'+d.vizLayout-1))]
	} else if pathName == "starToSeis" {
		path = d.paths[fmt.Sprintf("starToSeis%s", string('A'+d.seisLayout-1))]
	} else if pathName == "starToInfector" {
		path = d.paths[fmt.Sprintf("starToInfector%s", string('A'+d.infLayout-1))]
	}

	monsterFilter := d.getMonsterFilter(boss)

	for _, pos := range path {
		walkablePos := action.FindNearestWalkablePosition(pos)
		d.ctx.Logger.Debug("Moving to coords", slog.Any("original", pos), slog.Any("walkable", walkablePos))
		if err := action.MoveToCoords(walkablePos); err != nil {
			d.ctx.Logger.Error("Failed to move to coords", slog.Any("pos", walkablePos), slog.String("error", err.Error()))
			return err
		}

		// Clear the area without moving back
		monsters := monsterFilter(d.ctx.Data.Monsters.Enemies())
		for _, monster := range monsters {
			if d.ctx.PathFinder.DistanceFromMe(monster.Position) <= 35 {
				_ = d.ctx.Char.KillMonsterSequence(func(dat game.Data) (data.UnitID, bool) {
					return monster.UnitID, true
				}, nil)
			}
		}

		d.cleared = append(d.cleared, walkablePos)
	}

	return nil
}

func (d *Diablo) getMonsterFilter(boss string) func(data.Monsters) []data.Monster {
	return func(monsters data.Monsters) []data.Monster {
		// First, filter out off-grid monsters
		filteredMonsters := d.offGridFilter(monsters, boss)

		// If FocusOnElitePacks is enabled, only return elite monsters and seal bosses
		if d.ctx.CharacterCfg.Game.Diablo.FocusOnElitePacks {
			return slices.DeleteFunc(filteredMonsters, func(m data.Monster) bool {
				return !m.IsElite() && !action.IsMonsterSealElite(m)
			})
		}

		// If FocusOnElitePacks is not enabled, return all filtered monsters
		return filteredMonsters
	}
}
func (d *Diablo) sealActivationFilter() func(data.Monsters) []data.Monster {
	return func(monsters data.Monsters) []data.Monster {
		return slices.DeleteFunc(monsters, func(m data.Monster) bool {
			return !d.ctx.Data.AreaData.IsInside(m.Position)
		})
	}
}

func (d *Diablo) offGridFilter(monsters data.Monsters, boss string) []data.Monster {
	return slices.DeleteFunc(monsters, func(m data.Monster) bool {
		isOffGrid := !d.ctx.Data.AreaData.IsInside(m.Position)

		// Special case for Vizier: don't filter him out even if he's off-grid
		if boss == "Vizier" && action.IsMonsterSealElite(m) {
			return false
		}

		if isOffGrid {
			d.ctx.Logger.Debug("Skipping off-grid monster", slog.Any("monster", m.Name), slog.Any("position", m.Position))
		}
		return isOffGrid
	})
}
func (d *Diablo) isSealEliteAlive(boss string) bool {
	monsters := d.getMonsterFilter(boss)(d.ctx.Data.Monsters.Enemies())
	for _, m := range monsters {
		if action.IsMonsterSealElite(m) {
			return true
		}
	}
	return false
}

// TODO make this better it doesnt always work  .for walkable characters
func (d *Diablo) clearPathToChaos() error {
	d.ctx.Logger.Debug("Clearing path from River of Flame to Chaos Sanctuary")

	pathToChaos := d.paths["riverToChaos"]

	for _, pos := range pathToChaos {
		err := action.MoveToCoords(pos)
		if err != nil {
			d.ctx.Logger.Debug("Movement failed, checking player state", slog.Any("position", pos))

			// Check player mode and clear if necessary
			switch d.ctx.Data.PlayerUnit.Mode {
			case mode.GettingHit, mode.Blocking, mode.KnockedBack:
				d.ctx.Logger.Debug("Player under attack, clearing area", slog.Any("mode", d.ctx.Data.PlayerUnit.Mode))
				if clearErr := action.ClearAreaAroundPlayer(7, d.getMonsterFilter("")); clearErr != nil {
					d.ctx.Logger.Warn("Failed to clear area", slog.String("error", clearErr.Error()))
				}
			default:
			}

			// Retry moving after potential clearing
			if retryErr := action.MoveToCoords(pos); retryErr != nil {
				d.ctx.Logger.Error("Failed to move after checking state", slog.Any("position", pos), slog.String("error", retryErr.Error()))
				return retryErr
			}
		}
	}

	d.ctx.Logger.Debug("Path to Chaos Sanctuary cleared")
	return nil
}
