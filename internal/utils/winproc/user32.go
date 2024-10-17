package winproc

import "golang.org/x/sys/windows"

var (
	USER32             = windows.NewLazySystemDLL("user32.dll")
	PrintWindow        = USER32.NewProc("PrintWindow")
	GetWindowDC        = USER32.NewProc("GetWindowDC")
	SetProcessDpiAware = USER32.NewProc("SetProcessDPIAware")
	SetWindowText      = USER32.NewProc("SetWindowTextW")
)
