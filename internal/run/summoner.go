package run

import (
	"math"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"
)

var minChestDistanceLaneEnd = 5
var maxChestDistanceLaneEnd = 25

type Summoner struct {
	ctx *context.Status
}

var clearRange int = 30

var ArcSequencer = []int{
	1, 2, 6, 3, 4, 5, 1, 0,
}

func NewSummoner() *Summoner {
	return &Summoner{
		ctx: context.Get(),
	}
}

func (s Summoner) Name() string {
	return string(config.SummonerRun)
}

func (s Summoner) Run() error {
	// Set/Reset the checkpoint position data for the Bot to move to.
	var ArcCheckPointsList = []data.Position{
		{X: 25448, Y: 5448}, /*Center Point 0*/
		/*Base Lane Coordinates*/
		{ /*Start 1*/ X: 25544, Y: 5446}, { /*Center on Right Lane-a 2*/ X: 25637, Y: 5383}, { /*center on Right Lane-b 3*/ X: 25754, Y: 5384},
		{ /*End Point 4*/ X: 25853, Y: 5448}, { /*Center on Left Lane 5*/ X: 25637, Y: 5506},
		{ /*Center of Lane 6*/ X: 25683, Y: 5453},
	}
	// Use the waypoint to get to Arcane Sanctuary
	err := action.WayPoint(area.ArcaneSanctuary)
	if err != nil {
		return err
	}

	// Get the Summoner's position from the cached map data
	areaData := s.ctx.Data.Areas[area.ArcaneSanctuary]
	summonerNPC, found := areaData.NPCs.FindOne(npc.Summoner)
	if !found || len(summonerNPC.Positions) == 0 {
		return err
	}

	// Do the rounds looking for ghosts.
	if s.ctx.CharacterCfg.Game.Summoner.ClearGhosts || s.ctx.CharacterCfg.Game.Summoner.ClearArea {
		for i := 0; i <= 3; i++ {
			for _, i := range ArcSequencer {
				action.ClearThroughPath(ArcCheckPointsList[i], clearRange, s.getMonsterFilter())
				// Check for summoner
				if ArcCheckPointsList[i] == ArcCheckPointsList[4] &&
					summonerNPC.Positions[0].X-s.ctx.Data.PlayerUnit.Position.X < 20 &&
					summonerNPC.Positions[0].Y-s.ctx.Data.PlayerUnit.Position.Y < 20 {
					s.ctx.Char.KillSummoner()
				}
				// Locate Chests
				if ArcCheckPointsList[i] == ArcCheckPointsList[4] {
					for _, o := range s.ctx.Data.Objects {
						if o.Selectable && isChestWithinLaneEndRange(o, ArcCheckPointsList[4]) {
							err = action.MoveToCoords(o.Position)
							if err != nil {
								s.ctx.Logger.Warn("Failed moving to chest: %v", err)
								continue
							}
							err = action.InteractObject(o, func() bool {
								chest, _ := s.ctx.Data.Objects.FindByID(o.ID)
								return !chest.Selectable
							})
							if err != nil {
								s.ctx.Logger.Warn("Failed interacting with chest: %v", err)
							}
							utils.Sleep(500) // Add small delay to allow the game to open the chest and drop the content
						}
					}
				}
			}
			// Rotate points from initial lane to the lane counter clockwise from it.
			for d := 1; d < len(ArcCheckPointsList); d++ {
				ArcCheckPointsList[d] = rotatePoint(float64(ArcCheckPointsList[d].X), float64(ArcCheckPointsList[d].Y), float64(ArcCheckPointsList[0].X), float64(ArcCheckPointsList[0].Y), 90)
			}
		}
	} else {
		//Move to the Summoner's position using the static coordinates from map data
		if err = action.MoveToCoords(summonerNPC.Positions[0]); err != nil {
			return err
		}

		// Kill Summoner
		s.ctx.Char.KillSummoner()
	}

	return action.WayPoint(area.RogueEncampment)
}

func (s *Summoner) getMonsterFilter() data.MonsterFilter {
	return func(monsters data.Monsters) (filteredMonsters []data.Monster) {
		for _, m := range monsters {
			if !s.ctx.Data.AreaData.IsWalkable(m.Position) {
				continue
			}

			// If ClearGhosts is enabled, only return ghosts
			if s.ctx.CharacterCfg.Game.Summoner.ClearGhosts && !s.ctx.CharacterCfg.Game.Summoner.ClearArea {
				if m.Name == 40 {
					filteredMonsters = append(filteredMonsters, m)
				}
			} else {
				filteredMonsters = append(filteredMonsters, m)
			}
		}

		return filteredMonsters
	}
}

func isChestWithinLaneEndRange(chest data.Object, laneEnd data.Position) bool {
	distance := pather.DistanceFromPoint(chest.Position, laneEnd)
	return distance >= minChestDistanceLaneEnd && distance <= maxChestDistanceLaneEnd
}

func rotatePoint(x, y, centerX, centerY, angle float64) data.Position {
	// Translate to origin
	x -= centerX
	y -= centerY
	// Rotation calculation using radians
	radAngle := math.Pi * angle / 180
	newX := x*math.Cos(radAngle) - y*math.Sin(radAngle)
	newY := x*math.Sin(radAngle) + y*math.Cos(radAngle)
	var finalX int = int(math.Ceil(newX))
	var finalY int = int(math.Ceil(newY))
	var finalPosition = data.Position{X: finalX + int(centerX), Y: finalY + int(centerY)}
	// Translate back to original center
	return finalPosition
}
