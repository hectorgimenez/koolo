package koolo

import (
	"fmt"
	"go.uber.org/zap"
	"image"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"gocv.io/x/gocv"
)

type TemplateFinder struct {
	templates map[string]Template
	logger    *zap.Logger
}

type Template struct {
	RGB       gocv.Mat
	GrayScale gocv.Mat
	Mask      gocv.Mat
}

type TemplateMatch struct {
	Score     float32
	PositionX int
	PositionY int
	Found     bool
}

func NewTemplateFinder(logger *zap.Logger, templatesPath string) (TemplateFinder, error) {
	templates := map[string]Template{}
	start := time.Now()
	err := filepath.Walk(templatesPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fileName := info.Name()
		if !strings.Contains(fileName, ".png") {
			return nil
		}

		mat := gocv.IMRead(path, gocv.IMReadUnchanged)
		if mat.Empty() {
			return nil
		}
		rgb := mat.Clone()
		grayScale := mat.Clone()
		gocv.CvtColor(mat, &rgb, gocv.ColorBGRAToBGR)
		gocv.CvtColor(mat, &grayScale, gocv.ColorBGRAToGray)
		templates[fileName[:len(fileName)-4]] = Template{
			RGB:       rgb,
			GrayScale: grayScale,
			Mask:      createMask(mat),
		}

		return nil
	})
	if err != nil {
		return TemplateFinder{}, fmt.Errorf("error loading templates: %w", err)
	}

	logger.Debug(fmt.Sprintf(
		"Found a total of %d templates, loaded in %dms",
		len(templates),
		time.Since(start).Milliseconds()),
	)

	return TemplateFinder{
		templates: templates,
		logger:    logger,
	}, nil
}

func createMask(tpl gocv.Mat) gocv.Mat {
	mask := gocv.NewMat()
	if mask.Channels() == 4 {

	}
	gocv.Threshold(tpl, &mask, 1, 255, gocv.ThresholdBinary)

	return mask
}

func (tf *TemplateFinder) Find(tplName string, img image.Image) TemplateMatch {
	t := time.Now()
	threshold := float32(0.65)

	mat, err := gocv.ImageToMatRGB(img)
	if err != nil {
		return TemplateMatch{}
	}

	tpl, ok := tf.templates[tplName]
	if !ok {
		return TemplateMatch{
			Score: 0,
			Found: false,
		}
	}
	res := gocv.NewMat()
	gocv.MatchTemplate(mat, tpl.RGB, &res, gocv.TmCcoeffNormed, gocv.NewMat())
	_, maxVal, _, maxPos := gocv.MinMaxLoc(res)
	if maxVal > threshold {
		tf.logger.Debug(fmt.Sprintf(
			"Found Template (%dms): %s Score: %f",
			time.Since(t).Milliseconds(),
			tplName,
			maxVal,
		))
		return TemplateMatch{
			Score:     maxVal,
			PositionX: 0,
			PositionY: 0,
			Found:     true,
		}
	}
	fmt.Sprintf("", maxPos)
	return TemplateMatch{}
}
