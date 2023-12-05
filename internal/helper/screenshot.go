package helper

import (
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/kbinani/screenshot"
	"image"
	"image/png"
	"os"

	"github.com/hectorgimenez/koolo/internal/hid"
)

func Screenshot() image.Image {
	scale := ui.GameWindowScale()

	startX := int(float64(hid.WindowLeftX) * scale)
	startY := int(float64(hid.WindowTopY) * scale)

	width := int(float64(hid.GameAreaSizeX) * scale)
	height := int(float64(hid.GameAreaSizeY) * scale)

	// TODO: handle error
	img, _ := screenshot.Capture(startX, startY, width, height)

	return img
}

func SavePNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}
