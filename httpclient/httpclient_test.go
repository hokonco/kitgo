package httpclient_test

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/httpclient"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_client_http(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	httpCli, mock := httpclient.Test()
	defer mock.Close()
	defer mock.CloseClientConnections()

	mock.Expect("GET", "/404", http.NotFoundHandler())
	req, err := http.NewRequest("GET", mock.URL("/404"), nil)
	err_ := fmt.Errorf("error")
	Expect(err).To(BeNil())
	Expect(req).NotTo(BeNil())

	t.Run("err RequestFn", func(t *testing.T) {
		httpCli.Transport = &httpclient.Transport{
			OnRequest: func(r *http.Request) error { return err_ },
		}
		res, err := httpCli.Do(req)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("Get \"" + mock.URL("/404") + "\": " + err_.Error()))
		Expect(res).To(BeNil())
	})

	t.Run("err ResponseFn", func(t *testing.T) {
		httpCli.Transport = &httpclient.Transport{
			OnRequest:  func(r *http.Request) error { return nil },
			OnResponse: func(r *http.Response) error { return err_ },
		}
		res, err := httpCli.Do(req)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("Get \"" + mock.URL("/404") + "\": " + err_.Error()))
		Expect(res).To(BeNil())
	})

	t.Run("err RoundTrip", func(t *testing.T) {
		_req, err := http.NewRequest("GET", mock.URL("::"), nil)
		Expect(err).To(BeNil())
		Expect(_req).NotTo(BeNil())
		httpCli.Transport = &httpclient.Transport{}
		res, err := httpCli.Do(_req)
		Expect(err).NotTo(BeNil())
		Expect(res).To(BeNil())
	})

	t.Run("ok", func(t *testing.T) {
		bReq, bRes := &bytes.Buffer{}, &bytes.Buffer{}
		httpCli.Transport = &httpclient.Transport{
			Trace:      httpclient.NewTrace(func(t *httpclient.Trace) {}),
			OnRequest:  httpclient.DumpOnRequest(bReq, true),
			OnResponse: httpclient.DumpOnResponse(bRes, true),
		}
		res, err := httpCli.Do(req)
		Expect(err).To(BeNil())
		Expect(res).NotTo(BeNil())
	})
}
