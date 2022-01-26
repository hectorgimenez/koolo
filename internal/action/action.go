package action

import (
	"github.com/hectorgimenez/koolo/internal/hid"
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

type KeyPress struct {
	delayAfterOperation time.Duration
	keyToPress          string
	combinedKeys        []string
}

func NewKeyPress(key string, delay time.Duration, combinedKeys ...string) KeyPress {
	return KeyPress{
		delayAfterOperation: delay,
		keyToPress:          key,
		combinedKeys:        combinedKeys,
	}
}

func (k KeyPress) execute() {
	if len(k.combinedKeys) > 0 {
		hid.PressKeyCombination(k.keyToPress, k.combinedKeys...)
	} else {
		hid.PressKey(k.keyToPress)
	}
}

func (k KeyPress) delay() time.Duration {
	return k.delayAfterOperation
}
