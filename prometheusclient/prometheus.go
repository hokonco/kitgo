package prometheusclient

import (
	"io"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/client_golang/prometheus/testutil/promlint"
)

type Config struct {
	Namespace string `yaml:"namespace" json:"namespace"`
	Subsystem string `yaml:"subsystem" json:"subsystem"`
}

func New(cfg Config) *Client {
	return &Client{cfg.Namespace, cfg.Subsystem}
}

func Test() (*Client, *Mock) {
	return New(Config{"test", "test"}), &Mock{}
}

type Client struct {
	namespace string
	subsystem string
}

func (p *Client) CounterVec(name, help string, labelNames ...string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: p.namespace, Subsystem: p.subsystem, Name: name, Help: help,
	}, labelNames)
}
func (p *Client) GaugeVec(name, help string, labelNames ...string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: p.namespace, Subsystem: p.subsystem, Name: name, Help: help,
	}, labelNames)
}
func (p *Client) HistogramVec(name, help string, labelNames ...string) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: p.namespace, Subsystem: p.subsystem, Name: name, Help: help,
	}, labelNames)
}
func (p *Client) SummaryVec(name, help string, labelNames ...string) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: p.namespace, Subsystem: p.subsystem, Name: name, Help: help,
	}, labelNames)
}

// ========================================
// MOCK
// ========================================

type Mock struct{}

func (p *Mock) ToFloat64(c prometheus.Collector) float64 {
	return testutil.ToFloat64(c)
}
func (p *Mock) CollectAndCompare(c prometheus.Collector, expected io.Reader, metricNames ...string) error {
	return testutil.CollectAndCompare(c, expected, metricNames...)
}
func (p *Mock) CollectAndCount(c prometheus.Collector, metricNames ...string) int {
	return testutil.CollectAndCount(c, metricNames...)
}
func (p *Mock) CollectAndLint(c prometheus.Collector, metricNames ...string) ([]promlint.Problem, error) {
	return testutil.CollectAndLint(c, metricNames...)
}

// ========================================
// LABEL
// ========================================

// PrometheusLabel consists of 2 string, the first is "key" and the latter is "value"
type Label [2]string

// PrometheusLabels is an adapter to build prometheus.Labels as required in *.With()
func Labels(labels ...Label) (label prometheus.Labels) {
	label = make(prometheus.Labels)
	for _, l := range labels {
		if l[0] != "" && l[1] != "" {
			label[l[0]] = l[1]
		}
	}
	return
}
