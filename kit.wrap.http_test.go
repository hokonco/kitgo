package kitgo_test

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/hokonco/kitgo"
	. "github.com/onsi/gomega"
)

func Test_client_http(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	wrap, mock := kitgo.HTTP.Client.Test()
	defer mock.Close()
	defer mock.CloseClientConnections()

	mock.Expect("GET", "/404", http.NotFoundHandler())
	req, _ := http.NewRequest("GET", mock.URL("/404"), nil)
	err_ := fmt.Errorf("error")
	Expect(req).NotTo(BeNil())

	t.Run("err RequestFn", func(t *testing.T) {
		wrap.Transport = kitgo.HTTP.Transport.New().
			OnRequest(func(r *http.Request) error { return err_ })
		res, err := wrap.Do(req)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("Get \"" + mock.URL("/404") + "\": " + err_.Error()))
		Expect(res).To(BeNil())
	})

	t.Run("err ResponseFn", func(t *testing.T) {
		wrap.Transport = kitgo.HTTP.Transport.New().
			OnRequest(func(r *http.Request) error { return nil }).
			OnResponse(func(r *http.Response) error { return err_ })
		res, err := wrap.Do(req)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("Get \"" + mock.URL("/404") + "\": " + err_.Error()))
		Expect(res).To(BeNil())
	})

	t.Run("err RoundTrip", func(t *testing.T) {
		_req, _ := http.NewRequest("GET", mock.URL("::"), nil)
		Expect(_req).NotTo(BeNil())
		wrap.Transport = kitgo.HTTP.Transport.New()
		res, err := wrap.Do(_req)
		Expect(err).NotTo(BeNil())
		Expect(res).To(BeNil())
	})

	t.Run("ok", func(t *testing.T) {
		bReq, bRes := new(bytes.Buffer), new(bytes.Buffer)
		wrap.Transport = kitgo.HTTP.Transport.New().
			WithBase(nil).
			WithTrace(&kitgo.HTTPClientTrace{}).
			DumpOnRequest(bReq, true).
			DumpOnResponse(bRes, true)
		res, err := wrap.Do(req)
		Expect(err).To(BeNil())
		Expect(res).NotTo(BeNil())
	})
}

func Test_server_http(t *testing.T) {
	t.Parallel()

	t.Run("no-tls", func(t *testing.T) {
		t.Parallel()
		httpSrv := kitgo.HTTP.Server.New()
		go func() { _ = httpSrv.Run(nil, nil, 0) }()
		syscall.SIGKILL.Signal()
	})
	t.Run("use-tls", func(t *testing.T) {
		t.Parallel()
		httpSrv := kitgo.HTTP.Server.New()
		httpSrv.Addr = ":8011"
		httpSrv.TLSConfig = kitgo.Crypto.New().NewTLSConfig("", "", t.TempDir())
		httpSrv.Handler = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
		go func() { _ = httpSrv.Run(nil, nil, 0) }()
		syscall.SIGKILL.Signal()
		<-time.After(time.Nanosecond)
	})
}

func Test_handler_http(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	host := "http://example.com"
	code200, body200 := 200, []byte(http.StatusText(200))
	code404, body404 := 404, []byte(http.StatusText(404)+"\n")
	code500, body500 := 500, []byte(http.StatusText(500)+"\n")
	state200 := kitgo.HTTP.Handler.NewResponseState(code200, nil, body200)
	state500 := kitgo.HTTP.Handler.NewResponseState(0, nil, nil)
	headerError := http.Header{}
	headerError.Set("Content-Type", "text/plain; charset=utf-8")
	headerError.Set("X-Content-Type-Options", "nosniff")

	t.Run("mux", func(t *testing.T) {
		t.Run("without-panic-handler", func(t *testing.T) {
			w, r := kitgo.HTTP.Handler.Test("", host+"/", nil)
			kitgo.HTTP.Handler.Mux().
				With(
					http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(0) }),
					kitgo.HTTP.Handler.MuxMatcher.Mock(0, true, true)).
				ServeHTTP(w, r)
			Expect(kitgo.HTTP.Handler.TestResponse(w, code500, headerError, body500)).To(BeTrue())
		})
		t.Run("without-notfound-handler", func(t *testing.T) {
			w, r := kitgo.HTTP.Handler.Test("", host+"/", nil)
			kitgo.HTTP.Handler.Mux().
				ServeHTTP(w, r)
			Expect(kitgo.HTTP.Handler.TestResponse(w, code404, headerError, body404)).To(BeTrue())
		})
		t.Run("with-panic-handler", func(t *testing.T) {
			w, r := kitgo.HTTP.Handler.Test("", host+"/", nil)

			mux := kitgo.HTTP.Handler.Mux().
				With(
					http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { panic(99) }),
					kitgo.HTTP.Handler.MuxMatcher.Mock(0, true, true)).
				With(
					http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { panic(99) }),
					kitgo.HTTP.Handler.MuxMatcher.Mock(0, true, false))
			mux.PanicHandler = http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				Expect(kitgo.HTTP.Handler.PanicRecoveryFromRequest(r)).To(Equal(99))
				kitgo.HTTP.Handler.ResponseWith(state500).ServeHTTP(w, r)
			})
			_ = mux.GoString()
			mux.ServeHTTP(w, r)

			Expect(kitgo.HTTP.Handler.TestResponse(w, code500, headerError, body500)).To(BeTrue())
		})
		t.Run("with-notfound-handler", func(t *testing.T) {
			w, r := kitgo.HTTP.Handler.Test("", host+"/", nil)
			mux := kitgo.HTTP.Handler.Mux()
			mux.NotFoundHandler = http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				kitgo.HTTP.Handler.SetStateToRequest(r, state500)
				Expect(kitgo.HTTP.Handler.GetStateFromRequest(r)).To(Equal(state500))
				kitgo.HTTP.Handler.ResponseWith(state500).ServeHTTP(w, r)
			})
			mux.ServeHTTP(w, r)
			Expect(kitgo.HTTP.Handler.TestResponse(w, code500, headerError, body500)).To(BeTrue())
			Expect(kitgo.HTTP.Handler.TestResponse(w, code200, nil, body200)).To(BeFalse())
		})
	})
	t.Run("mock", func(t *testing.T) {
		w, r := kitgo.HTTP.Handler.Test("", host+"/", nil)
		kitgo.HTTP.Handler.Mux().
			With(kitgo.HTTP.Handler.ResponseWith(state200), kitgo.HTTP.Handler.MuxMatcher.Or(0,
				kitgo.HTTP.Handler.MuxMatcher.Mock(0, true, true),
				kitgo.HTTP.Handler.MuxMatcher.Mock(0, true, true),
			)).
			With(kitgo.HTTP.Handler.ResponseWith(state200), kitgo.HTTP.Handler.MuxMatcher.Or(0,
				kitgo.HTTP.Handler.MuxMatcher.Mock(.1, true, true),
				kitgo.HTTP.Handler.MuxMatcher.Mock(.2, true, true),
			)).
			ServeHTTP(w, r)
		Expect(kitgo.HTTP.Handler.TestResponse(w, code200, nil, body200)).To(BeTrue())
	})
	t.Run("methods", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			w, r := kitgo.HTTP.Handler.Test("", host+"/x/yyy/z", nil)
			kitgo.HTTP.Handler.Mux().
				With(kitgo.HTTP.Handler.ResponseWith(state200), kitgo.HTTP.Handler.MuxMatcher.Or(0,
					kitgo.HTTP.Handler.MuxMatcher.Methods(0),
					kitgo.HTTP.Handler.MuxMatcher.Methods(0, "GET"),
					kitgo.HTTP.Handler.MuxMatcher.Methods(0,
						"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE", "*"),
				)).
				ServeHTTP(w, r)
			Expect(kitgo.HTTP.Handler.TestResponse(w, code200, nil, body200)).To(BeTrue())
		})
		t.Run("fail", func(t *testing.T) {
			w, r := kitgo.HTTP.Handler.Test("", host+"/x/yyy/z", nil)
			kitgo.HTTP.Handler.Mux().
				With(nil, nil).
				With(kitgo.HTTP.Handler.ResponseWith(state200), kitgo.HTTP.Handler.MuxMatcher.Methods(0, "XXX")).
				With(kitgo.HTTP.Handler.ResponseWith(state200), nil).
				With(kitgo.HTTP.Handler.ResponseWith(state200), kitgo.HTTP.Handler.MuxMatcher.Methods(0, "GET")).
				ServeHTTP(w, r)
			Expect(kitgo.HTTP.Handler.TestResponse(w, code200, nil, body200)).To(BeTrue())
		})
	})
	t.Run("pattern", func(t *testing.T) {
		t.Run("colon-start", func(t *testing.T) {
			w, r := kitgo.HTTP.Handler.Test("", host+"/x/yyy/z", nil)
			kitgo.HTTP.Handler.Mux().
				Handle("*", "/makan", kitgo.HTTP.Handler.ResponseWith(state200)).
				Handle("*", "/:args1", kitgo.HTTP.Handler.ResponseWith(state200)).
				Handle("*", "/:args1/:args2", kitgo.HTTP.Handler.ResponseWith(state200)).
				Handle("GET", "/:args1/:args2/:args3", kitgo.HTTP.Handler.ResponseWith(state200)).
				ServeHTTP(w, r)
			Expect(kitgo.HTTP.Handler.TestResponse(w, code200, nil, body200)).To(BeTrue())
			Expect(kitgo.HTTP.Handler.GetNamedArgsFromRequest(r).Get("args1")).To(Equal("x"))
			Expect(kitgo.HTTP.Handler.GetNamedArgsFromRequest(r).Get("args2")).To(Equal("yyy"))
			Expect(kitgo.HTTP.Handler.GetNamedArgsFromRequest(r).Get("args3")).To(Equal("z"))
		})
		t.Run("colon-both", func(t *testing.T) {
			w, r := kitgo.HTTP.Handler.Test("", host+"/x/yyy/z", nil)
			kitgo.HTTP.Handler.Mux().
				With(kitgo.HTTP.Handler.ResponseWith(state200),
					kitgo.HTTP.Handler.MuxMatcher.Pattern(0, "/:args1:/:args2:/:args3:", ":", ":")).
				ServeHTTP(w, r)
			Expect(kitgo.HTTP.Handler.TestResponse(w, code200, nil, body200)).To(BeTrue())
			Expect(kitgo.HTTP.Handler.GetNamedArgsFromRequest(r).Get("args1")).To(Equal("x"))
			Expect(kitgo.HTTP.Handler.GetNamedArgsFromRequest(r).Get("args2")).To(Equal("yyy"))
			Expect(kitgo.HTTP.Handler.GetNamedArgsFromRequest(r).Get("args3")).To(Equal("z"))
		})
		t.Run("curly-braces", func(t *testing.T) {
			w, r := kitgo.HTTP.Handler.Test("", host+"/x/yyy/z", nil)
			kitgo.HTTP.Handler.Mux().
				With(kitgo.HTTP.Handler.ResponseWith(state200),
					kitgo.HTTP.Handler.MuxMatcher.Pattern(0, "/{{args1}}/{{args2}}/{{args3}}", "{{", "}}")).
				ServeHTTP(w, r)
			Expect(kitgo.HTTP.Handler.TestResponse(w, code200, nil, body200)).To(BeTrue())
			Expect(kitgo.HTTP.Handler.GetNamedArgsFromRequest(r).Get("args1")).To(Equal("x"))
			Expect(kitgo.HTTP.Handler.GetNamedArgsFromRequest(r).Get("args2")).To(Equal("yyy"))
			Expect(kitgo.HTTP.Handler.GetNamedArgsFromRequest(r).Get("args3")).To(Equal("z"))
		})
		t.Run("exact-no-pattern", func(t *testing.T) {
			w, r := kitgo.HTTP.Handler.Test("", host+"/x/yyy/z", nil)
			kitgo.HTTP.Handler.Mux().
				With(kitgo.HTTP.Handler.ResponseWith(state200),
					kitgo.HTTP.Handler.MuxMatcher.Pattern(0, "/x/yyy/z", "", "")).
				ServeHTTP(w, r)
			Expect(kitgo.HTTP.Handler.TestResponse(w, code200, nil, body200)).To(BeTrue())
			// Expect(kitgo.HTTP.Handler.NamedArgsFromRequest(r).Get("args1")).To(Equal("x"))
			// Expect(kitgo.HTTP.Handler.NamedArgsFromRequest(r).Get("args2")).To(Equal("yyy"))
			// Expect(kitgo.HTTP.Handler.NamedArgsFromRequest(r).Get("args3")).To(Equal("z"))
		})
	})
	t.Run("handler", func(t *testing.T) {
		t.Run("test response", func(t *testing.T) {
			w, r := kitgo.HTTP.Handler.Test("", host+"/", nil)
			Expect(w).NotTo(BeNil())
			Expect(r).NotTo(BeNil())
			kitgo.HTTP.Handler.Mux().ServeHTTP(w, r)
			Expect(kitgo.HTTP.Handler.TestResponse(w, code404, headerError, body404)).To(BeTrue())
		})
		t.Run("NegotiateContentEncoding", func(t *testing.T) {
			state := kitgo.HTTP.Handler.NewResponseState(code200, nil, body200, "br", "deflate", "gzip", "minify")
			w, r := kitgo.HTTP.Handler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "br")
			kitgo.HTTP.Handler.ResponseWith(state).ServeHTTP(w, r)

			w, r = kitgo.HTTP.Handler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "deflate")
			kitgo.HTTP.Handler.ResponseWith(state).ServeHTTP(w, r)

			w, r = kitgo.HTTP.Handler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "gzip")
			kitgo.HTTP.Handler.ResponseWith(state).ServeHTTP(w, r)

			w, r = kitgo.HTTP.Handler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "gzip;q=0, deflate;q=-0.7, *;q=0")
			kitgo.HTTP.Handler.ResponseWith(state).ServeHTTP(w, r)

			w, r = kitgo.HTTP.Handler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "gzip;q=1.0, br;qq=0.5, *;q=0.5")
			kitgo.HTTP.Handler.ResponseWith(state).ServeHTTP(w, r)

			w, r = kitgo.HTTP.Handler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "gzip;q=0.1, br;q=, *;q=")
			kitgo.HTTP.Handler.ResponseWith(state).ServeHTTP(w, r)

			w, r = kitgo.HTTP.Handler.Test("", host+"/", nil)
			r.Header.Add("Accept-Encoding", "")
			kitgo.HTTP.Handler.ResponseWith(state).ServeHTTP(w, r)
		})
		t.Run("misc", func(t *testing.T) {
			w, r := kitgo.HTTP.Handler.Test("", host+"/", nil)
			tw := kitgo.HTTP.Handler.ResponseWriter(w)
			tw = kitgo.HTTP.Handler.ResponseWriter(tw)
			Expect(w).NotTo(BeNil())
			Expect(r).NotTo(BeNil())
			_, _, _ = kitgo.HTTP.Handler.Hijack(tw)
			_, _ = kitgo.HTTP.Handler.SentEvent(tw, nil)
			_ = kitgo.HTTP.Handler.Push(tw, "", nil)

			kitgo.HTTP.Handler.MaxBytesReader(1).ServeHTTP(tw, r)
			kitgo.HTTP.Handler.SetCookie(nil).ServeHTTP(tw, r)
			kitgo.HTTP.Handler.Redirect(301, "").ServeHTTP(tw, r)
			kitgo.HTTP.Handler.ServeContent("", time.Time{}, strings.NewReader("")).ServeHTTP(tw, r)
			kitgo.HTTP.Handler.ServeFile("").ServeHTTP(tw, r)
			kitgo.HTTP.Handler.ServeReverseProxy("", func(*kitgo.HTTPReverseProxy) {}).ServeHTTP(tw, r)
		})
	})
}
