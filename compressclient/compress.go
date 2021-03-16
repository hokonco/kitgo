package compressclient

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"sync"

	"github.com/andybalholm/brotli"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
)

func New() *Client {
	gzipSign := []byte{0x1f, 0x8b, 0x08, 0, 0, 0x09, 0x6e, 0x88, 0, 0xff}
	return &Client{
		&sync.Pool{New: func() interface{} { r := brotli.NewReader(nil); return r }},
		&sync.Pool{New: func() interface{} { r, _ := gzip.NewReader(bytes.NewReader(gzipSign)); return r }},
		&sync.Pool{New: func() interface{} { r := flate.NewReader(nil); return r.(flateReader) }},
		&sync.Pool{New: func() interface{} { w := brotli.NewWriterLevel(nil, brotli.BestCompression); return w }},
		&sync.Pool{New: func() interface{} { w, _ := gzip.NewWriterLevel(nil, gzip.BestCompression); return w }},
		&sync.Pool{New: func() interface{} { w, _ := flate.NewWriter(nil, flate.BestCompression); return w }},
		&sync.Pool{New: func() interface{} {
			min := minify.New()
			min.AddFunc("text/css", css.Minify)
			min.AddFunc("text/xml", xml.Minify)
			min.AddFunc("text/html", html.Minify)
			min.AddFunc("image/svg+xml", svg.Minify)
			min.AddFunc("application/json", json.Minify)
			min.AddFunc("application/ld+json", json.Minify)
			min.AddFunc("application/javascript", js.Minify)
			return min
		}},
	}
}

type Client struct {
	poolRBr *sync.Pool
	poolRGz *sync.Pool
	poolRFl *sync.Pool
	poolWBr *sync.Pool
	poolWGz *sync.Pool
	poolWFl *sync.Pool
	poolMin *sync.Pool
}

func (x *Client) ReadBrotli(dst io.Writer, src io.Reader) (err error) {
	if dst != nil && src != nil {
		if rbr, ok := x.poolRBr.Get().(*brotli.Reader); ok && rbr != nil {
			func() { _ = rbr.Reset(src); _, err = io.Copy(dst, rbr) }()
			x.poolRBr.Put(rbr)
		}
	}
	return
}
func (x *Client) ReadGzip(dst io.Writer, src io.Reader) (err error) {
	if dst != nil && src != nil {
		if rgz, ok := x.poolRGz.Get().(*gzip.Reader); ok {
			func() { _ = rgz.Reset(src); _, err = io.Copy(dst, rgz); _ = rgz.Close() }()
			x.poolRGz.Put(rgz)
		}
	}
	return
}
func (x *Client) ReadDeflate(dst io.Writer, src io.Reader) (err error) {
	if dst != nil && src != nil {
		if rfl, ok := x.poolRFl.Get().(flateReader); ok && rfl != nil {
			func() { _ = rfl.Reset(src, nil); _, err = io.Copy(dst, rfl); _ = rfl.Close() }()
			x.poolRFl.Put(rfl)
		}
	}
	return
}
func (x *Client) WriteBrotli(dst io.Writer, src io.Reader) (err error) {
	if dst != nil && src != nil {
		if wbr, ok := x.poolWBr.Get().(*brotli.Writer); ok && wbr != nil {
			func() { wbr.Reset(dst); _, err = io.Copy(wbr, src); _ = wbr.Close() }()
			x.poolWBr.Put(wbr)
		}
	}
	return
}
func (x *Client) WriteGzip(dst io.Writer, src io.Reader) (err error) {
	if dst != nil && src != nil {
		if wgz, ok := x.poolWGz.Get().(*gzip.Writer); ok && wgz != nil {
			func() { wgz.Reset(dst); _, err = io.Copy(wgz, src); _ = wgz.Close() }()
			x.poolWGz.Put(wgz)
		}
	}
	return
}
func (x *Client) WriteDeflate(dst io.Writer, src io.Reader) (err error) {
	if dst != nil && src != nil {
		if wfl, ok := x.poolWFl.Get().(*flate.Writer); ok && wfl != nil {
			func() { wfl.Reset(dst); _, err = io.Copy(wfl, src); _ = wfl.Close() }()
			x.poolWFl.Put(wfl)
		}
	}
	return
}
func (x *Client) Minify(dst io.Writer, src io.Reader, mediatype string) (err error) {
	if dst != nil && src != nil && mediatype != "" {
		if min, ok := x.poolMin.Get().(*minify.M); ok && min != nil {
			err = min.Minify(mediatype, dst, src)
			x.poolMin.Put(min)
		}
	}
	return
}

type flateReader interface {
	io.ReadCloser
	Reset(r io.Reader, dict []byte) error
}
