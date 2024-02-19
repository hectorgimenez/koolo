package run

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const NameFollower = "Follower"
const MaxCoordinateDiff = 100

var leaderNotFoundInformed = false

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
		action.NewStepChain(func(d data.Data) []step.Step {
			leaderRosterMember, found := d.Roster.FindByName(config.Config.Follower.LeaderName)
			if !found {
				if !leaderNotFoundInformed {
					f.logger.Warn(fmt.Sprintf("Leader not found: %s", config.Config.Companion.LeaderName))
					leaderNotFoundInformed = true
				}

				// When leader has not been found, it is NOT an error situation. Just wait
				return []step.Step{step.Wait(100)}
			}

			// Is leader too far way? If yes, do not do anything
			if pather.DistanceFromMe(d, leaderRosterMember.Position) > MaxCoordinateDiff {
				return []step.Step{step.Wait(100)}
			}

			return []step.Step{step.MoveTo(leaderRosterMember.Position)}

		}, action.RepeatUntilNoSteps()),
	}
}
