package game

import (
	"image"
	"unsafe"

	winproc2 "github.com/hectorgimenez/koolo/internal/utils/winproc"
)

func (gd *MemoryReader) Screenshot() image.Image {
	// Create a device context compatible with the window
	hdcWindow, _, _ := winproc2.GetWindowDC.Call(uintptr(gd.HWND))
	hdcMem, _, _ := winproc2.CreateCompatibleDC.Call(hdcWindow)
	hbmMem, _, _ := winproc2.CreateCompatibleBitmap.Call(hdcWindow, uintptr(gd.GameAreaSizeX), uintptr(gd.GameAreaSizeY))
	_, _, _ = winproc2.SelectObject.Call(hdcMem, hbmMem)

	// Use PrintWindow to copy the window into the bitmap
	winproc2.PrintWindow.Call(uintptr(gd.HWND), hdcMem, 3) // use 3 to get window content only

	// map the bitmap structure
	bmpInfo := struct {
		BiSize            uint32
		BiWidth, BiHeight int32
		BiPlanes          uint16
		BiBitCount        uint16
		BiCompression     uint32
		BiSizeImage       uint32
		BiXPelsPerMeter   int32
		BiYPelsPerMeter   int32
		BiClrUsed         uint32
		BiClrImportant    uint32
	}{
		BiSize:        40, // The size of the BITMAPINFOHEADER structure
		BiWidth:       int32(gd.GameAreaSizeX),
		BiHeight:      -int32(gd.GameAreaSizeY), // negative to indicate top-down bitmap
		BiPlanes:      1,
		BiBitCount:    32, // 32 bits-per-pixel
		BiCompression: 0,  // BI_RGB, no compression
		BiSizeImage:   0,  // 0 for BI_RGB
	}

	bufSize := gd.GameAreaSizeX * gd.GameAreaSizeY * 4
	buf := make([]byte, bufSize)
	winproc2.GetDIBits.Call(
		hdcMem,
		hbmMem,
		0,
		uintptr(gd.GameAreaSizeY),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bmpInfo)),
		0, // DIB_RGB_COLORS
	)

	// Convert raw bytes to *image.RGBA
	img := image.NewRGBA(image.Rect(0, 0, gd.GameAreaSizeX, gd.GameAreaSizeY))
	copy(img.Pix, buf)

	// Windows is using BRG instead of RGB, let's swap red and blue layers
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			idx := y*img.Stride + x*4 // Calculate index for the start of the pixel
			// Swap red and blue (at idx and idx+2)
			img.Pix[idx], img.Pix[idx+2] = img.Pix[idx+2], img.Pix[idx]
		}
	}

	// Cleanup
	_, _, _ = winproc2.DeleteObject.Call(hbmMem)
	_, _, _ = winproc2.DeleteDC.Call(hdcMem)

	return img
}
