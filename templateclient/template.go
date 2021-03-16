package templateclient

import (
	"bytes"
	"html/template"
	"io"

	"github.com/hokonco/kitgo"
)

// New return a new html/template
//
// to add template source to module, do call .Parse that receive multiple io.Reader
func New() *Client { return &Client{template.New("")} }

type Client struct{ tpl *template.Template }

// Parse receive multiple io.Reader, this is useful for testing purpose,
// using strings.NewReader("pattern"), or even *os.File
func (t *Client) Parse(r ...io.Reader) error {
	errs := kitgo.Errors(nil)
	for i := range r {
		buf := &bytes.Buffer{}
		_, err := io.Copy(buf, r[i])
		errs = errs.Append(err)
		_, err = t.tpl.Parse(buf.String())
		errs = errs.Append(err)
	}
	return errs
}

// Execute receive a key as string and data (the easiest is using map[string]interface{})
// key is usually a `define` keyword in your template
//
// 	{{define "index"}} <html>{{.content}}</html> {{end}}
//
// to set `.content`, send map[string]interface{}{ "content": "hello, world" }
//
// note that set html need to convert to template.HTML, use the helper in `internal/datatype`
//
// to set `.content`, send map[string]interface{}{ "content": datatype.HTML("<div>hello</div> <div>world</div>") }
func (t *Client) Execute(key string, data interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := t.tpl.ExecuteTemplate(buf, key, data)
	return buf.Bytes(), err
}
