package game

import "github.com/lxn/win"

type HID struct {
	gr           *MemoryReader
	gi           *MemoryInjector
	buttonStates map[MouseButton]bool
}

func NewHID(gr *MemoryReader, gi *MemoryInjector) *HID {
	return &HID{
		gr:           gr,
		gi:           gi,
		buttonStates: make(map[MouseButton]bool),
	}
}
func (hid *HID) ReleaseMouseButton(btn MouseButton) {
	if hid.buttonStates[btn] {
		var pt win.POINT
		win.GetCursorPos(&pt)
		lParam := calculateLparam(int(pt.X), int(pt.Y))
		buttonUp := uint32(win.WM_LBUTTONUP)
		if btn == RightButton {
			buttonUp = win.WM_RBUTTONUP
		}

		win.SendMessage(hid.gr.HWND, buttonUp, 0, lParam)
		hid.buttonStates[btn] = false
	}
}
func (hid *HID) HoldMouseButton(btn MouseButton, x, y int) {
	if !hid.buttonStates[btn] {
		hid.MovePointer(x, y)
		x = hid.gr.WindowLeftX + x
		y = hid.gr.WindowTopY + y

		lParam := calculateLparam(x, y)
		buttonDown := uint32(win.WM_LBUTTONDOWN)
		if btn == RightButton {
			buttonDown = win.WM_RBUTTONDOWN
		}

		win.SendMessage(hid.gr.HWND, buttonDown, 1, lParam)
		hid.buttonStates[btn] = true
	}
}
func (hid *HID) IsButtonHeld(btn MouseButton) bool {
	return hid.buttonStates[btn]
}
