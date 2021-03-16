package compressclient_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/compressclient"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_client_compress(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	s := ` <p> <a href="#"> aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa </p> `
	s2 := `<p><a href=#>aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`
	compressCli := compressclient.New()
	dst := &bytes.Buffer{}
	dst2 := &bytes.Buffer{}
	src := func(s string) *bytes.Reader { return bytes.NewReader([]byte(s)) }

	dst.Reset()
	Expect(compressCli.Minify(dst, src(s), "text/html")).NotTo(HaveOccurred())
	Expect(dst.String()).To(Equal(s2))

	dst.Reset()
	dst2.Reset()
	Expect(compressCli.WriteBrotli(dst, src(s))).NotTo(HaveOccurred())
	Expect(compressCli.ReadBrotli(dst2, dst)).NotTo(HaveOccurred())
	Expect(dst2.String()).To(Equal(s))

	dst.Reset()
	dst2.Reset()
	Expect(compressCli.WriteGzip(dst, src(s))).NotTo(HaveOccurred())
	Expect(compressCli.ReadGzip(dst2, dst)).NotTo(HaveOccurred())
	Expect(dst2.String()).To(Equal(s))

	dst.Reset()
	dst2.Reset()
	Expect(compressCli.WriteDeflate(dst, src(s))).NotTo(HaveOccurred())
	Expect(compressCli.ReadDeflate(dst2, dst)).NotTo(HaveOccurred())
	Expect(dst2.String()).To(Equal(s))
}
