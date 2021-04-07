package httpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/hokonco/kitgo"
)

type Config struct {
	Addr              string       `yaml:"addr" json:"addr"`
	Handler           http.Handler `yaml:"-" json:"-"`
	TLSConfig         *tls.Config  `yaml:"-" json:"-"`
	ReadTimeout       string       `yaml:"read_timeout" json:"read_timeout"`
	ReadHeaderTimeout string       `yaml:"read_header_timeout" json:"read_header_timeout"`
	WriteTimeout      string       `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout       string       `yaml:"idle_timeout" json:"idle_timeout"`
	MaxHeaderBytes    int          `yaml:"max_header_bytes" json:"max_header_bytes"`

	TLSNextProto map[string]func(*http.Server, *tls.Conn, http.Handler) `yaml:"-" json:"-"`
	ConnState    func(net.Conn, http.ConnState)                         `yaml:"-" json:"-"`
	ErrorLog     *log.Logger                                            `yaml:"-" json:"-"`
	BaseContext  func(net.Listener) context.Context                     `yaml:"-" json:"-"`
	ConnContext  func(ctx context.Context, c net.Conn) context.Context  `yaml:"-" json:"-"`

	ShutdownTimeout string        `yaml:"shutdown_timeout" json:"shutdown_timeout"`
	shutdownTimeout time.Duration `yaml:"-" json:"-"`
	OnShutdown      func()        `yaml:"-" json:"-"`
	OnInfo          func(string)  `yaml:"-" json:"-"`
	OnError         func(error)   `yaml:"-" json:"-"`
}
type Server struct {
	srv *http.Server
	mu  sync.Mutex
	cfg Config
}

func New(cfg Config) *Server {
	srv := &http.Server{}
	srv.Addr = cfg.Addr
	srv.Handler = cfg.Handler
	srv.TLSConfig = cfg.TLSConfig
	srv.ReadTimeout = kitgo.ParseDuration(cfg.ReadTimeout, 3*time.Second)
	srv.ReadHeaderTimeout = kitgo.ParseDuration(cfg.ReadHeaderTimeout, time.Second)
	srv.WriteTimeout = kitgo.ParseDuration(cfg.WriteTimeout, 3*time.Second)
	srv.IdleTimeout = kitgo.ParseDuration(cfg.IdleTimeout, time.Minute)
	srv.MaxHeaderBytes = cfg.MaxHeaderBytes
	srv.TLSNextProto = cfg.TLSNextProto
	srv.ConnState = cfg.ConnState
	srv.BaseContext = cfg.BaseContext
	srv.ConnContext = cfg.ConnContext
	srv.ErrorLog = cfg.ErrorLog
	if cfg.OnShutdown != nil {
		srv.RegisterOnShutdown(cfg.OnShutdown)
	}
	if cfg.OnInfo == nil {
		cfg.OnInfo = func(string) {}
	}
	if cfg.OnError == nil {
		cfg.OnError = func(error) {}
	}
	cfg.shutdownTimeout = kitgo.ParseDuration(cfg.ShutdownTimeout, 5*time.Second)
	return &Server{srv: srv, cfg: cfg}
}
func (x *Server) Listen() error {
	x.mu.Lock()
	defer x.mu.Unlock()

	// ========================================
	// Listen
	// ========================================
	errChan := make(chan error, 1)
	go func() {
		// --> on receiving signal, return error
		sig := kitgo.ListenToSignal(
			syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT,
			syscall.SIGQUIT, syscall.SIGSTOP, syscall.SIGTSTP,
		)
		errChan <- fmt.Errorf("signal: [%d] %s\n%s", sig, sig, debug.Stack())
	}()
	if x.srv.TLSConfig == nil {
		go func() { errChan <- x.srv.ListenAndServe() }() // --> on srv error
		x.cfg.OnInfo(fmt.Sprintf("srv running on http://0.0.0.0%v", x.srv.Addr))
	} else {
		go func() { errChan <- x.srv.ListenAndServeTLS("", "") }() // --> on tls error
		x.cfg.OnInfo(fmt.Sprintf("tls running on https://0.0.0.0%v", x.srv.Addr))
	}
	x.cfg.OnError(<-errChan)

	// ========================================
	// Shutdown
	// ========================================
	ctx, cancel := context.WithTimeout(context.Background(), x.cfg.shutdownTimeout)
	defer cancel()
	err := x.srv.Shutdown(ctx)
	x.cfg.OnError(err)
	return err
}
