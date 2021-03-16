package prometheusclient_test

import (
	"os"
	"strings"
	"testing"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/prometheusclient"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_client_prometheus(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	promCli, mock := prometheusclient.Test()
	test := "test"
	promLabels := prometheusclient.Labels(prometheusclient.Label{test, test})
	promCounter := promCli.CounterVec(test+"_C_total", test+"_C_total", test).With(promLabels)
	promGauge := promCli.GaugeVec(test+"_G", test+"_G", test).With(promLabels)
	promHistogram := promCli.HistogramVec(test+"_H", test+"_H", test).With(promLabels)
	promSummary := promCli.SummaryVec(test+"_S", test+"_S", test).With(promLabels)

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

	Expect(mock.CollectAndCompare(promCounter, strings.NewReader(`
		# HELP test_test_test_C_total test_C_total
		# TYPE test_test_test_C_total counter
		test_test_test_C_total{test="test"} 3.14
	`))).To(BeNil())

	lint, err := mock.CollectAndLint(promCounter)
	Expect(err).To(BeNil())
	Expect(lint).To(BeNil())
}
