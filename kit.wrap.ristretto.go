package kitgo

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

var Ristretto ristretto_

type ristretto_ struct{}

func (ristretto_) New(conf *RistrettoConfig) *RistrettoWrapper {
	x, err := ristretto.NewCache(conf)
	var _ RistrettoCacheI = x
	var _ RistrettoCacheMetricsI = x.Metrics
	PanicWhen(err != nil || x == nil || x.Metrics == nil, err)
	return &RistrettoWrapper{x}
}

type RistrettoConfig = ristretto.Config
type RistrettoWrapper struct{ *ristretto.Cache }
type RistrettoCacheI interface {
	Clear()
	Close()
	Del(key interface{})
	Get(key interface{}) (interface{}, bool)
	Set(key, value interface{}, cost int64) bool
	SetWithTTL(key, value interface{}, cost int64, ttl time.Duration) bool
}
type RistrettoCacheMetricsI interface {
	Clear()
	CostAdded() uint64
	CostEvicted() uint64
	GetsDropped() uint64
	GetsKept() uint64
	Hits() uint64
	KeysAdded() uint64
	KeysEvicted() uint64
	KeysUpdated() uint64
	Misses() uint64
	Ratio() float64
	SetsDropped() uint64
	SetsRejected() uint64
	String() string
}
