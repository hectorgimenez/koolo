package run

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

const NameFollower = "Follower"
const MaxCoordinateDiff = 100
const EntranceMaxDiff = 15
const EntranceMaxSecondsDelay = time.Second * 5

var leaderNotFoundInformed = false
var lastEntranceEntered = time.Now()

type Follower struct {
	baseRun
}

func (f Follower) Name() string {
	return NameFollower
}

func (f Follower) BuildActions() []action.Action {
	// The proof of concept implementation of just following the leader for now
	// More actions to be added later

	return []action.Action{
		action.NewChain(func(d data.Data) []action.Action {
			leaderRosterMember, found := d.Roster.FindByName(config.Config.Follower.LeaderName)
			if !found {
				if !leaderNotFoundInformed {
					f.logger.Warn(fmt.Sprintf("Leader not found: %s", config.Config.Companion.LeaderName))
					leaderNotFoundInformed = true
				}

				// When leader has not been found, it is NOT an error situation. Just wait
				return []action.Action{
					f.builder.Wait(300),
				}
			}

			// Is leader too far away?
			if pather.DistanceFromMe(d, leaderRosterMember.Position) > MaxCoordinateDiff {
				// If we have an entrance, use it
				entrance := getClosestEntrances(d)
				if entrance != nil && time.Since(lastEntranceEntered) > EntranceMaxSecondsDelay {
					lastEntranceEntered = time.Now()

					return []action.Action{
						f.builder.MoveToArea(entrance.Area),
					}
				}

				// Is leader in the same act and in a waypoint location? Let's use waypoint to that location
				_, wpAreaFound := area.WPAddresses[leaderRosterMember.Area]
				if d.PlayerUnit.Area.Act() == leaderRosterMember.Area.Act() && d.PlayerUnit.Area.IsTown() && wpAreaFound {
					return []action.Action{
						f.builder.WayPoint(leaderRosterMember.Area),
						f.builder.Wait(time.Second * 1),
					}
				}

				return []action.Action{
					f.builder.Wait(100),
				}
			}

			return []action.Action{
				action.NewStepChain(func(d data.Data) []step.Step {
					return []step.Step{step.MoveTo(leaderRosterMember.Position)}
				}),
				f.builder.Wait(100),
			}
		}, action.RepeatUntilNoSteps()),
	}
}

func getClosestEntrances(d data.Data) *data.Level {
	for _, l := range d.AdjacentLevels {
		distFromMe := pather.DistanceFromMe(d, l.Position)
		if distFromMe <= EntranceMaxDiff {
			return &l
		}
	}

	return nil
}
