package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/memory"
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

const MaxDistanceFromLeader = 25

func (f *Follower) Name() string {
	return string(config.FollowerRun)
}

func (f *Follower) Run() error {
	leader, leaderFound := f.ctx.Data.Roster.FindByName(f.ctx.CharacterCfg.Companion.LeaderName)
	if !leaderFound {
		return nil
	}
	f.ctx.Logger.Info("Leader is ", slog.Any("leader", leader))

	for leaderFound {
		// TODO: Detect hostile player and make it a priority. Safety first.
		//hostileFound := false
		//if hostileFound {
		//	f.KillPlayer("")
		//}
		
		if leader.Area != f.ctx.Data.AreaData.Area {
			f.ctx.Logger.Info("Leader is not in the same area. Attempting to get to him.")
			lvl := data.Level{}
			for _, a := range f.ctx.Data.AdjacentLevels {
				if a.Area == leader.Area { // Only pick the first entrance
					lvl = a
					break // Break immediately after finding first valid entrance
				}
			}

			if leader.Area.IsTown() {
				f.ctx.Logger.Info("Leader is in town. Let's go wait there.")
				_ = action.ReturnTown()
				f.goToCorrectTown(leader)
			} else if lvl.Position.X == 0 && lvl.Position.Y == 0 {
				f.ctx.Logger.Info("Leader is not in a connecting area, returning to the correct town to use his portal.")
				_ = action.ReturnTown()
				f.goToCorrectTown(leader)
				err := action.UsePortalFrom(leader.Name)
				if err != nil {
					utils.Sleep(5000)
				}

			} else {
				f.ctx.Logger.Info("Leader is in a connection area. Moving to area ID", slog.Any("area", leader.Area))
				_ = action.MoveToArea(leader.Area)
			}
		} else if leader.Area == f.ctx.Data.AreaData.Area && !f.ctx.Data.PlayerUnit.Area.IsTown() {
			_ = action.ClearAreaAroundPosition(leader.Position, MaxDistanceFromLeader, f.ctx.Data.MonsterFilterAnyReachable())
			_ = action.MoveToCoords(leader.Position)
			action.Buff()

			for _, o := range f.ctx.Data.Objects {
				if o.IsChest() && o.Selectable && f.isChestInRange(o) {
					err := action.MoveToCoords(o.Position)
					if err != nil {
						f.ctx.Logger.Warn("Failed moving to chest: %v", err)
						continue
					}
					err = action.InteractObject(o, func() bool {
						chest, _ := f.ctx.Data.Objects.FindByID(o.ID)
						return !chest.Selectable
					})
					if err != nil {
						f.ctx.Logger.Warn("Failed interacting with chest: %v", err)
					}
					utils.Sleep(500) // Add small delay to allow the game to open the chest and drop the content
				}
			}

		} else if leader.Area == f.ctx.Data.AreaData.Area && f.ctx.Data.PlayerUnit.Area.IsTown() {
			f.ctx.Logger.Info("We followed the Leader to town, let's wait.")
			utils.Sleep(5000)
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

func (f *Follower) goToCorrectTown(leader data.RosterMember) {
	switch act := leader.Area.Act(); act {
	case 1:
		_ = action.WayPoint(area.RogueEncampment)
	case 2:
		_ = action.WayPoint(area.LutGholein)
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

func (f *Follower) isChestInRange(chest data.Object) bool {
	// Calculate the distance between the chest and the player
	distanceX := chest.Position.X - f.ctx.Data.PlayerUnit.Position.X
	distanceY := chest.Position.Y - f.ctx.Data.PlayerUnit.Position.Y
	distance := math.Sqrt(float64(distanceX*distanceX + distanceY*distanceY))

	// Check if the distance is within the allowed range
	return distance <= MaxDistanceFromLeader
}

var isHostile = false

// TODO: Make flag available in D2GO, then implement the following functions
func (f *Follower) isHostile(value bool) {
	isHostile = value
}

func (f *Follower) KillPlayer(playerName string) bool {
	step.SwapToCTA()
	action.Buff()
	step.SwapToMainWeapon()

	for isHostile {
		var target = memory.RawPlayerUnit{}
		for _, p := range f.ctx.GameReader.GetRawPlayerUnits() {
			if p.Name == playerName {
				target = p
			}
		}

		if f.ctx.Data.AreaData.Area == target.Area {
			step.PrimaryAttack(
				target.UnitID,
				4,
				true,
				step.Distance(2, 2), // X,Y coords of 2,2 is the perfect hammer angle attack for NPC targeting/attacking, you can adjust accordingly anything between 1,1 - 3,3 is acceptable, where the higher the number, the bigger the distance from the player (usually used for De Seis)
				step.EnsureAura(skill.Concentration),
			)
		} else {
			var targetAreaNearby = data.Level{}
			for _, a := range f.ctx.Data.AdjacentLevels {
				if a.Area == target.Area { // Only pick the first entrance
					targetAreaNearby = a
					break // Break immediately after finding first valid entrance
				}
			}

			if targetAreaNearby.Position.X != 0 || targetAreaNearby.Position.Y != 0 {
				_ = action.MoveToArea(targetAreaNearby.Area)
			} else {
				_ = action.ClearAreaAroundPlayer(15, f.ctx.Data.MonsterFilterAnyReachable())
				utils.Sleep(200)
			}
		}

		if target.Stats[stat.Life].Value <= 0 {
			break
		}
	}

	return true
}
