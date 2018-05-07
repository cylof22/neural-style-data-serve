package WaterMark

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"os"
	"strconv"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"

	"golang.org/x/image/draw"
)

// ImageSizeLimitation equal to Memcached default size
const ImageSizeLimitation = 1024000

// Service define the basic information for watermark
type Service struct {
	SourceImg image.Image
	Source    string
	Text      string
	TextColor color.Color
	Scale     float64
	Format    string
}

// CreateWaterMark generate composed image
func (wm *Service) CreateWaterMark(output io.Writer) (image.Image, error) {
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
	waterMarkImg := image.NewRGBA(image.Rect(0, 0, int(float64(bounds.Max.X)*wm.Scale), int(float64(bounds.Max.Y)*wm.Scale)))
	draw.BiLinear.Scale(waterMarkImg, waterMarkImg.Bounds(), img, bounds, draw.Src, nil)

	var err error
	var source = wm.SourceImg
	if len(wm.Source) != 0 {
		// get the source file size
		reader, err := os.Open(wm.Source)
		defer reader.Close()
		if err != nil {
			return nil, err
		}
		source, _, err = image.Decode(reader)
		if err != nil {
			return nil, err
		}
	}

	sourceBounds := source.Bounds()

	// composited image
	markedImage := image.NewRGBA(sourceBounds)
	draw.Draw(markedImage, sourceBounds, source, image.ZP, draw.Src)

	// horrizontal
	watermarkBounds := waterMarkImg.Bounds()
	offset := image.Point{X: sourceBounds.Max.X - watermarkBounds.Max.X + 25, Y: sourceBounds.Max.Y - watermarkBounds.Max.Y - 4}
	draw.Draw(markedImage, watermarkBounds.Add(offset), waterMarkImg, image.ZP, draw.Over)

	imgSize := len(markedImage.Pix)
	markedBytes := make([]byte, 0)
	markedBuffer := bytes.NewBuffer(markedBytes)
	quality := jpeg.DefaultQuality

	if imgSize > ImageSizeLimitation {
		// downgrade the image below the 1M size
		for {
			err = jpeg.Encode(markedBuffer, markedImage, &jpeg.Options{Quality: quality})
			imgSize = len(markedBuffer.Bytes())
			if imgSize < ImageSizeLimitation {
				break
			}
			markedBuffer.Reset()
			fmt.Println("The image Size is " + strconv.Itoa(imgSize))
			quality = quality - 10
		}
		_, err = output.Write(markedBuffer.Bytes())
	} else {
		err = jpeg.Encode(output, markedImage, &jpeg.Options{Quality: 100})
	}

	if err != nil {
		return nil, err
	}

	return markedImage, err
}
