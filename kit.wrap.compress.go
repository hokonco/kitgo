package kitgo

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"errors"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
)

var Compress compress_

type compress_ struct{}

func (compress_) New() *Client {
	gzipSign := []byte{0x1f, 0x8b, 0x08, 0, 0, 0x09, 0x6e, 0x88, 0, 0xff}
	return &Client{
		&Brotli{
			brotli.NewReader(nil),
			brotli.NewWriterLevel(nil, brotli.BestCompression),
		},
		&Flate{
			flate.NewReader(nil).(flateReader),
			func() *flate.Writer { w, _ := flate.NewWriter(nil, flate.BestCompression); return w }(),
		},
		&Gzip{
			func() *gzip.Reader { r, _ := gzip.NewReader(bytes.NewReader(gzipSign)); return r }(),
			func() *gzip.Writer { w, _ := gzip.NewWriterLevel(nil, gzip.BestCompression); return w }(),
		},
		&Minify{
			func() *minify.M {
				min := minify.New()
				min.AddFunc("text/css", css.Minify)
				min.AddFunc("text/xml", xml.Minify)
				min.AddFunc("text/html", html.Minify)
				min.AddFunc("image/svg+xml", svg.Minify)
				min.AddFunc("application/json", json.Minify)
				min.AddFunc("application/ld+json", json.Minify)
				min.AddFunc("application/javascript", js.Minify)
				return min
			}(),
			"",
		},
	}
}

type Client struct {
	*Brotli
	*Flate
	*Gzip
	*Minify
}

type Brotli struct {
	r *brotli.Reader
	w *brotli.Writer
}
type Flate struct {
	r flateReader
	w *flate.Writer
}
type Gzip struct {
	r *gzip.Reader
	w *gzip.Writer
}
type Minify struct {
	w         *minify.M
	mediatype string
}

func (x *Brotli) Read(dst io.Writer, src io.Reader) (err error) {
	if err = errors.New("invalid parameter"); dst != nil && src != nil {
		if err = x.r.Reset(src); err == nil {
			_, err = io.Copy(dst, x.r)
		}
	}
	return
}
func (x *Brotli) Write(dst io.Writer, src io.Reader) (err error) {
	if err = errors.New("invalid parameter"); dst != nil && src != nil {
		x.w.Reset(dst)
		if _, err = io.Copy(x.w, src); err == nil {
			err = x.w.Close()
		}
	}
	return
}
func (x *Flate) Read(dst io.Writer, src io.Reader) (err error) {
	if err = errors.New("invalid parameter"); dst != nil && src != nil {
		if err = x.r.Reset(src, nil); err == nil {
			if _, err = io.Copy(dst, x.r); err == nil {
				err = x.r.Close()
			}
		}
	}
	return
}
func (x *Flate) Write(dst io.Writer, src io.Reader) (err error) {
	if err = errors.New("invalid parameter"); dst != nil && src != nil {
		x.w.Reset(dst)
		if _, err = io.Copy(x.w, src); err == nil {
			err = x.w.Close()
		}
	}
	return
}
func (x *Gzip) Read(dst io.Writer, src io.Reader) (err error) {
	if err = errors.New("invalid parameter"); dst != nil && src != nil {
		if err = x.r.Reset(src); err == nil {
			if _, err = io.Copy(dst, x.r); err == nil {
				err = x.r.Close()
			}
		}
	}
	return
}
func (x *Gzip) Write(dst io.Writer, src io.Reader) (err error) {
	if err = errors.New("invalid parameter"); dst != nil && src != nil {
		x.w.Reset(dst)
		if _, err = io.Copy(x.w, src); err == nil {
			err = x.w.Close()
		}
	}
	return
}
func (x *Minify) WithMediaType(mediatype string) *Minify { x.mediatype = mediatype; return x }
func (x *Minify) Write(dst io.Writer, src io.Reader) (err error) {
	if err = errors.New("invalid parameter"); dst != nil && src != nil && x.mediatype != "" {
		err = x.w.Minify(x.mediatype, dst, src)
	}
	return
}

type flateReader interface {
	io.ReadCloser
	Reset(r io.Reader, dict []byte) error
}
