package kitgo

import (
	"io"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/client_golang/prometheus/testutil/promlint"
)

var Prometheus prometheus_

type prometheus_ struct{}

func (prometheus_) New(conf *PrometheusConfig) *PrometheusWrapper {
	PanicWhen(conf == nil, conf)
	return &PrometheusWrapper{conf}
}

func (prometheus_) Test() (*PrometheusWrapper, *PrometheusMock) {
	return Prometheus.New(&PrometheusConfig{}), &PrometheusMock{}
}

type PrometheusConfig struct {
	Namespace string
	Subsystem string
}
type PrometheusWrapper struct{ conf *PrometheusConfig }

func (p *PrometheusWrapper) CounterVec(name, help string, labelNames ...string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: p.conf.Namespace, Subsystem: p.conf.Subsystem, Name: name, Help: help,
	}, labelNames)
}
func (p *PrometheusWrapper) GaugeVec(name, help string, labelNames ...string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: p.conf.Namespace, Subsystem: p.conf.Subsystem, Name: name, Help: help,
	}, labelNames)
}
func (p *PrometheusWrapper) HistogramVec(name, help string, labelNames ...string) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: p.conf.Namespace, Subsystem: p.conf.Subsystem, Name: name, Help: help,
	}, labelNames)
}
func (p *PrometheusWrapper) SummaryVec(name, help string, labelNames ...string) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: p.conf.Namespace, Subsystem: p.conf.Subsystem, Name: name, Help: help,
	}, labelNames)
}

// ========================================
// MOCK
// ========================================

type PrometheusMock struct{}

func (p *PrometheusMock) ToFloat64(c prometheus.Collector) float64 {
	return testutil.ToFloat64(c)
}
func (p *PrometheusMock) CollectAndCompare(c prometheus.Collector, expected io.Reader, metricNames ...string) error {
	return testutil.CollectAndCompare(c, expected, metricNames...)
}
func (p *PrometheusMock) CollectAndCount(c prometheus.Collector, metricNames ...string) int {
	return testutil.CollectAndCount(c, metricNames...)
}
func (p *PrometheusMock) CollectAndLint(c prometheus.Collector, metricNames ...string) ([]promlint.Problem, error) {
	return testutil.CollectAndLint(c, metricNames...)
}
