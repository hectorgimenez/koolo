package step

import (
	"github.com/hectorgimenez/koolo/internal/game"
)

type MoveToLevelStep struct {
	basicStep
	level                 game.Level
	waitingForInteraction bool
}

func MoveToLevel(level game.Level) *MoveToLevelStep {
	return &MoveToLevelStep{
		basicStep: newBasicStep(),
		level:     level,
	}
}

func (m *MoveToLevelStep) Status(data game.Data) Status {
	//TODO implement me
	panic("implement me")
}

func (m *MoveToLevelStep) Run(data game.Data) error {
	//TODO implement me
	panic("implement me")
}
