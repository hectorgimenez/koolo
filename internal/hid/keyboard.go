package hid

import (
	"github.com/go-vgo/robotgo"
	"math/rand"
	"time"
)

const (
	keyPressMinTime = 30  // ms
	keyPressMaxTime = 160 // ms
)

// PressKey toggles a key, it holds the key between keyPressMinTime and keyPressMaxTime ms randomly
func PressKey(key string) {
	robotgo.KeyDown(key)
	sleepTime := rand.Intn(keyPressMaxTime-keyPressMinTime) + keyPressMinTime
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	robotgo.KeyUp(key)
}

func KeyDown(key string) {
	robotgo.KeyDown(key)
}

func KeyUp(key string) {
	robotgo.KeyUp(key)
}

func PressKeyCombination(keys ...string) {
	for _, k := range keys {
		robotgo.KeyDown(k)
	}
	time.Sleep(time.Millisecond * 200)
	for _, k := range keys {
		robotgo.KeyUp(k)
	}
}
