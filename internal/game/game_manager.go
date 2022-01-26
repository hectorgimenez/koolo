package game

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/event"
	"time"
)

type GameManager struct {
	actionChan chan<- action.Action
	eventChan  chan<- event.Event
}

func NewGameManager(actionChan chan<- action.Action, eventChan chan<- event.Event) GameManager {
	return GameManager{
		actionChan: actionChan,
		eventChan:  eventChan,
	}
}

func (gm GameManager) ExitGame() {
	a := action.NewAction(
		action.PriorityHigh,
		action.NewKeyPress("esc", time.Millisecond*500),
		action.NewKeyPress("down", time.Millisecond*50),
		action.NewKeyPress("enter", time.Millisecond*10),
	)
	gm.actionChan <- a
	gm.eventChan <- event.ExitedGame
}

func (gm GameManager) NewGame() {

}
