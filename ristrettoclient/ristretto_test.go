package ristrettoclient_test

import (
	"os"
	"testing"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/ristrettoclient"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_client_ristretto(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	ristrettoCli := ristrettoclient.New(ristrettoclient.Config{})
	Expect(ristrettoCli).NotTo(BeNil())
}
