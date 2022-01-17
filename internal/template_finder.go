package koolo

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"gocv.io/x/gocv"
)

type TemplateFinder struct {
	templates map[string]Template
}

type Template struct {
	RGB       *gocv.Mat
	GrayScale *gocv.Mat
}

type TemplateMatch struct {
	Score     float64
	PositionX int
	PositionY int
	Found     bool
}

func NewTemplateFinder(templatesPath string) (TemplateFinder, error) {
	templates := map[string]Template{}
	err := filepath.Walk(templatesPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.Contains(info.Name(), ".png") {
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
		templates[info.Name()] = Template{
			RGB:       &rgb,
			GrayScale: &grayScale,
		}

		return nil
	})
	if err != nil {
		return TemplateFinder{}, fmt.Errorf("error loading templates: %w", err)
	}

	return TemplateFinder{
		templates: templates,
	}, nil
}

func (tf *TemplateFinder) Search(tpl string, img gocv.Mat) TemplateMatch {
	mat, ok := tf.templates[tpl]
	if !ok {
		return TemplateMatch{}
	}

	fmt.Print(mat)

	return TemplateMatch{}
}
