package winproc

import "golang.org/x/sys/windows"

var (
	GDI32                  = windows.NewLazySystemDLL("gdi32.dll")
	CreateCompatibleDC     = GDI32.NewProc("CreateCompatibleDC")
	CreateCompatibleBitmap = GDI32.NewProc("CreateCompatibleBitmap")
	SelectObject           = GDI32.NewProc("SelectObject")
	DeleteObject           = GDI32.NewProc("DeleteObject")
	DeleteDC               = GDI32.NewProc("DeleteDC")
	GetDIBits              = GDI32.NewProc("GetDIBits")
)
