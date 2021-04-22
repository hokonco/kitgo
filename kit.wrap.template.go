package kitgo

import (
	"bytes"
	html "html/template"
	"io"
	text "text/template"
)

var Template template

type template struct {
	// HTML implement a wrapper around standard "html/template"
	// so that the HTMLTemplate is limited to Parse & Execute
	HTML html_

	// Text implement a wrapper around standard "text/template"
	// so that the TextTemplate is limited to Parse & Execute
	Text text_
}
type html_ struct{}
type text_ struct{}

// New return a new html/template
//
// to add template source to module, do call .Parse that receive multiple io.Reader
func (html_) New() *HTMLTemplateWrapper { return &HTMLTemplateWrapper{html.New("")} }

// New return a new text/template
//
// to add template source to module, do call .Parse that receive multiple io.Reader
func (text_) New() *TextTemplateWrapper { return &TextTemplateWrapper{text.New("")} }

type HTMLTemplateWrapper struct{ w *html.Template }
type TextTemplateWrapper struct{ w *text.Template }

func (html_) CSS(s string) html.CSS           { return html.CSS(s) }
func (html_) HTML(s string) html.HTML         { return html.HTML(s) }
func (html_) HTMLAttr(s string) html.HTMLAttr { return html.HTMLAttr(s) }
func (html_) JS(s string) html.JS             { return html.JS(s) }
func (html_) JSStr(s string) html.JSStr       { return html.JSStr(s) }
func (html_) URL(s string) html.URL           { return html.URL(s) }
func (html_) Srcset(s string) html.Srcset     { return html.Srcset(s) }

// Parse receive multiple io.Reader, this is useful for testing purpose,
// using strings.NewReader("pattern"), or even *os.File
func (x *HTMLTemplateWrapper) Parse(r ...io.Reader) error {
	errs := NewErrors()
	for i := range r {
		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, r[i])
		errs = errs.Append(err)
		_, err = x.w.Parse(buf.String())
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
func (x *HTMLTemplateWrapper) Execute(w io.Writer, key string, data interface{}) error {
	return x.w.ExecuteTemplate(w, key, data)
}

// Parse receive multiple io.Reader, this is useful for testing purpose,
// using strings.NewReader("pattern"), or even *os.File
func (x *TextTemplateWrapper) Parse(r ...io.Reader) error {
	errs := NewErrors()
	for i := range r {
		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, r[i])
		errs = errs.Append(err)
		_, err = x.w.Parse(buf.String())
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
func (x *TextTemplateWrapper) Execute(w io.Writer, key string, data interface{}) error {
	return x.w.ExecuteTemplate(w, key, data)
}
