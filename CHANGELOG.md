## 0.0.1

* [FEATURE] `compressclient` - a wrapper around de/compress implementation of brotli, gzip, deflate, and simple minify, Using a pool of io.Writer and io.Reader under the hood to increase performance.
* [FEATURE] `cryptoclient` - a wrapper around basic crypto, AES-GCM, RSA, ECDSA, Nonce, Argon2, BCrypt, SCrypt, and NaCl.
* [FEATURE] `dirclient` - a wrapper around file hierarchy given the path, so that user could explore the directory, using `ioutil.ReadDir` under the hood to resolve a list of `fs.FileInfo` from the current path.
* [FEATURE] `fsm` - a simple finite state machine implementation taken from https://venilnoronha.io/a-simple-state-machine-framework-in-go.
* [FEATURE] `graphicsclient` - a wrapper around image processing, `SubImage` implementation, de/encoding of JPEG,GIF,PNG,BMP,TIFF, and WEBP, and all higher level implementation of `Blur`, `Rotate`, `Scale` and `Thumbnail` generation based on https://github.com/BurntSushi/graphics-go/graphics.
* [FEATURE] `httpclient` - a wrapper of standard `net/http.Client` with additional `Test` to be mockable and a custom `Transport` that is a `net/http.RoundTripper` with additional support of `httptrace.ClientTrace`.
* [FEATURE] `httphandler` - a wrapper of standard `net/http.Handler` that are frequently used methods like `MaxBytesReader` to set a limit of `*net/http.Request.Body`, `ReverseProxy`, `Redirect`, `ServeContent`, `ServeFile` from a dir, `SetCookie` and special `ResponseWith` that implement an opinionated compression strategy given from `*net/http.Request.Header`. `httphandler` also implement `Test` and `TestResponse` to be able to test any `net/http.Handler`, and `Mux` a custom request muxer with some preloaded `MuxMatcher` like `MuxMatcherMock`,`MuxMatcherOr`,`MuxMatcherAnd`,`MuxMatcherMethods`,`MuxMatcherPattern`. `MuxMatcherPattern` is sophisticated enough to extract a matched pattern and user should be able to get the named arguments via `GetNamedArgsFromRequest`.
* [FEATURE] `httpserver` - a wrapper of standard `net/http.Server`, hiding all default methods and properties and returning a new `Listen` methods that simplify creation of http server.
* [FEATURE] `logclient` - a wrapper of standard `*log.Logger` with additional `zerolog.Logger` accessed from `Z`.
* [FEATURE] `prometheusclient` - a wrapper around prometheus vector like, `CounterVec`, `GaugeVec`, `HistogramVec`, `SummaryVec` and a `Test` using `prometheus/testutil`.
* [FEATURE] `redisclient` - a wrapper around `go-redis/redis` and a `Test` from `alicebob/miniredis`.
* [FEATURE] `ristrettoclient` - a wrapper around `dgraph-io/ristretto`, a memory bound local cache library from dgraph.
* [FEATURE] `smtpclient` - a wrapper of standard package `net/smtp`, contains a `ReverseLookup` to do an MX lookup and a `SendMail` that accepting a plain & html email format before sent, note that `kitgo.NetResolver` and `kitgo.NetDialer` is an interface so that sending email is also testable
* [FEATURE] `sqlclient` - a wrapper of standard `*database/sql.DB` extended by caching an `*sql.Stmt` before calling the query and added `Exec`, `Query`, and `QueryRow` with extra helper in `Mock`.
* [FEATURE] `templateclient` - a wrapper of standard `html/template`.
