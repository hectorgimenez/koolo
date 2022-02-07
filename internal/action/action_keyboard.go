package action

import (
	"github.com/go-vgo/robotgo"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

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

type KeyToggle struct {
	delayedOperation
	keyToPress string
	press      bool
}

func NewKeyDown(key string, delay time.Duration) KeyToggle {
	return KeyToggle{
		delayedOperation: delayedOperation{
			delayAfterOperation: delay,
		},
		keyToPress: key,
		press:      true,
	}
}

func NewKeyUp(key string, delay time.Duration) KeyToggle {
	return KeyToggle{
		delayedOperation: delayedOperation{
			delayAfterOperation: delay,
		},
		keyToPress: key,
		press:      false,
	}
}

func (k KeyToggle) execute() {
	if k.press {
		robotgo.KeyDown(k.keyToPress)
	} else {
		robotgo.KeyUp(k.keyToPress)
	}
}
