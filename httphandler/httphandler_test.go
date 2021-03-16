package httphandler_test

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/httphandler"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_handler_http(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	host := "http://example.com"
	code200, body200 := 200, []byte(http.StatusText(200))
	code404, body404 := 404, []byte(http.StatusText(404)+"\n")
	code500, body500 := 500, []byte(http.StatusText(500)+"\n")
	state200 := httphandler.NewResponseState(code200, nil, body200)
	state500 := httphandler.NewResponseState(0, nil, nil)
	headerError := http.Header{}
	headerError.Set("Content-Type", "text/plain; charset=utf-8")
	headerError.Set("X-Content-Type-Options", "nosniff")

	t.Run("mux", func(t *testing.T) {
		t.Run("without-panic-handler", func(t *testing.T) {
			w, r := httphandler.Test("", host+"/", nil)
			httphandler.New().
				With(
					http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(0) }),
					httphandler.MuxMatcherMock(0, true, true)).
				ServeHTTP(w, r)
			Expect(httphandler.TestResponse(w, code500, headerError, body500)).To(BeTrue())
		})
		t.Run("without-notfound-handler", func(t *testing.T) {
			w, r := httphandler.Test("", host+"/", nil)
			httphandler.New().
				ServeHTTP(w, r)
			Expect(httphandler.TestResponse(w, code404, headerError, body404)).To(BeTrue())
		})
		t.Run("with-panic-handler", func(t *testing.T) {
			w, r := httphandler.Test("", host+"/", nil)

			mux := httphandler.New().
				With(
					http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { panic(99) }),
					httphandler.MuxMatcherMock(0, true, true)).
				With(
					http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { panic(99) }),
					httphandler.MuxMatcherMock(0, true, false))
			mux.PanicHandler = http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				Expect(httphandler.PanicRecoveryFromRequest(r)).To(Equal(99))
				httphandler.ResponseWith(state500).ServeHTTP(w, r)
			})
			_ = mux.GoString()
			mux.ServeHTTP(w, r)

			Expect(httphandler.TestResponse(w, code500, headerError, body500)).To(BeTrue())
		})
		t.Run("with-notfound-handler", func(t *testing.T) {
			w, r := httphandler.Test("", host+"/", nil)
			mux := httphandler.New()
			mux.NotFoundHandler = http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				httphandler.SetStateToRequest(r, state500)
				Expect(httphandler.GetStateFromRequest(r)).To(Equal(state500))
				httphandler.ResponseWith(state500).ServeHTTP(w, r)
			})
			mux.ServeHTTP(w, r)
			Expect(httphandler.TestResponse(w, code500, headerError, body500)).To(BeTrue())
			Expect(httphandler.TestResponse(w, code200, nil, body200)).To(BeFalse())
		})
	})
	t.Run("mock", func(t *testing.T) {
		w, r := httphandler.Test("", host+"/", nil)
		httphandler.New().
			With(httphandler.ResponseWith(state200), httphandler.MuxMatcherOr(0,
				httphandler.MuxMatcherMock(0, true, true),
				httphandler.MuxMatcherMock(0, true, true),
			)).
			With(httphandler.ResponseWith(state200), httphandler.MuxMatcherOr(0,
				httphandler.MuxMatcherMock(.1, true, true),
				httphandler.MuxMatcherMock(.2, true, true),
			)).
			ServeHTTP(w, r)
		Expect(httphandler.TestResponse(w, code200, nil, body200)).To(BeTrue())
	})
	t.Run("methods", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			w, r := httphandler.Test("", host+"/x/yyy/z", nil)
			httphandler.New().
				With(httphandler.ResponseWith(state200), httphandler.MuxMatcherOr(0,
					httphandler.MuxMatcherMethods(0),
					httphandler.MuxMatcherMethods(0, "GET"),
					httphandler.MuxMatcherMethods(0,
						"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE", "*"),
				)).
				ServeHTTP(w, r)
			Expect(httphandler.TestResponse(w, code200, nil, body200)).To(BeTrue())
		})
		t.Run("fail", func(t *testing.T) {
			w, r := httphandler.Test("", host+"/x/yyy/z", nil)
			httphandler.New().
				With(nil, nil).
				With(httphandler.ResponseWith(state200), httphandler.MuxMatcherMethods(0, "XXX")).
				With(httphandler.ResponseWith(state200), nil).
				With(httphandler.ResponseWith(state200), httphandler.MuxMatcherMethods(0, "GET")).
				ServeHTTP(w, r)
			Expect(httphandler.TestResponse(w, code200, nil, body200)).To(BeTrue())
		})
	})
	t.Run("pattern", func(t *testing.T) {
		t.Run("colon-start", func(t *testing.T) {
			w, r := httphandler.Test("", host+"/x/yyy/z", nil)
			httphandler.New().
				Handle("*", "/makan", httphandler.ResponseWith(state200)).
				Handle("*", "/:args1", httphandler.ResponseWith(state200)).
				Handle("*", "/:args1/:args2", httphandler.ResponseWith(state200)).
				Handle("GET", "/:args1/:args2/:args3", httphandler.ResponseWith(state200)).
				ServeHTTP(w, r)
			Expect(httphandler.TestResponse(w, code200, nil, body200)).To(BeTrue())
			Expect(httphandler.GetNamedArgsFromRequest(r).Get("args1")).To(Equal("x"))
			Expect(httphandler.GetNamedArgsFromRequest(r).Get("args2")).To(Equal("yyy"))
			Expect(httphandler.GetNamedArgsFromRequest(r).Get("args3")).To(Equal("z"))
		})
		t.Run("colon-both", func(t *testing.T) {
			w, r := httphandler.Test("", host+"/x/yyy/z", nil)
			httphandler.New().
				With(httphandler.ResponseWith(state200),
					httphandler.MuxMatcherPattern(0, "/:args1:/:args2:/:args3:", ":", ":")).
				ServeHTTP(w, r)
			Expect(httphandler.TestResponse(w, code200, nil, body200)).To(BeTrue())
			Expect(httphandler.GetNamedArgsFromRequest(r).Get("args1")).To(Equal("x"))
			Expect(httphandler.GetNamedArgsFromRequest(r).Get("args2")).To(Equal("yyy"))
			Expect(httphandler.GetNamedArgsFromRequest(r).Get("args3")).To(Equal("z"))
		})
		t.Run("curly-braces", func(t *testing.T) {
			w, r := httphandler.Test("", host+"/x/yyy/z", nil)
			httphandler.New().
				With(httphandler.ResponseWith(state200),
					httphandler.MuxMatcherPattern(0, "/{{args1}}/{{args2}}/{{args3}}", "{{", "}}")).
				ServeHTTP(w, r)
			Expect(httphandler.TestResponse(w, code200, nil, body200)).To(BeTrue())
			Expect(httphandler.GetNamedArgsFromRequest(r).Get("args1")).To(Equal("x"))
			Expect(httphandler.GetNamedArgsFromRequest(r).Get("args2")).To(Equal("yyy"))
			Expect(httphandler.GetNamedArgsFromRequest(r).Get("args3")).To(Equal("z"))
		})
		t.Run("exact-no-pattern", func(t *testing.T) {
			w, r := httphandler.Test("", host+"/x/yyy/z", nil)
			httphandler.New().
				With(httphandler.ResponseWith(state200),
					httphandler.MuxMatcherPattern(0, "/x/yyy/z", "", "")).
				ServeHTTP(w, r)
			Expect(httphandler.TestResponse(w, code200, nil, body200)).To(BeTrue())
			// Expect(httphandler.NamedArgsFromRequest(r).Get("args1")).To(Equal("x"))
			// Expect(httphandler.NamedArgsFromRequest(r).Get("args2")).To(Equal("yyy"))
			// Expect(httphandler.NamedArgsFromRequest(r).Get("args3")).To(Equal("z"))
		})
	})
	t.Run("handler", func(t *testing.T) {
		t.Run("test response", func(t *testing.T) {
			w, r := httphandler.Test("", host+"/", nil)
			Expect(w).NotTo(BeNil())
			Expect(r).NotTo(BeNil())
			httphandler.New().ServeHTTP(w, r)
			Expect(httphandler.TestResponse(w, code404, headerError, body404)).To(BeTrue())
		})
		t.Run("NegotiateContentEncoding", func(t *testing.T) {
			state := httphandler.NewResponseState(code200, nil, body200, "br", "deflate", "gzip", "minify")
			w, r := httphandler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "br")
			httphandler.ResponseWith(state).ServeHTTP(w, r)

			w, r = httphandler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "deflate")
			httphandler.ResponseWith(state).ServeHTTP(w, r)

			w, r = httphandler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "gzip")
			httphandler.ResponseWith(state).ServeHTTP(w, r)

			w, r = httphandler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "gzip;q=0, deflate;q=-0.7, *;q=0")
			httphandler.ResponseWith(state).ServeHTTP(w, r)

			w, r = httphandler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "gzip;q=1.0, br;qq=0.5, *;q=0.5")
			httphandler.ResponseWith(state).ServeHTTP(w, r)

			w, r = httphandler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "gzip;q=0.1, br;q=, *;q=")
			httphandler.ResponseWith(state).ServeHTTP(w, r)

			w, r = httphandler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "")
			httphandler.ResponseWith(state).ServeHTTP(w, r)
		})
		t.Run("misc", func(t *testing.T) {
			w, r := httphandler.Test("", host+"/", nil)
			tw := httphandler.TestResponseWriter{w}
			Expect(w).NotTo(BeNil())
			Expect(r).NotTo(BeNil())
			_, _, _ = httphandler.Hijack(tw)
			_, _ = httphandler.SentEvent(tw, nil)
			_ = httphandler.Push(tw, "", nil)

			httphandler.MaxBytesReader(1).ServeHTTP(tw, r)
			httphandler.SetCookie(nil).ServeHTTP(tw, r)
			httphandler.Redirect(301, "").ServeHTTP(tw, r)
			httphandler.ServeContent("", time.Time{}, strings.NewReader("")).ServeHTTP(tw, r)
			httphandler.ServeFile("").ServeHTTP(tw, r)
			httphandler.ServeReverseProxy("", func(*httphandler.ReverseProxy) {}).ServeHTTP(tw, r)
		})
	})
}
