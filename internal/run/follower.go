package run

import (
	"errors"
	"log/slog"
	"math"
	"math/rand/v2"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/utils"
)

type Follower struct {
	ctx *context.Status
}

func NewFollower() *Follower {
	return &Follower{
		ctx: context.Get(),
	}
}

const MaxDistanceFromLeader = 30

func (f *Follower) Name() string {
	return string(config.FollowerRun)
}

func (f *Follower) Run() error {
	f.ctx.RefreshGameData()
	f.ctx.SetLastAction("Finding Leader")
	leader, leaderFound := f.ctx.Data.Roster.FindByName(f.ctx.CharacterCfg.Companion.LeaderName)
	if !leaderFound {
		f.ctx.Logger.Error("Leader not found.")
		// Re-enable this when companion PR gets merged if they don't implement a reset
		//f.resetCompanionGameInfo()
		return nil
	}

	f.ctx.Logger.Info("Leader is ", slog.Any("leader", leader))

	//Anti-stuck solution
	lastPosition := data.Position{}

	for leaderFound {
		f.ctx.RefreshGameData()
		utils.Sleep(200)
		leader, leaderFound = f.ctx.Data.Roster.FindByName(f.ctx.CharacterCfg.Companion.LeaderName)
		f.ctx.Logger.Info("Leader is still here.", slog.String("leader", leader.Name))
		if !leaderFound {
			f.ctx.Logger.Info("Leader is gone, leaving game.")
			// Re-enable this when companion PR gets merged if they don't implement a reset
			//f.resetCompanionGameInfo()
			return nil
		}

		if leader.Area.Area().Name == "" {
			continue
		}

		if leader.Area != f.ctx.Data.AreaData.Area {
			f.handleLeaderNotInSameArea(&leader)
		} else if leader.Area == f.ctx.Data.AreaData.Area && !f.ctx.Data.PlayerUnit.Area.IsTown() {
			lastPosition = f.ctx.Data.PlayerUnit.Position
			f.handleLeaderInSameArea(&leader)
			f.ctx.RefreshGameData()
			if f.ctx.Data.PlayerUnit.Position == lastPosition {
				f.ctx.PathFinder.RandomTeleport()
				utils.Sleep(2000)
			}
		} else if leader.Area == f.ctx.Data.AreaData.Area && f.ctx.Data.PlayerUnit.Area.IsTown() {
			f.handleLeaderInTown()
		}
	}

	return nil
}

func (f *Follower) handleLeaderNotInSameArea(leader *data.RosterMember) {
	f.ctx.SetLastAction("Handle leader not in same area")
	if leader.Area.IsTown() {
		f.ctx.Logger.Info("Leader is in town. Let's go wait there.")
		if !f.ctx.Data.AreaData.Area.IsTown() {
			_ = action.ReturnTown()
		}
		_ = f.InTownRoutine()
		_ = f.goToCorrectTown(leader)
		return
	}

	f.ctx.Logger.Info("Leader is not in the same area. Attempting to get to him.")
	entrances := f.getEntrancesToLeader(leader)
	lvl := f.findClosestEntrance(entrances)

	if lvl == nil || (lvl.Position.X == 0 && lvl.Position.Y == 0) {
		f.handleNoValidEntrance(leader)
	} else {
		f.ctx.Logger.Info("Leader is in a connection area. Moving to area ID", slog.Any("area", lvl.Area))
		_ = action.MoveToCoords(lvl.Position)

		if lvl.IsEntrance {
			_ = step.InteractEntrance(lvl.Area)
		}
		_ = action.MoveToCoords(leader.Position)
	}
}

func (f *Follower) getEntrancesToLeader(leader *data.RosterMember) []data.Level {
	f.ctx.SetLastAction("Identifying best entrance to leader")
	var entrances []data.Level
	for _, al := range f.ctx.Data.AdjacentLevels {
		if al.Area == leader.Area { // Only pick the first entrance
			entrances = append(entrances, al)
		}
	}

	return entrances
}

func (f *Follower) findClosestEntrance(entrances []data.Level) *data.Level {
	f.ctx.SetLastAction("Finding closest entrance")
	var lvl *data.Level
	minDistance := math.MaxFloat64 // Start with the largest possible distance

	for _, e := range entrances {
		dist := calculateDistance(f.ctx.Data.PlayerUnit.Position, e.Position)
		if dist < minDistance {
			minDistance = dist
			lvl = &e
		}
	}
	return lvl
}

func (f *Follower) handleNoValidEntrance(leader *data.RosterMember) error {
	if leader.Area == area.TheWorldstoneChamber && f.ctx.Data.AreaData.Area == area.ThroneOfDestruction {
		return f.handleBaalScenario()
	}

	f.ctx.Logger.Info("Leader is not in a connecting area, returning to the correct town to use his portal.")
	f.ctx.SetLastAction("Handle No Valid Entrance")
	if !f.ctx.Data.AreaData.Area.IsTown() {
		_ = action.ReturnTown()
	}
	_ = f.InTownRoutine()
	_ = f.goToCorrectTown(leader)

	err := f.UseCorrectPortalFromLeader(leader)
	if err != nil {
		utils.Sleep(5000)
	}

	return nil
}

func (f *Follower) handleLeaderInSameArea(leader *data.RosterMember) {
	f.ctx.SetLastAction("Handle leader in same area")
	_ = action.ClearAreaAroundPosition(leader.Position, MaxDistanceFromLeader, f.ctx.Data.MonsterFilterAnyReachable())
	_ = action.MoveToCoords(leader.Position)
	action.BuffIfRequired()
	f.interactWithChests()
}

func (f *Follower) interactWithChests() {
	f.ctx.SetLastAction("Interact with chests")
	for _, o := range f.ctx.Data.Objects {
		if o.IsChest() && o.Selectable && f.isChestInRange(o) {
			action.ClearAreaAroundPosition(o.Position, 10, data.MonsterAnyFilter())
			if err := f.moveToChestAndInteract(o); err != nil {
				f.ctx.Logger.Warn("Failed interacting with chest: %v", err)
			}
		}
	}
}

func (f *Follower) moveToChestAndInteract(o data.Object) error {
	if err := action.MoveToCoords(o.Position); err != nil {
		return err
	}
	return action.InteractObject(o, func() bool {
		chest, _ := f.ctx.Data.Objects.FindByID(o.ID)
		return !chest.Selectable
	})
}

func (f *Follower) handleLeaderInTown() {
	f.ctx.Logger.Info("We followed the Leader to town, let's wait.")
	f.ctx.SetLastAction("Handle Leader In Town")
	utils.Sleep(5000)
}

func (f *Follower) goToCorrectTown(leader *data.RosterMember) error {
	f.ctx.Logger.Info("Going to the correct town.", slog.Int("leader act", leader.Area.Act()), slog.String("leader area", leader.Area.Area().Name))
	f.ctx.SetLastAction("Going to the correct town")
	switch act := leader.Area.Act(); act {
	case 1:
		targetPos, _ := f.getKashyaPosition()
		_ = action.WayPoint(area.RogueEncampment)
		_ = action.MoveToCoords(data.Position{X: targetPos.X + randRange(-5, 5), Y: targetPos.Y + randRange(-5, 5)})
	case 2:
		targetPos, _ := f.getAtmaPosition()
		_ = action.WayPoint(area.LutGholein)
		_ = action.MoveToCoords(data.Position{X: targetPos.X + randRange(-5, 5), Y: targetPos.Y + randRange(5, 15)})
	case 3:
		targetPos, _ := f.getOrmusPosition()
		_ = action.WayPoint(area.KurastDocks)
		_ = action.MoveToCoords(data.Position{X: targetPos.X + randRange(-5, 5), Y: targetPos.Y + randRange(5, 15)})
	case 4:
		_ = action.WayPoint(area.ThePandemoniumFortress)
		f.ctx.PathFinder.RandomMovement()
	case 5:
		_ = action.WayPoint(area.Harrogath)
		f.ctx.PathFinder.RandomMovement()
	default:
		f.ctx.Logger.Error("Could not find the Leader's current Act location")
		return errors.New("Could not find the Leader's current Act location")
	}
	f.ctx.Logger.Info("Went to the correct town.")
	return nil
}

func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

func calculateDistance(pos1, pos2 data.Position) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return math.Sqrt(float64(dx*dx + dy*dy))
}

func (f *Follower) isChestInRange(chest data.Object) bool {
	distanceX := chest.Position.X - f.ctx.Data.PlayerUnit.Position.X
	distanceY := chest.Position.Y - f.ctx.Data.PlayerUnit.Position.Y
	distance := math.Sqrt(float64(distanceX*distanceX + distanceY*distanceY))
	return distance <= MaxDistanceFromLeader
}

func (f *Follower) UseCorrectPortalFromLeader(leader *data.RosterMember) error {
	f.ctx.SetLastAction("UsePortalFromLeader")

	if !f.ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}

	for _, obj := range f.ctx.Data.Objects {
		if obj.IsPortal() && obj.Owner == leader.Name && obj.PortalData.DestArea == leader.Area {
			action.Buff()
			return action.InteractObjectByID(obj.ID, nil)
		}
	}

	return errors.New("Waiting for correct portal...")
}

func (f *Follower) InTownRoutine() error {
	f.ctx.SetLastAction("In Town Routine")
	_ = action.Stash(false)
	_ = action.IdentifyAll(false)
	_ = action.VendorRefill(false, true)
	_ = action.Stash(false)
	action.ReviveMerc()
	_ = action.CubeRecipes()
	_ = f.goToTpArea()
	return nil
}

func (f *Follower) getAtmaPosition() (data.Position, bool) {
	atma, found := f.ctx.Data.NPCs.FindOne(npc.Atma)
	if found {
		return atma.Positions[0], true
	}
	return data.Position{}, false
}

func (f *Follower) getOrmusPosition() (data.Position, bool) {
	atma, found := f.ctx.Data.NPCs.FindOne(npc.Ormus)
	if found {
		return atma.Positions[0], true
	}
	return data.Position{}, false
}

func (f *Follower) getKashyaPosition() (data.Position, bool) {
	cain, found := f.ctx.Data.NPCs.FindOne(npc.Kashya)
	if found {
		return cain.Positions[0], true
	}
	return data.Position{}, false
}

func (f *Follower) handleBaalScenario() error {
	f.ctx.Logger.Info("Leader is in The Worldstone Chamber, going through the red portal.")
	baalPortal, _ := f.ctx.Data.Objects.FindOne(object.BaalsPortal)
	_ = action.InteractObject(baalPortal, nil)
	utils.Sleep(700)
	if err := f.ctx.Char.KillBaal(); err != nil {
		return action.ClearCurrentLevel(false, data.MonsterAnyFilter())
	}

	return nil
}

func (f *Follower) goToTpArea() error {
	tpArea := town.GetTownByArea(f.ctx.Data.PlayerUnit.Area).TPWaitingArea(*f.ctx.Data)
	return action.MoveToCoords(tpArea)
}

// Re-enable this when companion PR gets merged if they don't implement a reset
// func (f *Follower) resetCompanionGameInfo() {
// 	f.ctx.Context.CharacterCfg.Companion.CompanionGameName = ""
// 	f.ctx.Context.CharacterCfg.Companion.CompanionGamePassword = ""
// }
