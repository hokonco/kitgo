package templateclient_test

import (
	"os"
	"strings"
	"testing"

	"github.com/hokonco/kitgo"
	"github.com/hokonco/kitgo/templateclient"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) { os.Exit(kitgo.ShouldCover(m.Run(), 1.00)) }

func Test_client_template(t *testing.T) {
	t.Parallel()
	Expect := NewWithT(t).Expect

	driv := `
{{define "driv"}}{{template "alpha" .}}<h1>{{.driv_title}}</h1>{{template "omega" .}}{{end}}
{{define "alpha"}}<!DOCTYPE html><html {{.attr_html}}><head {{.attr_head}}><meta charset="UTF-8"></head><body {{.attr_body}}><main {{.attr_main}}>{{end}}
{{define "omega"}}</main></body></html>{{end}}`
	b2 := `<!DOCTYPE html><html ><head ><meta charset="UTF-8"></head><body ><main ><h1>derive</h1></main></body></html>`

	t.Run("parse map[string]interface{}", func(t *testing.T) {
		tmplCli := templateclient.New()
		err := tmplCli.Parse(
			strings.NewReader(driv),
			strings.NewReader(driv),
		)
		Expect(err).To(BeNil())
		if parsed, err := tmplCli.Execute("driv", map[string]interface{}{
			"attr_html":  kitgo.HTMLAttr(""),
			"attr_head":  kitgo.HTMLAttr(""),
			"attr_body":  kitgo.HTMLAttr(""),
			"attr_main":  kitgo.HTMLAttr(""),
			"driv_title": "derive",
		}); err != nil {
			panic(err)
		} else {
			Expect(parsed).To(Equal([]byte(b2)))
		}
	})
}
