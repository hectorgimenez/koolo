package helper

import (
	"image"

	"github.com/go-vgo/robotgo"
	"github.com/hectorgimenez/koolo/internal/hid"
)

func Screenshot() image.Image {
	scale := robotgo.ScaleF()

	startX := int(float64(hid.WindowLeftX) * scale)
	startY := int(float64(hid.WindowTopY) * scale)

	width := int(float64(hid.GameAreaSizeX) * scale)
	height := int(float64(hid.GameAreaSizeY) * scale)

	return robotgo.CaptureImg(startX, startY, width, height)
}
