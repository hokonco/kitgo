package redisclient

import (
	"crypto/tls"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/hokonco/kitgo"
)

type Config struct {
	// Either a single address or a seed list of host:port addresses
	// of cluster/sentinel nodes.
	Addrs []string `yaml:"addrs" json:"addrs"`

	// Database to be selected after connecting to the server.
	// Only single-node and failover clients.
	DB int `yaml:"db" json:"db"`

	// Common options.

	Password           string `yaml:"password" json:"password"`
	MaxRetries         int    `yaml:"max_retries" json:"max_retries"`
	MinRetryBackoff    string `yaml:"min_retry_backoff" json:"min_retry_backoff"`
	MaxRetryBackoff    string `yaml:"max_retry_backoff" json:"max_retry_backoff"`
	DialTimeout        string `yaml:"dial_timeout" json:"dial_timeout"`
	ReadTimeout        string `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout       string `yaml:"write_timeout" json:"write_timeout"`
	PoolSize           int    `yaml:"pool_size" json:"pool_size"`
	MinIdleConns       int    `yaml:"min_idle_conns" json:"min_idle_conns"`
	MaxConnAge         string `yaml:"max_conn_age" json:"max_conn_age"`
	PoolTimeout        string `yaml:"pool_timeout" json:"pool_timeout"`
	IdleTimeout        string `yaml:"idle_timeout" json:"idle_timeout"`
	IdleCheckFrequency string `yaml:"idle_check_frequency" json:"idle_check_frequency"`

	// Only cluster clients.

	MaxRedirects   int  `yaml:"max_redirects" json:"max_redirects"`
	ReadOnly       bool `yaml:"read_only" json:"read_only"`
	RouteByLatency bool `yaml:"route_by_latency" json:"route_by_latency"`
	RouteRandomly  bool `yaml:"route_randomly" json:"route_randomly"`

	// The sentinel master name.
	// Only failover clients.
	MasterName string `yaml:"master_name" json:"master_name"`

	OnConnect func(*redis.Conn) error
	TLSConfig *tls.Config
}

func New(cfg Config) *Client {
	o := &redis.UniversalOptions{}
	o.Addrs = cfg.Addrs
	o.DB = cfg.DB
	o.Password = cfg.Password
	o.MaxRetries = cfg.MaxRetries
	o.MinRetryBackoff = kitgo.ParseDuration(cfg.MinRetryBackoff, 5*time.Second)
	o.MaxRetryBackoff = kitgo.ParseDuration(cfg.MaxRetryBackoff, 30*time.Second)
	o.DialTimeout = kitgo.ParseDuration(cfg.DialTimeout, 30*time.Second)
	o.ReadTimeout = kitgo.ParseDuration(cfg.ReadTimeout, 30*time.Second)
	o.WriteTimeout = kitgo.ParseDuration(cfg.WriteTimeout, 30*time.Second)
	o.PoolSize = cfg.PoolSize
	o.MinIdleConns = cfg.MinIdleConns
	o.MaxConnAge = kitgo.ParseDuration(cfg.MaxConnAge, time.Minute)
	o.PoolTimeout = kitgo.ParseDuration(cfg.PoolTimeout, time.Minute)
	o.IdleTimeout = kitgo.ParseDuration(cfg.IdleTimeout, time.Minute)
	o.IdleCheckFrequency = kitgo.ParseDuration(cfg.IdleCheckFrequency, time.Minute)
	o.MaxRedirects = cfg.MaxRedirects
	o.ReadOnly = cfg.ReadOnly
	o.RouteByLatency = cfg.RouteByLatency
	o.RouteRandomly = cfg.RouteRandomly
	o.MasterName = cfg.MasterName

	o.OnConnect = cfg.OnConnect
	o.TLSConfig = cfg.TLSConfig

	return &Client{redis.NewUniversalClient(o)}
}

func Test() (client *Client, mock *miniredis.Miniredis) {
	m, err := miniredis.Run()
	kitgo.PanicWhen(err != nil, err)
	return New(Config{Addrs: []string{m.Addr()}}), m
}

type Client struct{ redis.UniversalClient }
