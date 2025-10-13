package utils

import (
	"image"
	"io"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

// GetOrientation membaca orientasi EXIF
func GetOrientation(r io.Reader) int {
	x, err := exif.Decode(r)
	if err != nil {
		return 1
	}

	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return 1
	}

	orient, err := tag.Int(0)
	if err != nil {
		return 1
	}

	return orient
}

// FixOrientation memutar gambar sesuai orientasi EXIF
func FixOrientation(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.Transpose(img)
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Transverse(img)
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}
