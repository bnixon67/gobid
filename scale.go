package main

import (
	"errors"
	"image"
	"io"

	"github.com/disintegration/imaging"
)

func ScaleDown(r io.ReadSeeker, maxWidth, maxHeight int) (image.Image, error) {
	if maxWidth == 0 && maxHeight == 0 {
		return nil, errors.New("invalid parameters: maxWidth and maxHeight are both 0")
	}

	src, err := imaging.Decode(r, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}

	srcWidth := src.Bounds().Dx()
	srcHeight := src.Bounds().Dy()

	// don't resize if source is already smaller
	if srcWidth <= maxWidth || srcHeight <= maxHeight {
		return src, err
	}

	/*
		// determine scaling ratio
		var ratio float64
		switch {
		case maxWidth == 0:
			ratio = float64(maxHeight) / float64(srcHeight)
		case maxHeight == 0:
			ratio = float64(maxWidth) / float64(srcWidth)
		default:
			ratio = math.Min(
				float64(maxWidth)/float64(srcWidth),
				float64(maxHeight)/float64(srcHeight),
			)
		}

		// scaled down dimension
		newWidth := int(math.Round(float64(srcWidth) * ratio))
		newHeight := int(math.Round(float64(srcHeight) * ratio))
	*/

	// Resize
	dst := imaging.Resize(src, maxWidth, maxHeight, imaging.Lanczos)

	return dst, err
}
