package utils

import (
	"image"
	"image/jpeg"
	"os"
)

func SaveImageJPEG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return jpeg.Encode(f, img, &jpeg.Options{Quality: 80})
}
