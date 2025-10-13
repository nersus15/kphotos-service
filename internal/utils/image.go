package utils

import (
	"github.com/h2non/bimg"
)

func GenerateThumbnail(inputPath, outputPath string) error {
	buffer, err := bimg.Read(inputPath)
	if err != nil {
		return err
	}

	newImage, err := bimg.NewImage(buffer).Resize(320, 0)
	if err != nil {
		return err
	}

	return bimg.Write(outputPath, newImage)
}
