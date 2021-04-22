package kitgo

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var Dir dir_

type dir_ struct{}

type DirConfig struct{ Path string }

func (dir_) New(cfg DirConfig) *DirWrapper {
	stat, err := ioutil.ReadDir(cfg.Path)
	PanicWhen(err != nil, err)
	return &DirWrapper{cfg.Path, stat}
}

type DirWrapper struct {
	path string
	stat []os.FileInfo
}
type FilterFunc func(i int, f os.FileInfo, fs []os.FileInfo) bool

func (x *DirWrapper) Dirs() []*DirWrapper {
	cs := make([]*DirWrapper, 0)
	for _, fi := range x.FilesInfo() {
		if fi.IsDir() {
			path := filepath.Join(x.path, fi.Name())
			stat, _ := ioutil.ReadDir(path)
			cs = append(cs, &DirWrapper{path, stat})
		}
	}
	return cs
}
func (x *DirWrapper) Files() []*os.File {
	fs := make([]*os.File, 0)
	for _, fi := range x.FilesInfo() {
		if !fi.IsDir() {
			if f, err := os.Open(filepath.Join(x.path, fi.Name())); err == nil && f != nil {
				fs = append(fs, f)
			}
		}
	}
	return fs
}
func (x *DirWrapper) FilesInfo() []os.FileInfo {
	return x.stat
}
func (x *DirWrapper) Filter(filter FilterFunc) *DirWrapper {
	c := x.clone()
	if filter != nil {
		c.path, c.stat = x.path, make([]os.FileInfo, 0)
		for i, f := range x.FilesInfo() {
			if filter(i, f, x.FilesInfo()) {
				c.stat = append(c.stat, f)
			}
		}
	}
	return c
}
func (x *DirWrapper) Walk(next string) *DirWrapper {
	c := (*DirWrapper)(nil)
	for _, p := range strings.Split(strings.Replace(next, x.path, "", -1), "/") {
		if cs := x.Filter(func(i int, f os.FileInfo, fs []os.FileInfo) bool {
			return f.IsDir() && f.Name() == p
		}).Dirs(); p != "" && len(cs) > 0 {
			c = cs[0]
			break
		}
	}
	return c
}
func (x *DirWrapper) clone() (c *DirWrapper) { c = &DirWrapper{}; *c = *x; return }
