package ui

import (
	"fmt"
	"image"
	"io/fs"
	"math"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

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

func NewTemplateFinder(logger *zap.Logger, templatesPath string) (*TemplateFinder, error) {
	templates := map[string]Template{}
	logger.Debug("Loading templates...")
	start := time.Now()
	err := filepath.Walk(templatesPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fileName := info.Name()
		if !strings.Contains(fileName, ".webp") {
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

		filename := strings.ReplaceAll(path, "assets\\", "")
		filename = strings.ReplaceAll(filename, "\\", "_")
		sanitizedName := strings.ReplaceAll(filename, ".webp", "")

		templates[sanitizedName] = Template{
			RGB:       rgb,
			GrayScale: grayScale,
			Mask:      createMask(mat),
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error loading templates: %w", err)
	}

	logger.Debug(fmt.Sprintf(
		"Found a total of %d templates, loaded in %dms",
		len(templates),
		time.Since(start).Milliseconds()),
	)

	return &TemplateFinder{
		templates: templates,
		logger:    logger,
	}, nil
}

func createMask(tpl gocv.Mat) gocv.Mat {
	mask := gocv.NewMat()
	if tpl.Channels() > 3 {
		imgChannel := gocv.Split(tpl)
		gocv.Threshold(imgChannel[3], &mask, 1, 255, gocv.ThresholdBinary)
	}

	return mask
}

func (tf *TemplateFinder) FindInArea(tplName string, img image.Image, x0, y0, x1, y1 int) TemplateMatch {
	t := time.Now()
	threshold := float32(0.80)
	colorDiffThreshold := float64(75)

	bigmat, err := gocv.ImageToMatRGB(img)
	if err != nil {
		return TemplateMatch{}
	}

	mat := gocv.NewMat()
	gocv.Resize(bigmat, &mat, image.Point{X: 1280, Y: 720}, 0, 0, gocv.InterpolationLinear)

	if x0 > 0 && y0 > 0 && x1 > 0 && y1 > 0 {
		croppedMat := mat.Region(image.Rect(x0, y0, x1, y1))
		mat = croppedMat.Clone()
	}

	tpl, ok := tf.templates[tplName]
	if !ok {
		tf.logger.Error("Template not found", zap.String("template", tplName))
		return TemplateMatch{
			Score: 0,
			Found: false,
		}
	}

	res := gocv.NewMat()
	gocv.MatchTemplate(mat, tpl.RGB, &res, gocv.TmCcoeffNormed, tpl.Mask)
	_, maxVal, _, maxPos := gocv.MinMaxLoc(res)
	if maxVal > threshold {
		region := mat.Region(image.Rect(maxPos.X, maxPos.Y, maxPos.X+tpl.RGB.Cols(), maxPos.Y+tpl.RGB.Rows()))
		regionMean := region.Mean()
		tplMean := tpl.RGB.Mean()
		absDiff := math.Abs((regionMean.Val1 - tplMean.Val1) + (regionMean.Val2 - tplMean.Val2) + (regionMean.Val3 - tplMean.Val3))
		if absDiff < colorDiffThreshold {
			tf.logger.Debug(fmt.Sprintf(
				"Found Template (%dms): %s Score: %f, ColorDiff: %f",
				time.Since(t).Milliseconds(),
				tplName,
				maxVal,
				absDiff,
			))

			return TemplateMatch{
				Score:     maxVal,
				PositionX: maxPos.X,
				PositionY: maxPos.Y,
				Found:     true,
			}
		}

	}

	return TemplateMatch{}
}

func (tf *TemplateFinder) Find(tplName string, img image.Image) TemplateMatch {
	return tf.FindInArea(tplName, img, 0, 0, 0, 0)
}
