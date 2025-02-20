package run

import (
	"errors"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
	"log/slog"
	"math"
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
	leader, leaderFound := f.ctx.Data.Roster.FindByName(f.ctx.CharacterCfg.Companion.LeaderName)
	if !leaderFound {
		return nil
	}
	f.ctx.Logger.Info("Leader is ", slog.Any("leader", leader))
	f.goToCorrectTown(leader)
	for leaderFound {
		if leader.Area != f.ctx.Data.AreaData.Area {
			f.handleLeaderNotInSameArea(leader)
		} else if leader.Area == f.ctx.Data.AreaData.Area && !f.ctx.Data.PlayerUnit.Area.IsTown() {
			f.handleLeaderInSameArea(leader)
		} else if leader.Area == f.ctx.Data.AreaData.Area && f.ctx.Data.PlayerUnit.Area.IsTown() {
			f.handleLeaderInTown()
		}

		f.ctx.RefreshGameData()
		utils.Sleep(200)
		leader, leaderFound = f.ctx.Data.Roster.FindByName(f.ctx.CharacterCfg.Companion.LeaderName)
		if !leaderFound {
			f.ctx.Logger.Info("Leader is gone, leaving game.")
		}
	}

	return nil
}

func (f *Follower) handleLeaderNotInSameArea(leader data.RosterMember) {
	if leader.Area.IsTown() {
		f.ctx.Logger.Info("Leader is in town. Let's go wait there.")
		_ = action.ReturnTown()
		action.VendorRefill(false, true)
		f.goToCorrectTown(leader)
		utils.Sleep(200)
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

func (f *Follower) getEntrancesToLeader(leader data.RosterMember) []data.Level {
	var entrances []data.Level
	for _, al := range f.ctx.Data.AdjacentLevels {
		if al.Area == leader.Area { // Only pick the first entrance
			entrances = append(entrances, al)
		}
	}
	return entrances
}

func (f *Follower) findClosestEntrance(entrances []data.Level) *data.Level {
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

func (f *Follower) handleNoValidEntrance(leader data.RosterMember) {
	f.ctx.Logger.Info("Leader is not in a connecting area, returning to the correct town to use his portal.")
	_ = action.ReturnTown()
	action.VendorRefill(false, true)
	f.goToCorrectTown(leader)

	err := f.UseCorrectPortalFromLeader(leader)
	if err != nil {
		utils.Sleep(5000)
	}
}

func (f *Follower) handleLeaderInSameArea(leader data.RosterMember) {
	_ = action.ClearAreaAroundPosition(leader.Position, MaxDistanceFromLeader, f.ctx.Data.MonsterFilterAnyReachable())
	_ = action.MoveToCoords(leader.Position)
	action.Buff()
	f.interactWithChests()
}

func (f *Follower) interactWithChests() {
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
	utils.Sleep(5000)
}

func (f *Follower) goToCorrectTown(leader data.RosterMember) {
	switch act := leader.Area.Act(); act {
	case 1:
		_ = action.WayPoint(area.RogueEncampment)
		_ = action.MoveTo(f.getKashyaPosition)
	case 2:
		_ = action.WayPoint(area.LutGholein)
		_ = action.MoveTo(f.getAtmaPosition)
	case 3:
		_ = action.WayPoint(area.KurastDocks)
	case 4:
		_ = action.WayPoint(area.ThePandemoniumFortress)
	case 5:
		_ = action.WayPoint(area.Harrogath)
	default:
		f.ctx.Logger.Info("Could not find the Leader's current Act location.")
	}
}

func (f *Follower) getAtmaPosition() (data.Position, bool) {
	atma, found := f.ctx.Data.NPCs.FindOne(npc.Atma)
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

func (f *Follower) UseCorrectPortalFromLeader(leader data.RosterMember) error {
	ctx := context.Get()
	ctx.SetLastAction("UsePortalFrom")

	if !ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}

	for _, obj := range ctx.Data.Objects {
		if obj.IsPortal() && obj.Owner == leader.Name && obj.PortalData.DestArea == leader.Area {

			return action.InteractObjectByID(obj.ID, nil)
		}
	}

	return errors.New("Waiting for corrected portal")
}
