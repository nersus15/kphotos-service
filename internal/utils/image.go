package utils

import (
	"image"
	"log"
	"os"

	"github.com/disintegration/imaging"
)

// GenerateThumbnail membuat thumbnail 320px (lebar) dari input ke output
func GenerateThumbnail(inputPath, outputPath string) error {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return err
	}

	thumb := imaging.Resize(src, 320, 0, imaging.Lanczos)

	err = imaging.Save(thumb, outputPath)
	if err != nil {
		return err
	}

	return nil
}

// Optional: fungsi cek dimensi (bisa dipakai di upload handler)
func GetImageSize(path string) (width, height int) {
	file, err := os.Open(path)
	if err != nil {
		log.Println("get size:", err)
		return 0, 0
	}
	defer file.Close()

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		log.Println("decode size:", err)
		return 0, 0
	}
	return img.Width, img.Height
}

func OpenFile(path string) (*os.File, error) {
	return os.Open(path)
}
