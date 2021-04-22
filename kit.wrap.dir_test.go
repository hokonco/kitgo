package kitgo_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hokonco/kitgo"
	. "github.com/onsi/gomega"
)

func Test_client_dir(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	dir := t.TempDir()
	tds := make([]string, 0)
	appendTDS := func(d string, _ error) { tds = append(tds, d) }
	appendTDS(ioutil.TempDir(dir, "*-sub1"))
	appendTDS(ioutil.TempDir(dir, "*-sub2"))
	appendTDS(ioutil.TempDir(dir, "*-sub3"))

	tfs := make([]*os.File, 0)
	appendTFS := func(f *os.File, _ error) { tfs = append(tfs, f) }
	appendTFS(ioutil.TempFile(dir, "test-1"))
	appendTFS(ioutil.TempFile(dir, "test-2"))
	appendTFS(ioutil.TempFile(dir, "test-3"))
	appendTFS(ioutil.TempFile(dir, "test-4"))
	appendTFS(ioutil.TempFile(dir, "test-5"))

	_ = ioutil.WriteFile(filepath.Join(tds[0], "sub1-test-1"), []byte("text sub1-test-1"), 0666)
	_ = ioutil.WriteFile(filepath.Join(tds[0], "sub1-test-2"), []byte("text sub1-test-2"), 0666)
	_ = ioutil.WriteFile(filepath.Join(tds[1], "sub2-test-1"), []byte("text sub2-test-1"), 0666)
	_ = ioutil.WriteFile(filepath.Join(tds[1], "sub2-test-2"), []byte("text sub2-test-2"), 0666)
	_ = ioutil.WriteFile(filepath.Join(tds[2], "sub3-test-1"), []byte("text sub3-test-1"), 0666)
	_ = ioutil.WriteFile(filepath.Join(tds[2], "sub3-test-2"), []byte("text sub3-test-2"), 0666)

	wrap := kitgo.Dir.New(kitgo.DirConfig{dir})
	Expect(wrap).NotTo(BeNil())

	Expect(wrap.Walk(tds[0])).NotTo(BeNil())
	Expect(wrap.Walk("xx/xx")).To(BeNil())

	skip := func(int, os.FileInfo, []os.FileInfo) bool { return false }
	pass := func(int, os.FileInfo, []os.FileInfo) bool { return true }
	Expect(len(wrap.Dirs())).To(Equal(len(tds)))
	Expect(len(wrap.Files())).To(Equal(len(tfs)))
	Expect(len(wrap.Filter(skip).Dirs())).To(Equal(0))
	Expect(len(wrap.Filter(skip).Files())).To(Equal(0))
	Expect(len(wrap.Filter(pass).Dirs())).To(Equal(len(tds)))
	Expect(len(wrap.Filter(pass).Files())).To(Equal(len(tfs)))
}
