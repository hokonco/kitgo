package kitgo_test

import (
	"testing"

	"github.com/hokonco/kitgo"
	. "github.com/onsi/gomega"
)

func Test_client_ristretto(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	wrap := kitgo.Ristretto.New(&kitgo.RistrettoConfig{NumCounters: 1, MaxCost: 1, BufferItems: 1})
	Expect(wrap).NotTo(BeNil())
}
