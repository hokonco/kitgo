package httphandler

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

func New() *Mux { return &Mux{} }

// ServeHTTP implement http.Handler interface
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var _ http.Handler = m
	code := 0
	defer func() {
		if rcv := recover(); rcv != nil {
			if m.PanicHandler == nil {
				code = http.StatusInternalServerError

				http.Error(w, http.StatusText(code), code)
				return
			}
			r = set(r, ctxKeyPanicRecovery{}, rcv)
			m.PanicHandler.ServeHTTP(w, r)
			return
		}
	}()
	found := false
	for _, e := range m.muxEntries {
		if e.matcher != nil && e.next != nil {
			if found = e.matcher.Match(r); found {
				e.next.ServeHTTP(w, r)
				return
			}
		}
	}
	if !found {
		if m.NotFoundHandler == nil {
			code = http.StatusNotFound
			http.Error(w, http.StatusText(code), code)
			return
		}
		m.NotFoundHandler.ServeHTTP(w, r)
		return
	}
}

// GoString is useful for inspection by listing all its properties,
// also called when encountering verb %#v in fmt.Printf
func (m *Mux) GoString() string {
	var _ fmt.GoStringer = m
	b := &strings.Builder{}
	for i := 0; i < len(m.muxEntries); i++ {
		if i > 0 {
			_, _ = b.WriteRune(',')
		}
		_, _ = b.WriteString(m.muxEntries[i].matcher.GoString())
	}
	return fmt.Sprintf("Mux:{Entries:[%s], NotFoundHandler:%t, PanicHandler:%t}",
		b.String(), m.NotFoundHandler != nil, m.PanicHandler != nil)
}

// With will register http.Handler with any implementation of MuxMatcher
func (m *Mux) With(next http.Handler, matcher MuxMatcher) *Mux {
	if next == nil || next == m || matcher == nil || !matcher.Test() {
		return m
	}
	exist := false
	for _, e := range m.muxEntries {
		exist = exist || e.matcher.GoString() == matcher.GoString()
		exist = exist || bytes.Equal(jsonBytes(e.matcher), jsonBytes(matcher))
	}
	if !exist {
		m.muxEntries = append(m.muxEntries, muxEntry{next, matcher})
		sort.SliceStable(m.muxEntries, func(i, j int) bool {
			e := m.muxEntries
			return e[i].matcher.Priority() > e[j].matcher.Priority()
		})
	}
	return m
}

// Handle will register http.Handler with MuxMatcherMethods on method and
// MuxMatcherPattern on pattern, see more details on each mux matcher
// implementation
func (m *Mux) Handle(method string, pattern string, next http.Handler) *Mux {
	m.With(next, MuxMatcherAnd(0,
		MuxMatcherMethods(0, method),
		MuxMatcherPattern(0, pattern, "", ""),
	))
	return m
}

// Mux holds a map of entries
type Mux struct {
	muxEntries muxEntries

	// PanicHandler can access the error recovered via PanicRecoveryFromRequest,
	// PanicRecoveryFromRequest is a helper under httphandler package
	PanicHandler    http.Handler
	NotFoundHandler http.Handler
}

// muxEntry is an element of entries listed in mux
type muxEntry struct {
	next    http.Handler
	matcher MuxMatcher
}

// muxEntries is a list of entry
type muxEntries []muxEntry
