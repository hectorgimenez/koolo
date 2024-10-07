package winproc

import "golang.org/x/sys/windows"

const (
	EXECUTION_STATE_ES_DISPLAY_REQUIRED = 0x00000002
	EXECUTION_STATE_ES_CONTINUOUS       = 0x80000000
)

var (
	KERNEL32                = windows.NewLazySystemDLL("kernel32.dll")
	SetThreadExecutionState = KERNEL32.NewProc("SetThreadExecutionState")
)
