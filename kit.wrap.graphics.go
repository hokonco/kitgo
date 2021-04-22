package kitgo

import (
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/BurntSushi/graphics-go/graphics"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

var Graphics graphics_

type graphics_ struct{}

func (graphics_) New(conf *GraphicsConfig) *GraphicsWrapper { return &GraphicsWrapper{conf} }

type GraphicsConfig struct {
	GIF  *gif.Options
	JPEG *jpeg.Options
	TIFF *tiff.Options
}

type GraphicsWrapper struct{ conf *GraphicsConfig }

func (x *GraphicsWrapper) Blur(dst io.Writer, src io.Reader, stdDev float64, size int) (err error) {
	if err = errors.New("invalid parameter"); dst != nil && src != nil && stdDev > 0 && size > 0 {
		var srcImg image.Image
		var format string
		if srcImg, format, err = x.Decode(src); err == nil {
			dstImg := image.NewRGBA(srcImg.Bounds())
			opt := &graphics.BlurOptions{StdDev: stdDev, Size: size}
			if err = graphics.Blur(dstImg, srcImg, opt); err == nil {
				err = x.Encode(dst, dstImg, format)
			}
		}
	}
	return
}
func (x *GraphicsWrapper) Rotate(dst io.Writer, src io.Reader, angle float64) (err error) {
	if err = errors.New("invalid parameter"); dst != nil && src != nil && angle != 0 {
		var srcImg image.Image
		var format string
		if srcImg, format, err = x.Decode(src); err == nil {
			dstImg := image.NewRGBA(srcImg.Bounds())
			opt := &graphics.RotateOptions{Angle: angle}
			if err = graphics.Rotate(dstImg, srcImg, opt); err == nil {
				err = x.Encode(dst, dstImg, format)
			}
		}
	}
	return
}
func (x *GraphicsWrapper) Scale(dst io.Writer, src io.Reader, w, h int) (err error) {
	if err = errors.New("invalid parameter"); dst != nil && src != nil && w > 0 && h > 0 {
		var srcImg image.Image
		var format string
		if srcImg, format, err = x.Decode(src); err == nil {
			dstImg := image.NewRGBA(image.Rect(0, 0, w, h))
			if err = graphics.Scale(dstImg, srcImg); err == nil {
				err = x.Encode(dst, dstImg, format)
			}
		}
	}
	return
}
func (x *GraphicsWrapper) Thumbnail(dst io.Writer, src io.Reader, w, h int) (err error) {
	if err = errors.New("invalid parameter"); dst != nil && src != nil && w > 0 && h > 0 {
		var srcImg image.Image
		var format string
		if srcImg, format, err = x.Decode(src); err == nil {
			dstImg := image.NewRGBA(image.Rect(0, 0, w, h))
			if err = graphics.Thumbnail(dstImg, srcImg); err == nil {
				err = x.Encode(dst, dstImg, format)
			}
		}
	}
	return
}
func (x *GraphicsWrapper) DecodeConfig(src io.Reader) (conf image.Config, format string, err error) {
	return image.DecodeConfig(src)
}
func (x *GraphicsWrapper) Decode(src io.Reader) (img image.Image, format string, err error) {
	return image.Decode(src)
}
func (x *GraphicsWrapper) Encode(dst io.Writer, img image.Image, format string) (err error) {
	if x.conf == nil {
		x.conf = &GraphicsConfig{}
	}
	switch format {
	case "bmp":
		err = bmp.Encode(dst, img)
	case "gif":
		err = gif.Encode(dst, img, x.conf.GIF)
	default:
		err = jpeg.Encode(dst, img, x.conf.JPEG)
	case "png":
		err = png.Encode(dst, img)
	case "tiff":
		err = tiff.Encode(dst, img, x.conf.TIFF)
	}
	return
}

func (x *GraphicsWrapper) SubImage(img image.Image, rect image.Rectangle) (sub image.Image) {
	if img != nil {
		type Image = image.Image
		if si, ok := img.(interface{ SubImage(image.Rectangle) Image }); ok {
			sub = si.SubImage(rect)
		}
	}
	return
}
func (x *GraphicsWrapper) Rect(x0, y0, x1, y1 int) image.Rectangle { return image.Rect(x0, y0, x1, y1) }
