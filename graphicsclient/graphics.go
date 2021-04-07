package graphicsclient

import (
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/BurntSushi/graphics-go/graphics"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

func New() *Client { return &Client{} }

type Client struct{}

func (x *Client) Blur(dst io.Writer, src io.Reader, stdDev float64, size int) (err error) {
	err = fmt.Errorf("invalid parameter")
	if dst != nil && src != nil && stdDev > 0 && size > 0 {
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
func (x *Client) Rotate(dst io.Writer, src io.Reader, angle float64) (err error) {
	err = fmt.Errorf("invalid parameter")
	if dst != nil && src != nil && angle != 0 {
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
func (x *Client) Scale(dst io.Writer, src io.Reader, w, h int) (err error) {
	err = fmt.Errorf("invalid parameter")
	if dst != nil && src != nil && w > 0 && h > 0 {
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
func (x *Client) Thumbnail(dst io.Writer, src io.Reader, w, h int) (err error) {
	err = fmt.Errorf("invalid parameter")
	if dst != nil && src != nil && w > 0 && h > 0 {
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
func (x *Client) DecodeConfig(src io.Reader) (conf image.Config, format string, err error) {
	return image.DecodeConfig(src)
}
func (x *Client) Decode(src io.Reader) (img image.Image, format string, err error) {
	return image.Decode(src)
}
func (x *Client) Encode(dst io.Writer, img image.Image, format string) (err error) {
	switch format {
	default:
		err = jpeg.Encode(dst, img, Options.jpeg)
	case "png":
		err = png.Encode(dst, img)
	case "gif":
		err = gif.Encode(dst, img, Options.gif)
	case "bmp":
		err = bmp.Encode(dst, img)
	case "tiff":
		err = tiff.Encode(dst, img, Options.tiff)
	}
	return
}
func (x *Client) SubImage(img image.Image, rect Rectangle) (sub image.Image) {
	if img != nil {
		if si, ok := img.(interface {
			SubImage(image.Rectangle) image.Image
		}); ok {
			sub = si.SubImage(rect)
		}
	}
	return
}

var Rect = image.Rect
var Options = &options{}

type Rectangle = image.Rectangle
type options struct {
	jpeg *jpeg.Options
	gif  *gif.Options
	tiff *tiff.Options
}

func (o *options) Jpeg(quality int) {
	o.jpeg = &jpeg.Options{Quality: quality}
}
func (o *options) Gif(numcolors int, quantizer draw.Quantizer, drawer draw.Drawer) {
	o.gif = &gif.Options{NumColors: numcolors, Quantizer: quantizer, Drawer: drawer}
}
func (o *options) Tiff(compression tiff.CompressionType, predictor bool) {
	o.tiff = &tiff.Options{Compression: compression, Predictor: predictor}
}
