package action

import (
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type MouseClick struct {
	delayedOperation
	button hid.MouseButton
}

func (m MouseClick) execute() {
	hid.Click(m.button)
}

func NewMouseClick(button hid.MouseButton, delay time.Duration) MouseClick {
	return MouseClick{
		delayedOperation: delayedOperation{delayAfterOperation: delay},
		button:           button,
	}
}
