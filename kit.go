package kitgo

import (
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
	"os"
	"os/signal"
	"strings"
	"testing"
	"time"
	"unicode"

	jsoniter "github.com/json-iterator/go"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Dict extends `map[string]interface{}`, and it should be used to represent json-like data
type Dict map[string]interface{}

func (d *Dict) do(do func()) {
	if len(*d) < 1 {
		*d = make(Dict)
	}
	if do != nil {
		do()
	}
}
func (d *Dict) Set(k string, v interface{}) { d.do(func() { (*d)[k] = v }) }

func (d *Dict) Delete(k string) { d.do(func() { delete(*d, k) }) }

func (d *Dict) Map(fn func(k string, v interface{})) {
	d.do(func() {
		if fn != nil {
			for k, v := range *d {
				fn(k, v)
			}
		}
	})
}

// List extends `[]interface{}`, and it should be used to represent json-like data
type List []interface{}

func (l *List) do(do func()) {
	if len(*l) < 1 {
		*l = make(List, 0)
	}
	if do != nil {
		do()
	}
}
func (l *List) Add(v interface{}) { l.do(func() { *l = append(*l, v) }) }

func (l *List) Delete(k int) {
	l.do(func() {
		if k < len(*l) {
			copy((*l)[k:], (*l)[k+1:])
			(*l)[len(*l)-1] = ""
			*l = (*l)[:len(*l)-1]
		}
	})
}
func (l *List) Map(fn func(k int, v interface{})) {
	l.do(func() {
		if fn != nil {
			for k, v := range *l {
				fn(k, v)
			}
		}
	})
}

func NewErrors(errs ...error) errors {
	return errors(nil).Append(errs...)
}

// errors is a wrapper to slice of `error`
type errors []error

func (e errors) Append(errs ...error) errors {
	for i := range errs {
		if errs[i] != nil {
			e = append(e, errs[i])
		}
	}
	return e
}
func (e errors) Error() string {
	var _ error = e
	var v []string
	for i := 0; e != nil && i < len(e); i++ {
		if e != nil && (e)[i] != nil {
			v = append(v, (e)[i].Error())
		}
	}
	return strings.Join(v, "\n")
}
func (e errors) MarshalJSON() ([]byte, error) {
	var _ json.Marshaler = e
	var v []string
	for i := 0; e != nil && i < len(e); i++ {
		if e != nil && (e)[i] != nil {
			v = append(v, (e)[i].Error())
		}
	}
	return JSON.Marshal(v)
}
func (e *errors) UnmarshalJSON(b []byte) error {
	var _ json.Unmarshaler = e
	var v []string
	err := JSON.Unmarshal(b, &v)
	for i := range v {
		if v[i] != "" {
			*e = e.Append(fmt.Errorf(v[i]))
		}
	}
	return err
}

// Currency immutable struct contains price & its format
type Currency struct {
	// Tag parse according to BCP47 string
	Tag string

	// Value contains the value of currency
	Value float64

	// Format default to "%[1]s %[2]s" means currency sign, followed by value
	// %[1]s is a currency sign by language
	// %[2]s is a parsed Value
	//
	// According to each format, e.g.
	Format string
}

// MarshalJSON implement json marshaler
func (c Currency) MarshalJSON() ([]byte, error) {
	var _ json.Marshaler = c
	_, err := language.Parse(c.Tag)
	return []byte(fmt.Sprintf(`"%s"`, c.String())), err
}

// String implement a stringer interface
func (c Currency) String() string {
	var _ fmt.Stringer = c
	if c.Format == "" {
		c.Format = "%[1]s %[2]s"
	}
	tag, _ := language.Parse(c.Tag)
	unit, _ := currency.FromTag(tag)
	region, _ := tag.Region()
	printer := message.NewPrinter(tag)
	formatter := currency.Symbol
	fallback := "%.2f"
	if f, ok := map[string]string{
		"BH": "%.3f", "IQ": "%.3f", "JO": "%.3f", "KW": "%.3f", "LY": "%.3f", "OM": "%.3f",
		"TN": "%.3f", "BI": "%.0f", "CL": "%.0f", "DJ": "%.0f", "GN": "%.0f", "IS": "%.0f",
		"JP": "%.0f", "KM": "%.0f", "KR": "%.0f", "PY": "%.0f", "RW": "%.0f", "UG": "%.0f",
		"VN": "%.0f", "VU": "%.0f", "CM": "%.0f", "CF": "%.0f", "CG": "%.0f", "TD": "%.0f",
		"GQ": "%.0f", "GA": "%.0f", "BJ": "%.0f", "BF": "%.0f", "CI": "%.0f", "GW": "%.0f",
		"ML": "%.0f", "NE": "%.0f", "SN": "%.0f", "TG": "%.0f", "PF": "%.0f", "NC": "%.0f",
		"WF": "%.0f",
	}[region.String()]; ok {
		fallback = f
	}
	str := fmt.Sprintf(c.Format,
		printer.Sprint(formatter(unit.Amount(nil))),           // sign
		printer.Sprintf(message.Key("%d", fallback), c.Value), // value
	)
	b := &strings.Builder{}
	b.Grow(len(str))
	for _, r := range str {
		if unicode.IsSpace(r) {
			r = ' '
		}
		b.WriteRune(r)
	}
	// return str
	return b.String()
}

// =============================================================================
// Public
// =============================================================================

type CSS = html.CSS
type HTML = html.HTML
type HTMLAttr = html.HTMLAttr
type JS = html.JS
type JSStr = html.JSStr
type URL = html.URL
type Srcset = html.Srcset

// JSON is a jsoniter.API with ConfigFastest, satisfied encoding/json package
var JSON jsoniter.API = jsoniter.ConfigFastest

// Parallel receive cancelable context and receiving error channel along with
// list of tasks, a func() error that will execute in asynchronous manners
//
// if receiving error channel is unbuffered, it will block the current execution
// process, would be better if receiving error channel is buffered with len(tasks)
//
// context is also cancelable and this is to run in go routine
func Parallel(ctx context.Context, chanErr chan<- error, tasks ...func() error) {
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	defer close(chanErr)

	// define nonEmptyTasks to be a filtered list of tasks
	nonEmptyTasks := [](func() error){}
	for i := range tasks {
		if tasks[i] != nil {
			nonEmptyTasks = append(nonEmptyTasks, tasks[i])
		}
	}
	if len(nonEmptyTasks) < 1 {
		return
	}

	// prepare a buffered channel
	ch := make(chan error, len(nonEmptyTasks))

	// execute asynchronously all nonEmptyTasks
	for i := range nonEmptyTasks {
		go func(i int) { ch <- nonEmptyTasks[i]() }(i)
	}

	// listen to all result of nonEmptyTasks
	for range nonEmptyTasks {
		select {
		case <-ctx.Done():
			chanErr <- ctx.Err()
			return
		case err := <-ch:
			chanErr <- err
			continue
		}
	}
}

// PanicWhen execute a panic when condition is met and v is not nil
func PanicWhen(condition bool, v interface{}) {
	if condition && v != nil {
		panic(v)
	}
}

// RecoverWith will catch error value passed when panic
func RecoverWith(catch func(recv interface{})) {
	if r := recover(); r != nil && catch != nil {
		catch(r)
	}
}

// ListenToSignal will block until receiving signal from input
func ListenToSignal(sigs ...os.Signal) os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)
	return <-ch
}

func ParseDuration(s string, fallback time.Duration) (d time.Duration) {
	var err error
	if d, err = time.ParseDuration(s); err != nil || d < 1 {
		d = fallback
	}
	return
}

// =============================================================================
// Testing
// =============================================================================

// ShouldCover
func ShouldCover(code int, required float64) int {
	covermode, coverage := testing.CoverMode(), testing.Coverage()
	if code == 0 && covermode != "" && coverage < required {
		fmt.Println("" +
			fmt.Sprintf("FAIL\trequired: %.1f", required*100) + "% " +
			fmt.Sprintf("but only cover %.1f", coverage*100) + "%",
		)
		code = -1
	}
	return code
}

var _ = [][2]string{
	{".aac", "audio/aac"},
	{".abw", "application/x-abiword"},
	{".arc", "application/x-freearc"},
	{".avi", "video/x-msvideo"},
	{".azw", "application/vnd.amazon.ebook"},
	{".bin", "application/octet-stream"},
	{".bmp", "image/bmp"},
	{".bz", "application/x-bzip"},
	{".bz2", "application/x-bzip2"},
	{".csh", "application/x-csh"},
	{".css", "text/css"},
	{".csv", "text/csv"},
	{".doc", "application/msword"},
	{".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	{".eot", "application/vnd.ms-fontobject"},
	{".epub", "application/epub+zip"},
	{".gz", "application/gzip"},
	{".gif", "image/gif"},
	{".htm", "text/html"},
	{".html", "text/html"},
	{".ico", "image/vnd.microsoft.icon"},
	{".ics", "text/calendar"},
	{".jar", "application/java-archive"},
	{".jpeg", "image/jpeg"},
	{".jpg", "image/jpeg"},
	{".js", "text/javascript"},
	{".json", "application/json"},
	{".jsonld", "application/ld+json"},
	{".mid", "audio/midi"},
	{".midi", "audio/midi"},
	{".mjs", "text/javascript"},
	{".mp3", "audio/mpeg"},
	{".mpeg", "video/mpeg"},
	{".mpkg", "application/vnd.apple.installer+xml"},
	{".odp", "application/vnd.oasis.opendocument.presentation"},
	{".ods", "application/vnd.oasis.opendocument.spreadsheet"},
	{".odt", "application/vnd.oasis.opendocument.text"},
	{".oga", "audio/ogg"},
	{".ogv", "video/ogg"},
	{".ogx", "application/ogg"},
	{".opus", "audio/opus"},
	{".otf", "font/otf"},
	{".png", "image/png"},
	{".pdf", "application/pdf"},
	{".php", "application/x-httpd-php"},
	{".ppt", "application/vnd.ms-powerpoint"},
	{".pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	{".rar", "application/vnd.rar"},
	{".rtf", "application/rtf"},
	{".sh", "application/x-sh"},
	{".svg", "image/svg+xml"},
	{".swf", "application/x-shockwave-flash"},
	{".tar", "application/x-tar"},
	{".tif", "image/tiff"},
	{".tiff", "image/tiff"},
	{".ts", "video/mp2t"},
	{".ttf", "font/ttf"},
	{".txt", "text/plain"},
	{".vsd", "application/vnd.visio"},
	{".wav", "audio/wav"},
	{".weba", "audio/webm"},
	{".webm", "video/webm"},
	{".webp", "image/webp"},
	{".woff", "font/woff"},
	{".woff2", "font/woff2"},
	{".xhtml", "application/xhtml+xml"},
	{".xls", "application/vnd.ms-excel"},
	{".xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	{".xml", "text/xml"},
	{".xul", "application/vnd.mozilla.xul+xml"},
	{".zip", "application/zip"},
	{".3gp", "video/3gpp"},
	{".3g2", "video/3gpp2"},
	{".7z", "application/x-7z-compressed"},
}
