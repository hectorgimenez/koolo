package action

import (
	"github.com/hectorgimenez/koolo/internal/hid"
	"math/rand"
	"time"
)

const (
	PriorityHigh   = "high"
	PriorityNormal = "normal"
)

type Priority string

type Action struct {
	Priority Priority
	sequence []HIDOperation
}

func NewAction(priority Priority, sequence ...HIDOperation) Action {
	return Action{
		Priority: priority,
		sequence: sequence,
	}
}

func (a Action) run() {
	for _, op := range a.sequence {
		op.execute()
		time.Sleep(op.delay())
	}
}

type HIDOperation interface {
	execute()
	delay() time.Duration
}

type delayedOperation struct {
	delayAfterOperation time.Duration
}

// Add random 1-30% extra delay to all operations
func (d delayedOperation) delay() time.Duration {
	rand.Seed(time.Now().UnixNano())
	randomPct := ((float32(rand.Intn(30-1+1)) + 1) / 100) + 1
	delayedMs := int64(float32(d.delayAfterOperation.Milliseconds()) * randomPct)

	return time.Duration(delayedMs) * time.Millisecond
}

type KeyPress struct {
	delayedOperation
	keyToPress   string
	combinedKeys []string
}

func NewKeyPress(key string, delay time.Duration, combinedKeys ...string) KeyPress {
	return KeyPress{
		delayedOperation: delayedOperation{
			delayAfterOperation: delay,
		},
		keyToPress:   key,
		combinedKeys: combinedKeys,
	}
}

func (k KeyPress) execute() {
	if len(k.combinedKeys) > 0 {
		hid.PressKeyCombination(k.keyToPress, k.combinedKeys...)
	} else {
		hid.PressKey(k.keyToPress)
	}
}

type MouseDisplacement struct {
	delayedOperation
	x int
	y int
}

func (m MouseDisplacement) execute() {
	hid.MovePointer(m.x, m.y)
}

func NewMouseDisplacement(delay time.Duration, x, y int) MouseDisplacement {
	return MouseDisplacement{
		delayedOperation: delayedOperation{delayAfterOperation: delay},
		x:                x,
		y:                y,
	}
}

type MouseClick struct {
	delayedOperation
	button hid.MouseButton
}

func (m MouseClick) execute() {
	hid.Click(m.button)
}

func NewMouseClick(delay time.Duration, button hid.MouseButton) MouseClick {
	return MouseClick{
		delayedOperation: delayedOperation{delayAfterOperation: delay},
		button:           button,
	}
}
