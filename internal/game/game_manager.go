package game

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
)

type GameManager struct {
	cfg        config.Config
	actionChan chan<- action.Action
	eventChan  chan<- event.Event
}

func NewGameManager(cfg config.Config, actionChan chan<- action.Action, eventChan chan<- event.Event) GameManager {
	return GameManager{
		cfg:        cfg,
		actionChan: actionChan,
		eventChan:  eventChan,
	}
}
