package kitgo_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hokonco/kitgo"
	. "github.com/onsi/gomega"
)

func Test_client_compress(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	const s = ` <p> <a href="#"> aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa </p> `
	const s2 = `<p><a href=#>aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`
	wrap := kitgo.Compress.New()
	dst := new(bytes.Buffer)
	dst2 := new(bytes.Buffer)

	dst.Reset()
	Expect(wrap.Minify.WithMediaType("text/html").Write(dst, strings.NewReader(s))).NotTo(HaveOccurred())
	Expect(dst.String()).To(Equal(s2))

	dst.Reset()
	dst2.Reset()
	Expect(wrap.Brotli.Write(dst, strings.NewReader(s))).NotTo(HaveOccurred())
	Expect(wrap.Brotli.Read(dst2, dst)).NotTo(HaveOccurred())
	Expect(dst2.String()).To(Equal(s))

	dst.Reset()
	dst2.Reset()
	Expect(wrap.Gzip.Write(dst, strings.NewReader(s))).NotTo(HaveOccurred())
	Expect(wrap.Gzip.Read(dst2, dst)).NotTo(HaveOccurred())
	Expect(dst2.String()).To(Equal(s))

	dst.Reset()
	dst2.Reset()
	Expect(wrap.Flate.Write(dst, strings.NewReader(s))).NotTo(HaveOccurred())
	Expect(wrap.Flate.Read(dst2, dst)).NotTo(HaveOccurred())
	Expect(dst2.String()).To(Equal(s))
}
