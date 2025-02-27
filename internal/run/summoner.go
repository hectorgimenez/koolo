package run

import (
	"math"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

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
	var finalPosition = data.Position{X: finalX, Y: finalY}
	// Translate back to original center
	return finalPosition
}

var ArcCheckPointsList = []data.Position{
	{X: 25448, Y: 5448}, /*Center Point 0*/
	/*East Lane Coordinates*/
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
	for i := 0; i < 3; i++ {
		for _, l := range ArcSequencer {
			action.ClearTargetsThroughPath(ArcCheckPointsList[l], clearRange, 40)
		}
		for d := 1; d < len(ArcCheckPointsList); d++ {
			ArcCheckPointsList[d] = rotatePoint(float64(ArcCheckPointsList[d].X), float64(ArcCheckPointsList[d].Y), float64(ArcCheckPointsList[0].X), float64(ArcCheckPointsList[0].Y), 45)
		}
	}

	// Move to the Summoner's position using the static coordinates from map data
	if err = action.MoveToCoords(summonerNPC.Positions[0]); err != nil {
		return err
	}

	// Kill Summoner
	s.ctx.Char.KillSummoner()
	hasTouchedBook := false
	book, _ := s.ctx.Data.Objects.FindOne(object.YetAnotherTome)
	action.MoveToCoords(book.Position)
	if !hasTouchedBook {
		action.InteractObject(book, func() bool {
			return hasTouchedBook
		})
	}
	time.Sleep(100 * time.Millisecond)
	townPortal, found := s.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
	if !found {
		s.ctx.Logger.Debug("No portal found.")
		return action.WayPoint(area.RogueEncampment)
	}

	action.InteractObject(townPortal, func() bool {
		return s.ctx.Data.AreaData.Area == area.CanyonOfTheMagi
	})

	return action.WayPoint(area.RogueEncampment)

}
