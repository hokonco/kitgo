package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/http/httputil"
	"time"

	"github.com/hokonco/kitgo"
)

type Config struct {
	Timeout string `yaml:"timeout" json:"timeout"`
}

func New(cfg Config) *Client {
	d := kitgo.ParseDuration(cfg.Timeout, 30*time.Second)
	return &Client{&http.Client{Timeout: d}}
}

func Test() (client *Client, mock *Mock) {
	mux := http.NewServeMux()
	return New(Config{}), &Mock{mux, httptest.NewServer(mux)}
}

type Client struct{ *http.Client }

// =============================================================================
// MOCK
// =============================================================================

type Mock struct {
	mux *http.ServeMux
	*httptest.Server
}

func (x *Mock) Expect(method, pattern string, handler http.Handler) {
	x.mux.Handle(pattern, handler)
	x.Server = httptest.NewServer(x.mux)
}

func (x *Mock) URL(path string) string {
	return x.Server.URL + path
}

// =============================================================================
// TRANSPORT
// =============================================================================

// Transport implement http.RoundTripper, with 3 useful fields
//
// - Trace (*httptrace.ClientTrace) one can use `HTTPTrace` directly an modify its fields
//
// - RequestFn callback when creating *http.Request
//
// - ResponseFn callback when receiving *http.Response
type Transport struct {
	Base       http.RoundTripper
	Trace      *httptrace.ClientTrace
	OnRequest  func(*http.Request) error
	OnResponse func(*http.Response) error
}

type Trace = httptrace.ClientTrace

func (x *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	var _ http.RoundTripper = (*Transport)(nil)
	if t := x.Trace; t != nil {
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), t))
	}
	if x.Base == nil {
		x.Base = http.DefaultTransport
	}
	if x.OnRequest != nil {
		if _err := x.OnRequest(req); _err != nil {
			return nil, _err
		}
	}
	res, err := x.Base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if x.OnResponse != nil {
		if _err := x.OnResponse(res); _err != nil {
			return nil, _err
		}
	}
	return res, err
}

func DumpOnResponse(dst io.Writer, includeBody bool) func(*http.Response) error {
	return func(r *http.Response) error {
		b, err := httputil.DumpResponse(r, includeBody)
		_, _ = io.Copy(dst, bytes.NewReader(b))
		return err
	}
}
func DumpOnRequest(dst io.Writer, includeBody bool) func(*http.Request) error {
	return func(r *http.Request) error {
		b, err := httputil.DumpRequestOut(r, includeBody)
		_, _ = io.Copy(dst, bytes.NewReader(b))
		return err
	}
}
func NewTrace(cb func(*Trace)) *Trace {
	t := &Trace{}
	if cb != nil {
		cb(t)
	}
	return t
}
