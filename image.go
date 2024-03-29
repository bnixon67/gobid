// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"image"
	"io"
	"log/slog"
	"os"

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

	// Resize
	dst := imaging.Resize(src, maxWidth, maxHeight, imaging.Lanczos)

	return dst, err
}

func SaveScaledJPEG(imgFile io.ReadSeeker, name string, maxWidth, maxHeight int) error {
	imgFile.Seek(0, io.SeekStart)

	img, err := ScaleDown(imgFile, maxWidth, maxHeight)
	if err != nil {
		return fmt.Errorf("could not ScaleDown: %v", err)
	}

	flag := os.O_CREATE | os.O_WRONLY | os.O_EXCL
	perm := os.FileMode(0o400)
	output, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return err
	}
	defer output.Close()

	imaging.Encode(output, img, imaging.JPEG, imaging.JPEGQuality(95))

	slog.Info("SaveScaledJPEG", "name", name, "flag", flag, "perm", perm)

	return err
}
