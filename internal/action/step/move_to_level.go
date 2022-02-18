package step

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type MoveToLevel struct {
	basicStep
	level                 game.Level
	pf                    helper.PathFinderV2
	waitingForInteraction bool
}

func NewMoveToLevel(level game.Level, pf helper.PathFinderV2) *MoveToLevel {
	return &MoveToLevel{
		basicStep: newBasicStep(),
		level:     level,
		pf:        pf,
	}
}

func (m *MoveToLevel) Status(data game.Data) Status {
	//TODO implement me
	panic("implement me")
}

func (m *MoveToLevel) Run(data game.Data) error {
	//TODO implement me
	panic("implement me")
}
