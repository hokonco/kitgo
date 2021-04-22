package kitgo_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hokonco/kitgo"
	. "github.com/onsi/gomega"
)

func Test_Template(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	derive := `
{{define "derive"}}{{template "alpha" .}}<h1>{{.title}}</h1>{{template "omega" .}}{{end}}
{{define "alpha"}}<!DOCTYPE html><html {{.attr_html}}><head {{.attr_head}}><meta charset="UTF-8"></head><body {{.attr_body}}><main {{.attr_main}}>{{end}}
{{define "omega"}}</main></body></html>{{end}}`
	b2 := `<!DOCTYPE html><html ><head ><meta charset="UTF-8"></head><body ><main ><h1>derive</h1></main></body></html>`

	t.Run("html", func(t *testing.T) {
		wrap := kitgo.Template.HTML.New()
		err := wrap.Parse(
			strings.NewReader(derive),
			strings.NewReader(derive),
		)
		Expect(err).To(BeNil())
		buf := new(bytes.Buffer)
		if err := wrap.Execute(buf, "derive", map[string]interface{}{
			"attr_html": kitgo.Template.HTML.HTMLAttr(""),
			"attr_head": kitgo.Template.HTML.HTMLAttr(""),
			"attr_body": kitgo.Template.HTML.HTMLAttr(""),
			"attr_main": kitgo.Template.HTML.HTMLAttr(""),
			"title":     "derive",
		}); err != nil {
			panic(err)
		} else {
			Expect(buf.String()).To(Equal(b2))
		}
	})
	t.Run("text", func(t *testing.T) {
		wrap := kitgo.Template.Text.New()
		err := wrap.Parse(
			strings.NewReader(derive),
			strings.NewReader(derive),
		)
		Expect(err).To(BeNil())
		buf := new(bytes.Buffer)
		if err := wrap.Execute(buf, "derive", map[string]interface{}{
			"attr_html": kitgo.Template.HTML.HTMLAttr(""),
			"attr_head": kitgo.Template.HTML.HTMLAttr(""),
			"attr_body": kitgo.Template.HTML.HTMLAttr(""),
			"attr_main": kitgo.Template.HTML.HTMLAttr(""),
			"title":     "derive",
		}); err != nil {
			panic(err)
		} else {
			Expect(buf.String()).To(Equal(b2))
		}
	})
}
