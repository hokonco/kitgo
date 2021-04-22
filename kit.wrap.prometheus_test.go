package kitgo_test

import (
	"strings"
	"testing"

	"github.com/hokonco/kitgo"
	. "github.com/onsi/gomega"
)

func Test_client_prometheus(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	test := "test"
	wrap, mock := kitgo.Prometheus.Test()
	promLabels := map[string]string{test: test}
	promCounter := wrap.CounterVec(test+"_C_total", test+"_C_total", test).With(promLabels)
	promGauge := wrap.GaugeVec(test+"_G", test+"_G", test).With(promLabels)
	promHistogram := wrap.HistogramVec(test+"_H", test+"_H", test).With(promLabels)
	promSummary := wrap.SummaryVec(test+"_S", test+"_S", test).With(promLabels)

	Expect(mock.ToFloat64(promCounter)).To(Equal(0.0))
	Expect(mock.ToFloat64(promGauge)).To(Equal(0.0))

	promCounter.Add(3.14)
	promGauge.Set(3.3)
	promHistogram.Observe(5.5)
	promSummary.Observe(7.7)

	Expect(mock.ToFloat64(promCounter)).To(Equal(3.14))
	Expect(mock.ToFloat64(promGauge)).To(Equal(3.3))

	Expect(mock.CollectAndCount(promCounter)).To(Equal(1))
	Expect(mock.CollectAndCount(promGauge)).To(Equal(1))

	Expect(mock.CollectAndCompare(promCounter, strings.NewReader("\n\n"+
		"# HELP test_C_total test_C_total\n"+
		"# TYPE test_C_total counter\n"+
		"test_C_total{test=\"test\"} 3.14\n",
	))).To(BeNil())

	lint, err := mock.CollectAndLint(promCounter)
	Expect(err).To(BeNil())
	Expect(lint).To(BeNil())
}
