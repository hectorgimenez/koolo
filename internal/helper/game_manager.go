package helper

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

func ExitGame(actionChan chan<- action.Action, eventChan chan<- event.Event) {
	a := action.NewAction(
		action.PriorityHigh,
		action.NewKeyPress("esc", time.Millisecond*200),
		action.NewMouseDisplacement(640, 328, time.Millisecond*50),
		action.NewMouseClick(hid.LeftButton, time.Millisecond*120),
	)
	actionChan <- a
	eventChan <- event.ExitedGame
}

func NewGame(actionChan chan<- action.Action, difficulty string) {
	difficultyPosition := map[string]struct {
		X, Y int
	}{
		"normal":    {X: 640, Y: 311},
		"nightmare": {X: 640, Y: 355},
		"hell":      {X: 640, Y: 403},
	}

	createX := difficultyPosition[difficulty].X
	createY := difficultyPosition[difficulty].Y
	a := action.NewAction(
		action.PriorityNormal,
		action.NewMouseDisplacement(640, 672, time.Millisecond*50),
		action.NewMouseClick(hid.LeftButton, time.Millisecond*350),
		action.NewMouseDisplacement(createX, createY, time.Millisecond*87),
		action.NewMouseClick(hid.LeftButton, time.Millisecond*65),
	)
	actionChan <- a
}
