package game

import (
	"github.com/inkeliz/w32"
	"github.com/lxn/win"
	"math/rand"
	"strings"
	"time"
)

const (
	keyPressMinTime = 40 // ms
	keyPressMaxTime = 90 // ms
)

// PressKey toggles a key, it holds the key between keyPressMinTime and keyPressMaxTime ms randomly
func (hid *HID) PressKey(key string) {
	keyCode := hid.getASCIICode(key)
	win.PostMessage(hid.gr.HWND, win.WM_KEYDOWN, keyCode, hid.calculatelParam(keyCode, true))
	sleepTime := rand.Intn(keyPressMaxTime-keyPressMinTime) + keyPressMinTime
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	win.PostMessage(hid.gr.HWND, win.WM_KEYUP, keyCode, hid.calculatelParam(keyCode, false))
}

func (hid *HID) KeyDown(key string) {
	keyCode := hid.getASCIICode(key)
	win.PostMessage(hid.gr.HWND, win.WM_KEYDOWN, keyCode, hid.calculatelParam(keyCode, true))
}

func (hid *HID) KeyUp(key string) {
	keyCode := hid.getASCIICode(key)
	win.PostMessage(hid.gr.HWND, win.WM_KEYUP, keyCode, hid.calculatelParam(keyCode, false))
}

func (hid *HID) getASCIICode(key string) uintptr {
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
	"lwin":      win.VK_LWIN,
	"rwin":      win.VK_RWIN,
	"end":       win.VK_END,
	"-":         win.VK_OEM_MINUS,
}

func (hid *HID) calculatelParam(keyCode uintptr, down bool) uintptr {
	scanCode := int(w32.MapVirtualKey(uint(keyCode), w32.MAPVK_VK_TO_VSC))
	repeatCount := 1
	extendedKeyFlag := 0
	contextCode := 0
	previousKeyState := 0
	transitionState := 0
	if !down {
		transitionState = 1
	}

	lParam := uintptr((repeatCount & 0xFFFF) | (scanCode << 16) | (extendedKeyFlag << 24) | (contextCode << 29) | (previousKeyState << 30) | (transitionState << 31))
	return lParam
}
