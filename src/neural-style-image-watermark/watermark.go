package WaterMark

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"

	"github.com/disintegration/gift"
	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"

	"golang.org/x/image/draw"
)

// WaterMark define the basic information for watermark
type WaterMark struct {
	Source    string
	Text      string
	TextColor color.Color
	Scale     float64
	Format    string
}

// CreateWaterMark generate composed image
func (wm *WaterMark) CreateWaterMark(output io.Writer) image.Image {
	var padding float64 = 2
	w := 8 * (float64(len(wm.Text)) + (padding * 2))
	h := 16 * padding
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	point := fixed.Point26_6{X: fixed.Int26_6(64 * padding), Y: fixed.Int26_6(h * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(wm.TextColor),
		Face: inconsolata.Regular8x16,
		Dot:  point,
	}
	d.DrawString(wm.Text)

	bounds := img.Bounds()
	scaled := image.NewRGBA(image.Rect(0, 0, int(float64(bounds.Max.X)*wm.Scale), int(float64(bounds.Max.Y)*wm.Scale)))
	draw.BiLinear.Scale(scaled, scaled.Bounds(), img, bounds, draw.Src, nil)

	g := gift.New(
		gift.Rotate(45, color.Transparent, gift.CubicInterpolation),
	)
	waterMarkImg := image.NewNRGBA(g.Bounds(scaled.Bounds()))
	g.Draw(waterMarkImg, scaled)

	// get the source file size
	reader, err := os.Open(wm.Source)
	defer reader.Close()

	source, format, err := image.Decode(reader)
	sourceBounds := source.Bounds()

	// composited image
	markedImage := image.NewRGBA(sourceBounds)
	//draw.Draw(markedImage, sourceBounds, source, image.ZP, draw.Src)

	// horrizontal
	watermarkBounds := waterMarkImg.Bounds()
	var offset image.Point
	for offset.X = watermarkBounds.Max.X / -2; offset.X < sourceBounds.Max.X; offset.X += watermarkBounds.Max.X {
		for offset.Y = watermarkBounds.Max.Y / -2; offset.Y < sourceBounds.Max.Y; offset.Y += watermarkBounds.Max.Y {
			draw.Draw(markedImage, watermarkBounds.Add(offset), waterMarkImg, image.ZP, draw.Over)
		}
	}

	fmt.Println("Start output file")
	// output the image
	switch wm.Format {
	case "png":
		err = png.Encode(output, markedImage)
	case "gif":
		err = gif.Encode(output, markedImage, &gif.Options{NumColors: 265})
	case "jpeg":
		err = jpeg.Encode(output, markedImage, &jpeg.Options{Quality: jpeg.DefaultQuality})
	default:
		log.Fatalf("unknown format %s", format)
	}
	if err != nil {
		log.Fatalf("unable to encode image: %s", err)
	}

	return markedImage
}
