package httpserver_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/cryptoclient"
	"github.com/hokonco/kitgo/httpserver"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_server_http(t *testing.T) {
	t.Parallel()

	t.Run("no-tls", func(t *testing.T) {
		t.Parallel()
		httpSrv := httpserver.New(httpserver.Config{
			OnShutdown: func() {},
		})
		go func() { _ = httpSrv.Listen() }()
		os.Kill.Signal()
	})
	t.Run("use-tls", func(t *testing.T) {
		t.Parallel()
		cryptoCli := cryptoclient.New()
		httpSrv := httpserver.New(httpserver.Config{
			Addr:      ":8011",
			TLSConfig: cryptoCli.NewTLSConfig("", "", t.TempDir()),
			Handler:   http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}),
		})
		go func() { _ = httpSrv.Listen() }()
		os.Kill.Signal()
		<-time.After(time.Nanosecond)
	})
}
