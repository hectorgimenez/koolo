package helper

import (
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/hectorgimenez/koolo/internal/hid"
	"os"
	"time"
)

func Screenshot() {
	if _, err := os.Stat("screenshots"); os.IsNotExist(err) {
		err = os.MkdirAll("screenshots", 0700)
		if err != nil {
			return
		}
	}

	fileName := fmt.Sprintf("screenshots/error-%s.png", time.Now().Format("2006-01-02 15_04_05"))

	scale := robotgo.ScaleF()

	startX := int(float64(hid.WindowLeftX) * scale)
	startY := int(float64(hid.WindowTopY) * scale)

	width := int(float64(hid.GameAreaSizeX) * scale)
	height := int(float64(hid.GameAreaSizeY) * scale)

	robotgo.SaveCapture(fileName, startX, startY, width, height)
}
