package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	action2 "github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

type ArachnidLair struct {
	ctx *context.Status
}

func NewArachnidLair() *ArachnidLair {
	return &ArachnidLair{
		ctx: context.Get(),
	}
}

func (a ArachnidLair) Name() string {
	return string(config.ArachnidLairRun)
}

func (a ArachnidLair) Run() error {
	err := action2.WayPoint(area.SpiderForest)
	if err != nil {
		return err
	}

	err = action2.MoveToArea(area.SpiderCave)
	if err != nil {
		return err
	}

	action2.OpenTPIfLeader()

	// Clear ArachnidLair
	return action2.ClearCurrentLevel(true, data.MonsterAnyFilter())
}
