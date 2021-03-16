package dirclient

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hokonco/kitgo"
)

type Config struct {
	Path string `yaml:"path" json:"path"`
}

func New(cfg Config) *Client {
	stat, err := ioutil.ReadDir(cfg.Path)
	kitgo.PanicWhen(err != nil, err)
	return &Client{cfg.Path, stat}
}

type Client struct {
	path string
	stat []os.FileInfo
}
type FilterFunc func(i int, f os.FileInfo, fs []os.FileInfo) bool

func (x *Client) Dirs() []*Client {
	cs := make([]*Client, 0)
	for _, fi := range x.FilesInfo() {
		if fi.IsDir() {
			path := filepath.Join(x.path, fi.Name())
			stat, _ := ioutil.ReadDir(path)
			cs = append(cs, &Client{path, stat})
		}
	}
	return cs
}
func (x *Client) Files() []*os.File {
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
func (x *Client) FilesInfo() []os.FileInfo {
	return x.stat
}
func (x *Client) Filter(filter FilterFunc) *Client {
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
func (x *Client) Walk(next string) *Client {
	c := (*Client)(nil)
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
func (x *Client) clone() (c *Client) { c = &Client{}; *c = *x; return }
