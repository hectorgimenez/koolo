package run

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

type Leveling struct {
	ctx *context.Status
}

func NewLeveling() *Leveling {
	return &Leveling{
		ctx: context.Get(),
	}
}

func (a Leveling) Name() string {
	return string(config.LevelingRun)
}

func (a Leveling) Run() error {
	a.act1()
	a.act2()
	a.act3()
	a.act4()
	a.act5()

	return nil
}
