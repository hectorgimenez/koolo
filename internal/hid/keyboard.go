package hid

import (
	"github.com/go-vgo/robotgo"
	"math/rand"
	"time"
)

const keyPressMinTime = 30  // ms
const keyPressMaxTime = 160 // ms

// PressKey toggles a key, it holds the key between keyPressMinTime and keyPressMaxTime ms randomly
func PressKey(key string) {
	robotgo.KeyDown(key)
	sleepTime := rand.Intn(keyPressMaxTime-keyPressMinTime) + keyPressMinTime
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	robotgo.KeyUp(key)
}
