package step

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/hid"
)

type KeySequenceStep struct {
	basicStep
	keysToPress []string
}

func KeySequence(keysToPress ...string) *KeySequenceStep {
	return &KeySequenceStep{
		basicStep:   newBasicStep(),
		keysToPress: keysToPress,
	}
}

func (o *KeySequenceStep) Status(_ data.Data) Status {
	if o.status == StatusCompleted {
		return StatusCompleted
	}

	if len(o.keysToPress) > 0 {
		return o.status
	}
	o.tryTransitionStatus(StatusCompleted)

	return o.status
}

func (o *KeySequenceStep) Run(_ data.Data) error {
	if time.Since(o.lastRun) < time.Millisecond*200 {
		return nil
	}

	var k string
	k, o.keysToPress = o.keysToPress[0], o.keysToPress[1:]
	hid.PressKey(k)

	o.lastRun = time.Now()
	return nil
}
