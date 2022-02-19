package step

import (
	"github.com/hectorgimenez/koolo/internal/game"
)

type MoveToLevel struct {
	basicStep
	level                 game.Level
	waitingForInteraction bool
}

func NewMoveToLevel(level game.Level) *MoveToLevel {
	return &MoveToLevel{
		basicStep: newBasicStep(),
		level:     level,
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
