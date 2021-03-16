package ristrettoclient

import (
	"github.com/dgraph-io/ristretto"
	"github.com/hokonco/kitgo"
)

type Config struct {
	NumCounters int64 `yaml:"num_counters" json:"num_counters"`
	MaxCost     int64 `yaml:"max_cost" json:"max_cost"`
	BufferItems int64 `yaml:"buffer_items" json:"buffer_items"`
	Metrics     bool  `yaml:"metrics" json:"metrics"`
	OnEvict     func(key, conflict uint64, value interface{}, cost int64)
	KeyToHash   func(key interface{}) (uint64, uint64)
	Cost        func(value interface{}) int64
}

func New(cfg Config) *Client {
	if cfg.NumCounters < 1 {
		cfg.NumCounters = 100_000_000
	}
	if cfg.MaxCost < 1 {
		cfg.MaxCost = 10_000_000
	}
	if cfg.BufferItems < 1 {
		cfg.BufferItems = 64
	}
	conf_ := ristretto.Config(cfg)
	r, err := ristretto.NewCache(&conf_)
	kitgo.PanicWhen(err != nil, err)
	return &Client{r}
}

type Client struct{ *ristretto.Cache }
