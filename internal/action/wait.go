package action

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

func (b Builder) Wait(duration time.Duration) *StaticAction {
	return BuildStatic(func(d data.Data) (steps []step.Step) {
		return []step.Step{
			step.Wait(duration),
		}
	})
}
