package httphandler

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/compressclient"
)

// =============================================================================
// Public
// =============================================================================

// ServeReverseProxy return a new http.Handler by setting *ReverseProxy
func ServeReverseProxy(target string, cb func(*ReverseProxy)) http.Handler {
	u, _ := url.Parse(target)
	rp := httputil.NewSingleHostReverseProxy(u)
	if cb != nil {
		cb(rp)
	}
	return rp
}

// MaxBytesReader buffered the *http.Request.Body with n length
func MaxBytesReader(n int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mr := r
		mr.Body = http.MaxBytesReader(w, mr.Body, n)
		*r = *mr
	})
}

// Redirect is alias of http.Redirect
func Redirect(code int, url string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, code)
	})
}

// ServeContent is alias of http.ServeContent
func ServeContent(name string, modtime time.Time, content io.ReadSeeker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, name, modtime, content)
	})
}

// ServeFile is alias of http.ServeFile
func ServeFile(dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, dir)
	})
}

// SetCookie is alias of http.SetCookie
func SetCookie(cookie *http.Cookie) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, cookie)
	})
}

// ServeHTTP with code, header, body & compression
func ResponseWith(state *responseState) http.Handler {
	statusCode, header, body, compression := state.statusCode, state.header, state.body, state.compression
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if http.StatusText(statusCode) == "" {
			statusCode = http.StatusInternalServerError
			body = []byte(http.StatusText(statusCode) + "\n")
		}
		if len(header) < 1 {
			header = http.Header{}
		}
		// when error, header set from http.Error
		if statusCode >= 400 {
			header.Set(ContentType, "text/plain; charset=utf-8")
			header.Set(XContentTypeOptions, "nosniff")
		}

		b := &bytes.Buffer{}
		c := compressclient.New()
		read := func(p []byte) io.Reader { return bytes.NewReader(p) }
		defer b.Reset()
		for i := 0; i < len(compression); i++ {
			if compression[i] == minify {
				if w.Header().Get(ContentType) == "" {
					w.Header().Set(ContentType, http.DetectContentType(body))
				}
				b.Reset()
				_ = c.Minify(b, read(body), w.Header().Get(ContentType))
				body = b.Bytes()
			}
		}
		nce := NegotiateContentEncoding(r, compression...)
		switch {
		case nce == br:
			b.Reset()
			_ = c.WriteBrotli(b, read(body))
			body = b.Bytes()
			w.Header().Set(ContentEncoding, br)
		case nce == gzip:
			b.Reset()
			_ = c.WriteGzip(b, read(body))
			body = b.Bytes()
			w.Header().Set(ContentEncoding, gzip)
		case nce == deflate:
			b.Reset()
			_ = c.WriteDeflate(b, read(body))
			body = b.Bytes()
			w.Header().Set(ContentEncoding, deflate)
		}
		for k := range header {
			w.Header()[k] = header[k]
		}
		w.WriteHeader(statusCode)
		_, _ = w.Write(body)
	})
}

// NewResponseState will create ResponseState
func NewResponseState(code int, header http.Header, body []byte, compression ...string) *responseState {
	return &responseState{code, header, body, compression}
}

// SetStateToRequest will set code, header, and body to *http.Request context,
// so that the next handler that shared the same *http.Request are able
// to access it, compression parameter is optional so that this handler
// properly compress the response
//
// *responseState is created via NewResponseState
func SetStateToRequest(r *http.Request, state *responseState) {
	*r = *(set(r, ctxKeyResponseState{}, state))
}

// GetStateFromRequest will get code, header, and body from *http.Request
// context set by previous handler
func GetStateFromRequest(r *http.Request) (state *responseState) {
	state, _ = r.Context().Value(ctxKeyResponseState{}).(*responseState)
	return
}

// GetNamedArgsFromRequest is a helper function that extract url.Values that have
// been parsed using MuxMatcherPattern, url.Values should not be empty if
// parsing is successful and should be able to extract further following
// url.Values, same keys in the pattern result in new value added in url.Values
func GetNamedArgsFromRequest(r *http.Request) url.Values {
	u, _ := r.Context().Value(ctxKeyNamedArgs{}).(url.Values)
	return u
}

// PanicRecoveryFromRequest is a helper function that extract error value
// when panic occured, the value is saved to *http.Request after recovery
// process and right before calling mux.PanicHandler
func PanicRecoveryFromRequest(r *http.Request) interface{} {
	return get(r, ctxKeyPanicRecovery{})
}

// ServerHijack
func Hijack(w http.ResponseWriter) (c net.Conn, rw *bufio.ReadWriter, err error) {
	err = kitgo.NewError(`http.ResponseWriter is not implementing http.Hijacker`)
	if x, ok := w.(http.Hijacker); ok && x != nil {
		c, rw, err = x.Hijack()
	}
	return
}

// ServerPush
func Push(w http.ResponseWriter, target string, opts *http.PushOptions) (err error) {
	err = kitgo.NewError(`http.ResponseWriter is not implementing http.Pusher`)
	if x, ok := w.(http.Pusher); ok && x != nil {
		err = x.Push(target, opts)
	}
	return
}

// ServerSentEvent
func SentEvent(w http.ResponseWriter, p []byte) (n int, err error) {
	err = kitgo.NewError(`http.ResponseWriter is not implementing http.Flusher`)
	if x, ok := w.(http.Flusher); ok && x != nil {
		n, err = w.Write(p)
		x.Flush()
	}
	return
}

// NegotiateContentEncoding returns the best offered content encoding for the
// request's Accept-Encoding header. If two offers match with equal weight and
// then the offer earlier in the list is preferred. If no offers are
// acceptable, then "" is returned.
func NegotiateContentEncoding(r *http.Request, offers ...string) string {
	bestQ, bestOffer, key := -1.0, "identity", AcceptEncoding
	expectTokenSlash := func(s string) (token, rest string) {
		i := 0
		for ; i < len(s); i++ {
			b := s[i]
			if (octetTypes[b]&isToken == 0) && b != '/' {
				break
			}
		}
		return s[:i], s[i:]
	}
	expectQuality := func(s string) (q float64, rest string) {
		switch {
		case s[0] == '0':
			q = 0
		case s[0] == '1':
			q = 1
		default:
			return -1, ""
		}
		s = s[1:]
		if !strings.HasPrefix(s, ".") {
			return q, s
		}
		s = s[1:]
		i := 0
		n := 0
		d := 1
		for ; i < len(s); i++ {
			b := s[i]
			if b < '0' || b > '9' {
				break
			}
			n = n*10 + int(b) - '0'
			d *= 10
		}
		return q + float64(n)/float64(d), s[i:]
	}
	skipSpace := func(s string) (rest string) {
		i := 0
		for ; i < len(s); i++ {
			if octetTypes[s[i]]&isSpace == 0 {
				break
			}
		}
		return s[i:]
	}
	type acceptSpec struct {
		Value string
		Q     float64
	}
	var specs []acceptSpec
loop:
	for _, s := range r.Header[key] {
		for {
			var spec acceptSpec
			spec.Value, s = expectTokenSlash(s)
			if spec.Value == "" {
				continue loop
			}
			spec.Q = 1.0
			s = skipSpace(s)
			if strings.HasPrefix(s, ";") {
				s = skipSpace(s[1:])
				if !strings.HasPrefix(s, "q=") {
					continue loop
				}
				spec.Q, s = expectQuality(s[2:])
				if spec.Q < 0.0 {
					continue loop
				}
			}
			specs = append(specs, spec)
			s = skipSpace(s)
			if !strings.HasPrefix(s, ",") {
				continue loop
			}
			s = skipSpace(s[1:])
		}
	}
	for _, offer := range offers {
		for _, spec := range specs {
			if spec.Q > bestQ && (spec.Value == "*" || spec.Value == offer) {
				bestQ = spec.Q
				bestOffer = offer
			}
		}
	}
	if bestQ == 0 {
		bestOffer = ""
	}
	return bestOffer
}

// =============================================================================
// Test
// =============================================================================

// Test is a convenient function with purpose of wrapping *httptest.ResponseRecorder
// and *http.Request in the same function
func Test(method, target string, body io.Reader) (*httptest.ResponseRecorder, *http.Request) {
	return httptest.NewRecorder(), httptest.NewRequest(method, target, body)
}

// TestResponse check w so that it fulfill code, header & body accordingly
func TestResponse(w *httptest.ResponseRecorder, code int, header http.Header, body []byte) bool {
	if len(header) < 1 {
		header = http.Header{}
	}
	codeEq := w.Code == code
	bodyEq := bytes.Equal(w.Body.Bytes(), body)
	headerEq := len(w.Header()) == len(header)
	for k := range w.Header() {
		for i := range w.Header()[k] {
			headerEq = headerEq && i < len(header[k])
			headerEq = headerEq && len(header[k]) == len(w.Header()[k])
			headerEq = headerEq && header[k][i] == w.Header()[k][i]
		}
	}
	if !headerEq {
		fmt.Printf("Header:	\n\tExpect:	%s\n\tActual:	%s\n", header, w.Header())
	}
	if !codeEq {
		fmt.Printf("Code:	\n\tExpect:	%d\n\tActual:	%d\n", code, w.Code)
	}
	if !bodyEq {
		fmt.Printf("Body:	\n\tExpect:	%q\n\tActual:	%q\n", body, w.Body.Bytes())
	}
	return headerEq && codeEq && bodyEq
}

type TestResponseWriter struct{ http.ResponseWriter }

func (TestResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	var _ http.Hijacker = (*TestResponseWriter)(nil)
	return nil, nil, nil
}
func (TestResponseWriter) Flush() {
	var _ http.Flusher = (*TestResponseWriter)(nil)
}
func (TestResponseWriter) Push(string, *http.PushOptions) error {
	var _ http.Pusher = (*TestResponseWriter)(nil)
	return nil
}

// =============================================================================
// Private
// =============================================================================

func jsonBytes(v interface{}) []byte { b, _ := kitgo.JSON.Marshal(v); return b }

func uniqueMuxMatcher(muxes []MuxMatcher) (nMuxes []MuxMatcher) {
	for i := range muxes {
		_ = muxes[i].Test()
		skip, si, bi := false, muxes[i].GoString(), jsonBytes(muxes[i])
		for j := range muxes {
			_ = muxes[j].Test()
			skip = i != j && (skip || (si == muxes[j].GoString() && bytes.Equal(bi, jsonBytes(muxes[j]))))
		}
		if !skip {
			nMuxes = append(nMuxes, muxes[i])
		}
	}
	return
}
func uniqueString(strs []string) (nStrs []string) {
	for i := range strs {
		skip := false
		for j := range strs {
			skip = i != j && (skip || strs[i] == strs[j])
		}
		if !skip {
			nStrs = append(nStrs, strs[i])
		}
	}
	return
}
func set(r *http.Request, key, val interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, val))
}
func get(r *http.Request, key interface{}) interface{} {
	return r.Context().Value(key)
}

// responseState is a state prior writing to http.ResponseWriter,
// use NewResponseState to provide optional compression type
type responseState struct {
	statusCode  int
	header      http.Header
	body        []byte
	compression []string
}
type ctxKeyNamedArgs struct{}
type ctxKeyPanicRecovery struct{}
type ctxKeyResponseState struct{}
type ReverseProxy = httputil.ReverseProxy
type octetType byte

const (
	isToken octetType = 1 << iota
	isSpace

	minify  = "minify"
	br      = "br"
	gzip    = "gzip"
	deflate = "deflate"

	AcceptEncoding      = "Accept-Encoding"
	ContentType         = "Content-Type"
	ContentEncoding     = "Content-Encoding"
	XContentTypeOptions = "X-Content-Type-Options"
)

// Octet types from RFC 2616.
var octetTypes = func() (octetTypes [256]octetType) {
	// OCTET      = <any 8-bit sequence of data>
	// CHAR       = <any US-ASCII character (octets 0 - 127)>
	// CTL        = <any US-ASCII control character (octets 0 - 31) and DEL (127)>
	// CR         = <US-ASCII CR, carriage return (13)>
	// LF         = <US-ASCII LF, linefeed (10)>
	// SP         = <US-ASCII SP, space (32)>
	// HT         = <US-ASCII HT, horizontal-tab (9)>
	// <">        = <US-ASCII double-quote mark (34)>
	// CRLF       = CR LF
	// LWS        = [CRLF] 1*( SP | HT )
	// TEXT       = <any OCTET except CTLs, but including LWS>
	// separators = "(" | ")" | "<" | ">" | "@" | "," | ";" | ":" | "\" | <">
	//              | "/" | "[" | "]" | "?" | "=" | "{" | "}" | SP | HT
	// token      = 1*<any CHAR except CTLs or separators>
	// qdtext     = <any TEXT except <">>

	for c := 0; c < 256; c++ {
		var t octetType
		isCtl := c <= 31 || c == 127
		isChar := 0 <= c && c <= 127
		isSeparator := strings.ContainsRune(" \t\"(),/:;<=>?@[]\\{}", rune(c))
		if strings.ContainsRune(" \t\r\n", rune(c)) {
			t |= isSpace
		}
		if isChar && !isCtl && !isSeparator {
			t |= isToken
		}
		octetTypes[c] = t
	}
	return
}()
