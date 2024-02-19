package hid

import (
	"github.com/hectorgimenez/koolo/internal/memory"
	"math/rand"
	"strings"
	"time"

	"github.com/lxn/win"
)

const (
	keyPressMinTime = 10 // ms
	keyPressMaxTime = 40 // ms
)

// PressKey toggles a key, it holds the key between keyPressMinTime and keyPressMaxTime ms randomly
func PressKey(key string) {
	asciiChar := getASCIICode(key)
	win.PostMessage(memory.HWND, win.WM_KEYDOWN, asciiChar, 0)
	sleepTime := rand.Intn(keyPressMaxTime-keyPressMinTime) + keyPressMinTime
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	win.PostMessage(memory.HWND, win.WM_KEYUP, asciiChar, 0)
}

func KeyDown(key string) {
	win.PostMessage(memory.HWND, win.WM_KEYDOWN, getASCIICode(key), 0)
}

func KeyUp(key string) {
	win.PostMessage(memory.HWND, win.WM_KEYUP, getASCIICode(key), 0)
}

func getASCIICode(key string) uintptr {
	if len(key) == 1 {
		return uintptr(strings.ToUpper(key)[0])
	}

	return specialChars[strings.ToLower(key)]
}

var specialChars = map[string]uintptr{
	"esc":       win.VK_ESCAPE,
	"enter":     win.VK_RETURN,
	"f1":        win.VK_F1,
	"f2":        win.VK_F2,
	"f3":        win.VK_F3,
	"f4":        win.VK_F4,
	"f5":        win.VK_F5,
	"f6":        win.VK_F6,
	"f7":        win.VK_F7,
	"f8":        win.VK_F8,
	"f9":        win.VK_F9,
	"f10":       win.VK_F10,
	"f11":       win.VK_F11,
	"f12":       win.VK_F12,
	"lctrl":     win.VK_LCONTROL,
	"home":      win.VK_HOME,
	"down":      win.VK_DOWN,
	"up":        win.VK_UP,
	"left":      win.VK_LEFT,
	"right":     win.VK_RIGHT,
	"tab":       win.VK_TAB,
	"space":     win.VK_SPACE,
	"alt":       win.VK_MENU,
	"lalt":      win.VK_LMENU,
	"ralt":      win.VK_RMENU,
	"shift":     win.VK_LSHIFT,
	"backspace": win.VK_BACK,
}
