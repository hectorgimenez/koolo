package koolo

import (
	"fmt"
	"github.com/kbinani/screenshot"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
	"image"
	"time"
)

const (
	gameWindowSearchTimeout = time.Second * 20
	ingameReferenceX        = 625
	ingameReferenceY        = 669
)

type Display struct {
	OffsetTop  int
	OffsetLeft int
	logger     *zap.Logger
}

func NewDisplay(display int, logger *zap.Logger) (Display, error) {
	logger.Info("Looking for D2R window. Make sure you are on the character selection screen.")
	tm := gocv.IMRead("assets/templates/main_menu_top_left.png", gocv.IMReadColor)
	tmIngame := gocv.IMRead("assets/templates/window_ingame_offset_reference.png", gocv.IMReadColor)

	t := time.Now()
	for time.Since(t) < gameWindowSearchTimeout {
		img, err := screenshot.CaptureDisplay(display)
		if err != nil {
			return Display{}, fmt.Errorf("error capturing display: %w", err)
		}
		mat, _ := gocv.ImageToMatRGB(img)
		res := gocv.NewMat()
		resIngame := gocv.NewMat()
		gocv.MatchTemplate(mat, tm, &res, gocv.TmCcoeffNormed, gocv.NewMat())
		gocv.MatchTemplate(mat, tmIngame, &resIngame, gocv.TmCcoeffNormed, gocv.NewMat())

		_, maxVal, _, maxPos := gocv.MinMaxLoc(res)
		_, maxValIngame, _, maxPosIngame := gocv.MinMaxLoc(resIngame)

		offsetLeft := maxPos.X
		offsetTop := maxPos.Y

		if maxValIngame > maxVal {
			maxVal = maxValIngame
			offsetLeft = maxPosIngame.X - ingameReferenceX
			offsetTop = maxPosIngame.Y - ingameReferenceY
		}

		if maxVal > 0.84 {
			logger.Info(fmt.Sprintf("D2R Window found, offsets: left %dpx, top %dpx", offsetLeft, offsetTop))

			return Display{
				OffsetTop:  offsetTop,
				OffsetLeft: offsetLeft,
				logger:     nil,
			}, err
		}
	}

	return Display{}, fmt.Errorf("D2R character selection screen could not be found")
}

func (d Display) Capture() image.Image {
	img, _ := screenshot.Capture(d.OffsetLeft, d.OffsetTop, 1280, 720)

	// TODO: Remove after debugging
	mat, _ := gocv.ImageToMatRGB(img)
	gocv.IMWrite(fmt.Sprintf("debug/%d.png", time.Now().Unix()), mat)

	return img
}
