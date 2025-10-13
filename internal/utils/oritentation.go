package utils

import (
	"image"
	"os"

	"github.com/rwcarlsen/goexif/exif"
)

func GetOrientation(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 1
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return 1
	}

	o, err := x.Get(exif.Orientation)
	if err != nil {
		return 1
	}

	orientation, err := o.Int(0)
	if err != nil {
		return 1
	}
	return orientation
}

func FixOrientation(img image.Image, orientation int) image.Image {
	switch orientation {
	case 3:
		return rotate180(img)
	case 6:
		return rotate90(img)
	case 8:
		return rotate270(img)
	default:
		return img
	}
}

// --- Rotasi utilitas sederhana ---
func rotate90(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dy(), b.Dx()))
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			dst.Set(b.Max.Y-y-1, x, img.At(x, y))
		}
	}
	return dst
}

func rotate180(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(b)
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			dst.Set(b.Max.X-x-1, b.Max.Y-y-1, img.At(x, y))
		}
	}
	return dst
}

func rotate270(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dy(), b.Dx()))
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			dst.Set(y, b.Max.X-x-1, img.At(x, y))
		}
	}
	return dst
}
