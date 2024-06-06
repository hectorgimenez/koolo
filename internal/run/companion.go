package run

import (
	"context"
	"strings"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/lxn/win"
)

type Companion struct {
	baseRun
}

func (s Companion) Name() string {
	return "companion"
}

func (s Companion) BuildActions() []action.Action {
	var lastInteractionEvent *event.InteractedToEvent
	var leaderUnitIDTarget data.UnitID
	tpRequested := false
	var portalUsedToGoCity data.UnitID

	// TODO: Deregister this listener or will leak
	s.EventListener.Register(func(ctx context.Context, e event.Event) error {
		if strings.EqualFold(config.Characters[e.Supervisor()].CharacterName, s.CharacterCfg.Companion.LeaderName) {
			if evt, ok := e.(event.CompanionLeaderAttackEvent); ok {
				leaderUnitIDTarget = evt.TargetUnitID
			}

			if evt, ok := e.(event.InteractedToEvent); ok {
				lastInteractionEvent = &evt
			}
		}

		return nil
	})

	return []action.Action{
		action.NewChain(func(d game.Data) []action.Action {
			leaderRosterMember, _ := d.Roster.FindByName(s.CharacterCfg.Companion.LeaderName)

			// Leader is NOT in the same act, so we will try to change to the corresponding act
			if leaderRosterMember.Area != area.None && leaderRosterMember.Area.Act() != d.PlayerUnit.Area.Act() {
				// Follower is NOT in town
				if !d.PlayerUnit.Area.IsTown() {

					// Portal is found nearby
					if _, foundPortal := getClosestPortal(d, leaderRosterMember.Name); foundPortal {
						return []action.Action{
							s.builder.UsePortalFrom(leaderRosterMember.Name),
						}
					}

					// Portal is not found nearby
					if hasEnoughPortals(d) {
						return []action.Action{
							s.builder.ReturnTown(),
						}
					}

					// there is NO portal open and follower does NOT have enough portals. Just exit
					return []action.Action{}
				}

				// Follower is in town. Just change the act
				return []action.Action{
					s.builder.WayPoint(town.GetTownByArea(leaderRosterMember.Area).TownArea()),
				}
			}

			if lastInteractionEvent != nil {
				switch lastInteractionEvent.InteractionType {
				case event.InteractionTypeEntrance:
					a := area.ID(lastInteractionEvent.ID)
					lastInteractionEvent = nil
					if !d.PlayerUnit.Area.IsTown() {
						return []action.Action{
							s.builder.MoveToArea(a),
						}
					}
				case event.InteractionTypeObject:
					oName := object.Name(lastInteractionEvent.ID)
					lastInteractionEvent = nil
					o, found := d.Objects.FindOne(oName)
					if found && ((o.IsWaypoint() && !d.PlayerUnit.Area.IsTown()) || o.IsRedPortal()) {
						return []action.Action{
							s.builder.InteractObject(oName, func(dat game.Data) bool {
								if o.IsWaypoint() {
									return dat.OpenMenus.Waypoint
								}

								return d.PlayerUnit.Area != dat.PlayerUnit.Area
							}),
						}
					}
				case event.InteractionTypeNPC:
					npcID := npc.ID(lastInteractionEvent.ID)
					lastInteractionEvent = nil
					switch npcID {
					case npc.Warriv, npc.Meshif:
						return []action.Action{
							s.builder.ReturnTown(),
							s.builder.InteractNPC(npcID, step.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)),
						}
					}
				}
			}

			if leaderRosterMember.Area.IsTown() && !d.PlayerUnit.Area.IsTown() && s.CharacterCfg.Companion.FollowLeader {
				return []action.Action{
					s.builder.ReturnTown(),
					action.NewStepChain(func(d game.Data) []step.Step {
						return []step.Step{
							step.SyncStep(func(g game.Data) error {
								time.Sleep(time.Second * 2)
								portal, found := getClosestPortal(d, leaderRosterMember.Name)
								if found {
									portalUsedToGoCity = portal.ID
								}
								return nil
							}),
						}
					}),
				}
			}

			// Is leader too far away?
			if pather.DistanceFromMe(d, leaderRosterMember.Position) > 100 {
				// In some cases this "follower in town -> use portal -> follower outside town -> use portal"
				// loop can go on forever. But it is responsibility of a leader to not cause it...

				// Follower in town
				if d.PlayerUnit.Area.IsTown() {
					// Request a TP
					if pather.DistanceFromMe(d, town.GetTownByArea(d.PlayerUnit.Area).TPWaitingArea(d)) < 10 && !tpRequested {
						event.Send(event.CompanionRequestedTP(event.Text(s.Supervisor, "TP Requested")))
						tpRequested = true
						return []action.Action{
							s.builder.Wait(time.Second),
						}
					}

					if p, foundPortal := getClosestPortal(d, leaderRosterMember.Name); foundPortal && !leaderRosterMember.Area.IsTown() {
						if p.ID != portalUsedToGoCity {
							tpRequested = false
							return []action.Action{
								s.builder.UsePortalFrom(leaderRosterMember.Name),
							}
						}
					}

					// Go to TP waiting area
					return []action.Action{
						s.builder.MoveToCoords(town.GetTownByArea(d.PlayerUnit.Area).TPWaitingArea(d)),
					}
				}

				return []action.Action{
					s.builder.Wait(100),
				}
			}

			if !s.CharacterCfg.Companion.FollowLeader {
				return []action.Action{
					s.builder.Wait(100),
				}
			}

			// If distance from leader is acceptable and is attacking, support him
			distanceFromMe := pather.DistanceFromMe(d, leaderRosterMember.Position)
			if distanceFromMe < 30 {
				monster, found := d.Monsters.FindByID(leaderUnitIDTarget)
				if s.CharacterCfg.Companion.Attack && found {
					return []action.Action{s.killMonsterInCompanionMode(monster)}
				}

				// If there is no monster to attack, and we are close enough to the leader
				if distanceFromMe < 4 {
					// If we're not leveling AND we have at least some monsters nearby, let's kill them
					for _, m := range d.Monsters.Enemies() {
						if d := pather.DistanceFromMe(d, m.Position); d <= 8 {
							return []action.Action{s.killMonsterInCompanionMode(m)}
						}
					}

					return []action.Action{
						s.builder.ItemPickup(false, 8),
						s.builder.Wait(100),
					}
				}
			}

			// If follower is in town and we are NOT leveling, let's NOT follow the leader
			_, isLevelingChar := s.char.(action.LevelingCharacter)
			if !isLevelingChar && d.PlayerUnit.Area.IsTown() {
				return []action.Action{
					s.builder.MoveToCoords(town.GetTownByArea(d.PlayerUnit.Area).TPWaitingArea(d)),
				}
			}

			return []action.Action{
				action.NewStepChain(func(d game.Data) []step.Step {
					return []step.Step{step.MoveTo(leaderRosterMember.Position, step.WithTimeout(time.Millisecond*500), step.StopAtDistance(20))}
				}),
			}
		}, action.RepeatUntilNoSteps()),
	}
}

func getClosestPortal(d game.Data, leaderName string) (*data.Object, bool) {
	for _, o := range d.Objects {
		if o.IsPortal() && pather.DistanceFromMe(d, o.Position) <= 40 && strings.EqualFold(o.Owner, leaderName) {
			return &o, true
		}
	}

	return nil, false
}

func hasEnoughPortals(d game.Data) bool {
	portalTome, pFound := d.Inventory.Find(item.TomeOfTownPortal, item.LocationInventory)
	if pFound {
		st, found := portalTome.FindStat(stat.Quantity, 0)
		if found && st.Value > 0 {
			return true
		}
	}

	return false
}

func (s Companion) killMonsterInCompanionMode(m data.Monster) action.Action {
	switch m.Name {
	case npc.Andariel:
		return s.char.KillAndariel()
	case npc.Duriel:
		return s.char.KillDuriel()
	case npc.Mephisto:
		return s.char.KillMephisto()
	case npc.Diablo:
		return s.char.KillDiablo()
	}

	return s.char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		return m.UnitID, true
	}, nil)
}
