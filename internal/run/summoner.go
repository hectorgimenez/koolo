package run

import (
	"math"

	//"slices"

	"strconv"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"

	//"github.com/hectorgimenez/d2go/pkg/data/object"
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

var ArcCheckPointsList = []data.Position{
	{X: 25448, Y: 5448}, /*Center Point 0*/
	/*Base Lane Coordinates*/
	{ /*Start 1*/ X: 25544, Y: 5446}, { /*Center on Right Lane-a 2*/ X: 25637, Y: 5383}, { /*center on Right Lane-b 3*/ X: 25754, Y: 5384},
	{ /*End Point 4*/ X: 25853, Y: 5448}, { /*Center on Left Lane 5*/ X: 25637, Y: 5506},
	{ /*Center of Lane 6*/ X: 25683, Y: 5453},
}

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
	//sctx := context.Get()
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
				istrg := strconv.Itoa(i)
				s.ctx.Logger.Debug("Heading to point" + istrg)
				action.ClearThroughPath(ArcCheckPointsList[i], clearRange, s.getMonsterFilter())
				if ArcCheckPointsList[i] == ArcCheckPointsList[4] &&
					summonerNPC.Positions[0].X-s.ctx.Data.PlayerUnit.Position.X < 20 &&
					summonerNPC.Positions[0].Y-s.ctx.Data.PlayerUnit.Position.Y < 20 {
					s.ctx.Char.KillSummoner()
				}
				if ArcCheckPointsList[i] == ArcCheckPointsList[4] {
					for _, o := range s.ctx.Data.Objects {
						if /*(o.IsChest() || o.IsSuperChest()) &&*/ o.Selectable && isChestWithinLaneEndRange(o, ArcCheckPointsList[4]) {
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
					/*interactableObjects := []object.Name{object.ArcaneLargeChestLeft, object.ArcaneLargeChestRight,
					object.ArcaneSmallChestLeft, object.ArcaneSmallChestRight, object.Act2LargeChestLeft,
					object.Act2LargeChestRight, object.Act2MediumChestRight}*/

					// Find the interactable objects
					/*var objects []data.Object
					for _, o := range s.ctx.Data.Objects {
						if o.IsChest() || o.IsSuperChest() {
							objects = append(objects, o)
							s.ctx.Logger.Debug("We added a chest to the list.")
						}

						// Interact with objects in the order of shortest travel
						for len(objects) > 0 {

							//playerPos := s.ctx.Data.PlayerUnit.Position

							sort.Slice(objects, func(x, j int) bool {
								return pather.DistanceFromPoint(objects[x].Position, ArcCheckPointsList[4]) <
									pather.DistanceFromPoint(objects[j].Position, ArcCheckPointsList[4])
							})

							// Interact with the closest object
							closestObject := objects[0]
							if !isChestWithinLaneEndRange(o, ArcCheckPointsList[4]) {
								objects = objects[1:]
								s.ctx.Logger.Debug("We removed a chest because it was too far.")
							} else {
								err = action.InteractObject(closestObject, func() bool {
									object, _ := s.ctx.Data.Objects.FindByID(closestObject.ID)
									s.ctx.Logger.Debug("We popped a chest")
									return !object.Selectable
								})
							}
							if err != nil {
								s.ctx.Logger.Warn(fmt.Sprintf("[%s] failed interacting with object [%v] in Area: [%s]", s.ctx.Name, closestObject.Name, s.ctx.Data.PlayerUnit.Area.Area().Name), err)
							}
							utils.Sleep(500) // Add small delay to allow the game to open the object and drop the content

							// Remove the interacted container from the list
							objects = objects[1:]
							s.ctx.Logger.Debug("We removed a chest after interacting with it.")
						}
					}*/
				}
			}
			s.ctx.Logger.Debug("We've completed a lane loop.")
			for d := 1; d < len(ArcCheckPointsList); d++ {
				dstrgX := strconv.Itoa(ArcCheckPointsList[d].X)
				dstrgy := strconv.Itoa(ArcCheckPointsList[d].Y)
				s.ctx.Logger.Debug("The Old point was: X: " + dstrgX + ", Y: " + dstrgy)
				ArcCheckPointsList[d] = rotatePoint(float64(ArcCheckPointsList[d].X), float64(ArcCheckPointsList[d].Y), float64(ArcCheckPointsList[0].X), float64(ArcCheckPointsList[0].Y), 90)
				s.ctx.Logger.Debug("We've rotated a checkpoint")
				dstrgX = strconv.Itoa(ArcCheckPointsList[d].X)
				dstrgy = strconv.Itoa(ArcCheckPointsList[d].Y)
				s.ctx.Logger.Debug("The New point is: X: " + dstrgX + ", Y: " + dstrgy)
			}
		}
		s.ctx.Logger.Debug("We've completed a lane loop.")
	} else {
		s.ctx.Logger.Debug("We've left the ghost killing loop.")
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
