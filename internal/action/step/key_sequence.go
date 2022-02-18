package step

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type KeySequence struct {
	basicStep
	keysToPress []string
}

func NewKeySequence(keysToPress ...string) *KeySequence {
	return &KeySequence{
		basicStep:   newBasicStep(),
		keysToPress: keysToPress,
	}
}

func (o *KeySequence) Status(_ game.Data) Status {
	if len(o.keysToPress) > 0 {
		return o.status
	}
	o.tryTransitionStatus(StatusCompleted)

	return o.status
}

func (o *KeySequence) Run(_ game.Data) error {
	if time.Since(o.lastRun) < time.Millisecond*200 {
		return nil
	}

	var k string
	k, o.keysToPress = o.keysToPress[0], o.keysToPress[1:]
	hid.PressKey(k)

	o.lastRun = time.Now()
	return nil
}
