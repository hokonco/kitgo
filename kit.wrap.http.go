package kitgo

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"runtime/debug"
	"sort"
	"strings"
	"syscall"
	"time"
)

var HTTP http_

type http_ struct {
	Client    httpClient_
	Server    httpServer_
	Transport httpTransport_
	Handler   httpHandler_
}

// =============================================================================
// CLIENT
// =============================================================================

type httpClient_ struct{}

func (httpClient_) New() *HTTPClientWrapper {
	return &HTTPClientWrapper{&http.Client{}}
}

func (httpClient_) Test() (*HTTPClientWrapper, *HTTPClientMock) {
	mux := http.NewServeMux()
	mock := &HTTPClientMock{httptest.NewServer(mux), mux}
	return HTTP.Client.New(), mock
}

type HTTPClientWrapper struct{ *http.Client }

type HTTPClientMock struct {
	*httptest.Server

	mux *http.ServeMux
}

// func (x *HTTPClientMock) NewRequest(method, target string, body io.Reader) *http.Request {
// 	return httptest.NewRequest(method, target, body)
// }
func (x *HTTPClientMock) Expect(method, pattern string, handler http.Handler) {
	x.mux.Handle(pattern, handler)
	x.Server = httptest.NewServer(x.mux)
}

func (x *HTTPClientMock) URL(path string) string {
	return x.Server.URL + path
}

// =============================================================================
// TRANSPORT
// =============================================================================

type httpTransport_ struct{}

func (httpTransport_) New() HTTPTransportWrapper { return HTTPTransportWrapper{} }

// HTTPTransportWrapper implement http.RoundTripper, with 3 useful fields
//
// - Trace (*httptrace.ClientTrace) one can use `HTTPTrace` directly an modify its fields
//
// - RequestFn callback when creating *http.Request
//
// - ResponseFn callback when receiving *http.Response
type HTTPTransportWrapper struct {
	base       http.RoundTripper
	trace      *HTTPClientTrace
	onRequest  func(*http.Request) error
	onResponse func(*http.Response) error
}

type HTTPClientTrace = httptrace.ClientTrace

func (x HTTPTransportWrapper) WithBase(v http.RoundTripper) HTTPTransportWrapper {
	x.base = v
	return x
}
func (x HTTPTransportWrapper) WithTrace(v *HTTPClientTrace) HTTPTransportWrapper {
	x.trace = v
	return x
}
func (x HTTPTransportWrapper) OnRequest(v func(*http.Request) error) HTTPTransportWrapper {
	x.onRequest = v
	return x
}
func (x HTTPTransportWrapper) OnResponse(v func(*http.Response) error) HTTPTransportWrapper {
	x.onResponse = v
	return x
}

func (x HTTPTransportWrapper) RoundTrip(req *http.Request) (res *http.Response, err error) {
	var _ http.RoundTripper = HTTPTransportWrapper{}
	if x.trace != nil {
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), x.trace))
	}
	// request
	if x.onRequest != nil {
		if err = x.onRequest(req); err != nil {
			return nil, err
		}
	}
	// roundtrip
	if x.base == nil {
		x.base = http.DefaultTransport
	}
	res, err = x.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	// response
	if x.onResponse != nil {
		if err = x.onResponse(res); err != nil {
			return nil, err
		}
	}
	return res, err
}

func (x HTTPTransportWrapper) DumpOnResponse(dst io.Writer, includeBody bool) HTTPTransportWrapper {
	x.onResponse = func(r *http.Response) error {
		b, err := httputil.DumpResponse(r, includeBody)
		_, _ = io.Copy(dst, bytes.NewReader(b))
		return err
	}
	return x
}
func (x HTTPTransportWrapper) DumpOnRequest(dst io.Writer, includeBody bool) HTTPTransportWrapper {
	x.onRequest = func(r *http.Request) error {
		b, err := httputil.DumpRequestOut(r, includeBody)
		_, _ = io.Copy(dst, bytes.NewReader(b))
		return err
	}
	return x
}

// =============================================================================
// SERVER
// =============================================================================

type httpServer_ struct{}

func (httpServer_) New() HTTPServerWrapper { return HTTPServerWrapper{new(http.Server)} }

type HTTPServerWrapper struct{ *http.Server }

// Run is extended method that wrap `ListenAndServe` and `ListenAndServeTLS`
// together and supposedly doing a graceful shutdown
func (x HTTPServerWrapper) Run(onInfo func(string), onError func(error), shutdownTimeout time.Duration) error {
	if onInfo == nil {
		onInfo = func(string) {}
	}
	if onError == nil {
		onError = func(error) {}
	}
	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Second
	}

	errChan := make(chan error, 1)
	go func() {
		// --> on receiving signal, return error
		sig := ListenToSignal(syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGSTOP, syscall.SIGTSTP)
		errChan <- fmt.Errorf("signal: [%d] %s\n%s", sig, sig, debug.Stack())
	}()
	if x.Server.TLSConfig == nil {
		go func() { errChan <- x.Server.ListenAndServe() }() // --> on srv error
		onInfo(fmt.Sprintf("srv running on http://0.0.0.0%v", x.Server.Addr))
	} else {
		go func() { errChan <- x.Server.ListenAndServeTLS("", "") }() // --> on tls error
		onInfo(fmt.Sprintf("tls running on https://0.0.0.0%v", x.Server.Addr))
	}
	onError(<-errChan)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err := x.Server.Shutdown(ctx)
	onError(err)
	return err
}

// =============================================================================
// HANDLER
// =============================================================================

type httpHandler_ struct{ MuxMatcher httpMuxMatcher }

// ServeReverseProxy return a new http.Handler by setting *ReverseProxy
func (httpHandler_) ServeReverseProxy(target string, cb func(*HTTPReverseProxy)) http.Handler {
	u, _ := url.Parse(target)
	rp := httputil.NewSingleHostReverseProxy(u)
	if cb != nil {
		cb(rp)
	}
	return rp
}

// MaxBytesReader buffered the *http.Request.Body with n length
func (httpHandler_) MaxBytesReader(n int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mr := r
		mr.Body = http.MaxBytesReader(w, mr.Body, n)
		*r = *mr
	})
}

// Redirect is alias of http.Redirect
func (httpHandler_) Redirect(code int, url string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, code)
	})
}

// ServeContent is alias of http.ServeContent
func (httpHandler_) ServeContent(name string, modtime time.Time, content io.ReadSeeker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, name, modtime, content)
	})
}

// ServeFile is alias of http.ServeFile
func (httpHandler_) ServeFile(dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, dir)
	})
}

// SetCookie is alias of http.SetCookie
func (httpHandler_) SetCookie(cookie *http.Cookie) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, cookie)
	})
}

// ResponseWith given state of code, header, body & compression that can be
// build from NewResponseState, `*responseState` is also able to set in
// different middleware by calling SetStateToRequest and read from other
// middleware via GetStateFromRequest
func (httpHandler_) ResponseWith(state *responseState) http.Handler {
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

		b := new(bytes.Buffer)
		c := Compress.New()
		read := func(p []byte) io.Reader { return bytes.NewReader(p) }
		defer b.Reset()
		for i := 0; i < len(compression); i++ {
			if compression[i] == "minify" {
				if w.Header().Get(ContentType) == "" {
					w.Header().Set(ContentType, http.DetectContentType(body))
				}
				b.Reset()
				_ = c.Minify.WithMediaType(w.Header().Get(ContentType)).Write(b, read(body))
				body = b.Bytes()
			}
		}
		nce := HTTP.Handler.NegotiateContentEncoding(r, compression...)
		switch {
		case nce == "br":
			b.Reset()
			_ = c.Brotli.Write(b, read(body))
			body = b.Bytes()
			w.Header().Set(ContentEncoding, "br")
		case nce == "gzip":
			b.Reset()
			_ = c.Gzip.Write(b, read(body))
			body = b.Bytes()
			w.Header().Set(ContentEncoding, "gzip")
		case nce == "deflate":
			b.Reset()
			_ = c.Flate.Write(b, read(body))
			body = b.Bytes()
			w.Header().Set(ContentEncoding, "deflate")
		}
		for k := range header {
			w.Header()[k] = header[k]
		}
		w.WriteHeader(statusCode)
		_, _ = w.Write(body)
	})
}

// NewResponseState will create ResponseState
func (httpHandler_) NewResponseState(code int, header http.Header, body []byte, compression ...string) *responseState {
	return &responseState{code, header, body, compression}
}

// SetStateToRequest will set code, header, and body to *http.Request context,
// so that the next handler that shared the same *http.Request are able
// to access it, compression parameter is optional so that this handler
// properly compress the response
//
// *responseState is created via NewResponseState
func (httpHandler_) SetStateToRequest(r *http.Request, state *responseState) {
	set(r, ctxKeyResponseState{}, state)
}

// GetStateFromRequest will get code, header, and body from *http.Request
// context set by previous handler
func (httpHandler_) GetStateFromRequest(r *http.Request) (state *responseState) {
	state, _ = r.Context().Value(ctxKeyResponseState{}).(*responseState)
	return
}

// GetNamedArgsFromRequest is a helper function that extract url.Values that have
// been parsed using MuxMatcherPattern, url.Values should not be empty if
// parsing is successful and should be able to extract further following
// url.Values, same keys in the pattern result in new value added in url.Values
func (httpHandler_) GetNamedArgsFromRequest(r *http.Request) url.Values {
	u, _ := r.Context().Value(ctxKeyNamedArgs{}).(url.Values)
	return u
}

// PanicRecoveryFromRequest is a helper function that extract error value
// when panic occured, the value is saved to *http.Request after recovery
// process and right before calling mux.PanicHandler
func (httpHandler_) PanicRecoveryFromRequest(r *http.Request) interface{} {
	return get(r, ctxKeyPanicRecovery{})
}

// ServerHijack
func (httpHandler_) Hijack(w http.ResponseWriter) (c net.Conn, rw *bufio.ReadWriter, err error) {
	err = fmt.Errorf(`http.ResponseWriter is not implementing http.Hijacker`)
	if x, ok := w.(http.Hijacker); ok && x != nil {
		c, rw, err = x.Hijack()
	}
	return
}

// ServerPush
func (httpHandler_) Push(w http.ResponseWriter, target string, opts *http.PushOptions) (err error) {
	err = fmt.Errorf(`http.ResponseWriter is not implementing http.Pusher`)
	if x, ok := w.(http.Pusher); ok && x != nil {
		err = x.Push(target, opts)
	}
	return
}

// ServerSentEvent
func (httpHandler_) SentEvent(w http.ResponseWriter, p []byte) (n int, err error) {
	err = fmt.Errorf(`http.ResponseWriter is not implementing http.Flusher`)
	if x, ok := w.(http.Flusher); ok && x != nil {
		n, err = w.Write(p)
		x.Flush()
	}
	return
}

// type httpStreamReport struct {
// 	t   Duration
// 	p   []byte
// 	n   int
// 	err error
// }

// func (httpHandler_) Stream(w http.ResponseWriter, n int) (chan<- []byte, <-chan httpStreamReport) {
// 	pub, sub := make(chan []byte, n), make(chan httpStreamReport, n)
// 	go func() {
// 		for p := range pub {
// 			t := Now()
// 			n, err := w.Write(p)
// 			sub <- httpStreamReport{time.Since(t), p, n, err}
// 		}
// 	}()
// 	return pub, sub
// }

// NegotiateContentEncoding returns the best offered content encoding for the
// request's Accept-Encoding header. If two offers match with equal weight and
// then the offer earlier in the list is preferred. If no offers are
// acceptable, then "" is returned.
func (httpHandler_) NegotiateContentEncoding(r *http.Request, offers ...string) string {
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

// Test is a convenient function with purpose of wrapping *httptest.ResponseRecorder
// and *http.Request in the same function
func (httpHandler_) Test(method, target string, body io.Reader) (*httptest.ResponseRecorder, *http.Request) {
	return httptest.NewRecorder(), httptest.NewRequest(method, target, body)
}

// TestResponse check w so that it fulfill code, header & body accordingly
func (httpHandler_) TestResponse(w *httptest.ResponseRecorder, code int, header http.Header, body []byte) bool {
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
		_, _ = fmt.Printf("Header:	\n\tExpect:	%s\n\tActual:	%s\n", header, w.Header())
	}
	if !codeEq {
		_, _ = fmt.Printf("Code:	\n\tExpect:	%d\n\tActual:	%d\n", code, w.Code)
	}
	if !bodyEq {
		_, _ = fmt.Printf("Body:	\n\tExpect:	%q\n\tActual:	%q\n", body, w.Body.Bytes())
	}
	return headerEq && codeEq && bodyEq
}

func (httpHandler_) ResponseWriter(w http.ResponseWriter) httpResponseWriterWrapper {
	return httpResponseWriterWrapper{w}
}

type httpResponseWriterWrapper struct{ http.ResponseWriter }

func (x httpResponseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if x, ok := x.ResponseWriter.(http.Hijacker); ok {
		return x.Hijack()
	}
	return nil, nil, nil
}
func (x httpResponseWriterWrapper) Flush() {
	if x, ok := x.ResponseWriter.(http.Flusher); ok {
		x.Flush()
	}
}
func (x httpResponseWriterWrapper) Push(target string, opts *http.PushOptions) error {
	if x, ok := x.ResponseWriter.(http.Pusher); ok {
		return x.Push(target, opts)
	}
	return nil
}

func (httpHandler_) Mux() *HTTPMux { return new(HTTPMux) }

// ServeHTTP implement http.Handler interface
func (m *HTTPMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var _ http.Handler = m
	code := 0
	defer func() {
		if rcv := recover(); rcv != nil {
			if m.PanicHandler == nil {
				code = http.StatusInternalServerError

				http.Error(w, http.StatusText(code), code)
				return
			}
			set(r, ctxKeyPanicRecovery{}, rcv)
			m.PanicHandler.ServeHTTP(w, r)
			return
		}
	}()
	found := false
	for _, e := range m.entries {
		if e.matcher != nil && e.next != nil {
			if found = e.matcher.Match(r); found {
				e.next.ServeHTTP(w, r)
				return
			}
		}
	}
	if !found {
		if m.NotFoundHandler == nil {
			code = http.StatusNotFound
			http.Error(w, http.StatusText(code), code)
			return
		}
		m.NotFoundHandler.ServeHTTP(w, r)
		return
	}
}

// GoString is useful for inspection by listing all its properties,
// also called when encountering verb %#v in fmt.fmt.Printf
func (m *HTTPMux) GoString() string {
	var _ fmt.GoStringer = m
	b := new(strings.Builder)
	for i := 0; i < len(m.entries); i++ {
		if i > 0 {
			_, _ = b.WriteRune(',')
		}
		_, _ = b.WriteString(m.entries[i].matcher.GoString())
	}
	return fmt.Sprintf("Mux:{Entries:[%s], NotFoundHandler:%t, PanicHandler:%t}",
		b.String(), m.NotFoundHandler != nil, m.PanicHandler != nil)
}

// With will register http.Handler with any implementation of MuxMatcher
func (m *HTTPMux) With(next http.Handler, matcher HTTPMuxMatcher) *HTTPMux {
	if next == nil || next == m || matcher == nil || !matcher.Test() {
		return m
	}
	exist := false
	for _, e := range m.entries {
		exist = exist || e.matcher.GoString() == matcher.GoString()
		be, _ := JSON.Marshal(e.matcher)
		bm, _ := JSON.Marshal(matcher)
		exist = exist || bytes.Equal(be, bm)
	}
	if !exist {
		m.entries = append(m.entries, httpMuxEntry{next, matcher})
		sort.SliceStable(m.entries, func(i, j int) bool {
			e := m.entries
			return e[i].matcher.Priority() > e[j].matcher.Priority()
		})
	}
	return m
}

// Handle will register http.Handler with MuxMatcherMethods on method and
// MuxMatcherPattern on pattern, see more details on each mux matcher
// implementation
func (m *HTTPMux) Handle(method string, pattern string, next http.Handler) *HTTPMux {
	return m.With(next, HTTP.Handler.MuxMatcher.And(0,
		HTTP.Handler.MuxMatcher.Methods(0, method),
		HTTP.Handler.MuxMatcher.Pattern(0, pattern, "", ""),
	))
}

// HTTPMux holds a map of entries
type HTTPMux struct {
	entries []httpMuxEntry

	// PanicHandler can access the error recovered via PanicRecoveryFromRequest,
	// PanicRecoveryFromRequest is a helper under httphandler package
	PanicHandler    http.Handler
	NotFoundHandler http.Handler
}

// httpMuxEntry is an element of entries listed in mux
type httpMuxEntry struct {
	next    http.Handler
	matcher HTTPMuxMatcher
}

func set(r *http.Request, key, val interface{}) {
	*r = *(r.WithContext(context.WithValue(r.Context(), key, val)))
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
type HTTPReverseProxy = httputil.ReverseProxy
type octetType byte

const (
	isToken octetType = 1 << iota
	isSpace

	// minify  = "minify"
	// br      = "br"
	// gzip    = "gzip"
	// deflate = "deflate"

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
