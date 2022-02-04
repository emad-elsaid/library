package main

import (
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path"

	"github.com/google/uuid"
	"golang.org/x/image/draw"
)

func ImageResize(in io.Reader, out io.Writer, w, h int) error {
	src, _, err := image.Decode(in)
	if err != nil {
		return err
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	return jpeg.Encode(out, dst, &jpeg.Options{Quality: 90})
}

func UploadImage(in io.Reader, p string, w, h int) (string, error) {
	name := uuid.New().String()

	out, err := os.Create(path.Join(p, name))
	if err != nil {
		return "", err
	}

	err = ImageResize(in, out, w, h)
	if err != nil {
		return "", err
	}

	return name, nil
}
