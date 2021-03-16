package httphandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// MuxMatcher is an incoming *http.Request matcher
type MuxMatcher interface {
	// Test is called when MuxMatcher is added to Mux, this should be an
	// opportunity to set priority and test the implementation parameter
	// e.g. MuxMatcherPattern should check pattern, start, end
	Test() bool

	// Priority is called after Test() returning true to set a priority queue
	Priority() float64

	// Match is called by order of Priority, after its turn, it will validate
	// *http.Request and if the result is true, the http.Handler registered
	// in the entry will be served
	Match(*http.Request) bool

	// fmt.GoStringer is useful inspection methods, e.g. comparing parameter
	// or priority
	fmt.GoStringer

	// json.Marshaler and json.Unmarshaler should be supported too, in order to
	// build the MuxMatcher based on config
	json.Marshaler
	// json.Unmarshaler
}

const (
	multiplierExactPattern int = 10
	multiplierNKeys        int = 2
)

// =============================================================================
// MuxMatcherMock
// =============================================================================

type muxMatcherMock struct {
	priority    float64
	test, match bool
}

func MuxMatcherMock(priority float64, test, match bool) *muxMatcherMock {
	var _ MuxMatcher = (*muxMatcherMock)(nil)
	return &muxMatcherMock{priority, test, match}
}
func (m *muxMatcherMock) Test() bool                 { return m.test }
func (m *muxMatcherMock) Match(r *http.Request) bool { return m.match }
func (m *muxMatcherMock) Priority() float64          { return m.priority }
func (m *muxMatcherMock) GoString() string {
	return fmt.Sprintf("Mock:{Priority:%.2f, Test:%t, Match:%t}", m.priority, m.test, m.match)
}
func (m *muxMatcherMock) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"mock":{"priority":%.2f,"test":%t,"match":%t}}`,
		m.priority, m.test, m.match)), nil
}

// =============================================================================
// MuxMatcherOr
// =============================================================================

type muxMatcherOr struct {
	priority float64
	muxes    []MuxMatcher
}

func MuxMatcherOr(priority float64, muxes ...MuxMatcher) *muxMatcherOr {
	var _ MuxMatcher = (*muxMatcherOr)(nil)
	return &muxMatcherOr{priority, uniqueMuxMatcher(muxes)}
}
func (m *muxMatcherOr) Test() bool {
	match, p, c := false, 0.0, 0.0
	for i := range m.muxes {
		p, c = p+m.muxes[i].Priority(), c+1.0
		match = match || m.muxes[i] != nil && m.muxes[i].Test()
	}
	if m.priority == 0 {
		m.priority = p / c
	}
	return match
}
func (m *muxMatcherOr) Match(r *http.Request) bool {
	match := false
	for i := range m.muxes {
		if m.muxes[i] != nil && m.muxes[i].Match(r) {
			match = true
		}
	}
	return match
}
func (m *muxMatcherOr) Priority() float64 { return m.priority }

func (m *muxMatcherOr) GoString() string {
	b := strings.Builder{}
	for i := 0; i < len(m.muxes); i++ {
		if i > 0 {
			_, _ = b.WriteRune(',')
		}
		_, _ = b.WriteString(m.muxes[i].GoString())
	}
	return fmt.Sprintf("Or:{Priority:%.2f, Mux:[%s]}", m.priority, b.String())
}
func (m *muxMatcherOr) MarshalJSON() ([]byte, error) {
	b := strings.Builder{}
	for i := 0; i < len(m.muxes); i++ {
		if i > 0 {
			_, _ = b.WriteRune(',')
		}
		bs, _ := m.muxes[i].MarshalJSON()
		_, _ = b.WriteString(fmt.Sprintf(`{"priority":%.2f,"mux":%s}`,
			m.muxes[i].Priority(), bs))
	}
	return []byte(fmt.Sprintf(`{"or":{"priority":%.2f,"muxes":[%s]}}`,
		m.Priority(), b.String())), nil
}

// =============================================================================
// MuxMatcherAnd
// =============================================================================

type muxMatcherAnd struct {
	priority float64
	muxes    []MuxMatcher
}

func MuxMatcherAnd(priority float64, muxes ...MuxMatcher) *muxMatcherAnd {
	var _ MuxMatcher = (*muxMatcherAnd)(nil)
	return &muxMatcherAnd{priority, uniqueMuxMatcher(muxes)}
}
func (m *muxMatcherAnd) Test() bool {
	match, p := false, 0.0
	for i := range m.muxes {
		if m.muxes[i] != nil {
			if i == 0 {
				match = true
			}
			match = match && m.muxes[i].Test()
			p = p + m.muxes[i].Priority()
		}
	}
	if m.priority == 0 {
		m.priority = p
	}
	return match
}
func (m *muxMatcherAnd) Match(r *http.Request) bool {
	match := false
	for i := range m.muxes {
		if m.muxes[i] != nil {
			if i == 0 {
				match = true
			}
			match = match && m.muxes[i].Match(r)
		}
	}
	return match
}
func (m *muxMatcherAnd) Priority() float64 { return m.priority }

func (m *muxMatcherAnd) GoString() string {
	b := strings.Builder{}
	for i := 0; i < len(m.muxes); i++ {
		if i > 0 {
			_, _ = b.WriteRune(',')
		}
		_, _ = b.WriteString(m.muxes[i].GoString())
	}
	return fmt.Sprintf("And:{Priority:%.2f, Mux:[%s]}", m.priority, b.String())
}
func (m *muxMatcherAnd) MarshalJSON() ([]byte, error) {
	b := strings.Builder{}
	for i := 0; i < len(m.muxes); i++ {
		if i > 0 {
			_, _ = b.WriteRune(',')
		}
		bs, _ := m.muxes[i].MarshalJSON()
		_, _ = b.WriteString(fmt.Sprintf(`{"priority":%.2f,"mux":%s}`,
			m.muxes[i].Priority(), bs))
	}
	return []byte(fmt.Sprintf(`{"and":{"priority":%.2f,"muxes":[%s]}}`,
		m.Priority(), b.String())), nil
}

// =============================================================================
// MuxMatcherMethods
// =============================================================================

type muxMatcherMethods struct {
	priority float64
	methods  []string
}

// MuxMatcherMethods receive multiple methods, if contains asterisk `*` then
// the priority should be set to 0
func MuxMatcherMethods(priority float64, methods ...string) *muxMatcherMethods {
	var _ MuxMatcher = (*muxMatcherMethods)(nil)
	sort.Strings(methods)
	return &muxMatcherMethods{priority, uniqueString(methods)}
}
func (m *muxMatcherMethods) Test() bool {
	if len(m.methods) < 1 {
		return false
	}
	for i := range m.methods {
		switch m.methods[i] {
		case
			"*":
			m.priority = 0
			return true
		case
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace:
			m.priority = 1
			if l, max := float64(len(m.methods)), 10.0; l < max {
				m.priority = max - l
			}
		default:
			return false
		}
	}
	return true
}
func (m *muxMatcherMethods) Match(r *http.Request) bool {
	for i := range m.methods {
		switch m.methods[i] {
		case "*":
			return true
		case r.Method:
			return true
		}
	}
	return false
}
func (m *muxMatcherMethods) Priority() float64 { return m.priority }

func (m *muxMatcherMethods) GoString() string {
	return fmt.Sprintf("Methods:{Priority:%.2f, Methods:%s}", m.priority, m.methods)
}
func (m *muxMatcherMethods) MarshalJSON() ([]byte, error) {
	b := strings.Builder{}
	for i := 0; i < len(m.methods); i++ {
		if i > 0 {
			_, _ = b.WriteRune(',')
		}
		_, _ = b.WriteString(fmt.Sprintf(`%q`, m.methods[i]))
	}
	return []byte(fmt.Sprintf(`{"methods":{"priority":%.2f,"methods":[%s]}}`,
		m.priority, b.String())), nil
}

// =============================================================================
// MuxMatcherPattern
// =============================================================================

type muxMatcherPattern struct {
	// priority scale with pattern
	priority float64

	// pattern of named arguments using colon, e.g. `/:args1/:args2/:args3`
	pattern string

	// pair of start & end token
	start, end string

	// parse receive path *http.Request.URL.Path and extract its values
	// according to a pattern supplied to url.Values, if path followed the
	// pattern it should return true
	parse func(string) (url.Values, bool)
}

// MuxMatcherPattern receive pattern of named arguments using a pair of start
// and end string; if start is empty string, then assuming start is colon `:`,
// when end is empty string, then assuming end is slash `/`
//
// colon at start of arguments e.g. `/:args1/:args2/:args3`
//
// colon at both start and end e.g. `/:args1:/:args2:/:args3:`
//
// curly-braces at both start and end e.g. `/{args1}/{args2}/{args3}`
//
func MuxMatcherPattern(priority float64, pattern, start, end string) *muxMatcherPattern {
	var _ MuxMatcher = (*muxMatcherPattern)(nil)
	if start == "" {
		start = ":"
	}
	if end == "" {
		end = "/"
	}
	return &muxMatcherPattern{priority, pattern, start, end, nil}
}
func (m *muxMatcherPattern) Test() bool {
	if len(m.pattern) < 1 || len(m.start) < 1 || len(m.end) < 1 {
		return false
	}
	keys := []string{}
	setKeys := map[string]struct{}{}
	isInArgs, skipCount, l, b := false, 0, 0, strings.Builder{}
	lookahead := func(src, sub string, i int) bool {
		return i+len(sub) <= len(src) && src[i:i+len(sub)] == sub
	}
	for i := range m.pattern {
		if skipCount > 0 {
			skipCount--
			continue
		}

		if !isInArgs && lookahead(m.pattern, m.start, i) {
			skipCount, isInArgs = len(m.start)-1, true
			continue
		}
		isEnd, isLastChar := lookahead(m.pattern, m.end, i), i == len(m.pattern)-1
		if isInArgs && (isEnd || isLastChar) {
			skipCount, isInArgs = len(m.end)-1, false
			if i > l && i-l > 0 {
				j := i
				if !isEnd && isLastChar {
					j++
				}
				key := m.pattern[i-l : j]
				l, keys, setKeys[key] = 0, append(keys, key), struct{}{}
				_, _ = b.WriteString(`%s`)
				if lookahead(m.pattern, m.start, i+1) {
					_, _ = b.WriteString(m.end)
				}
			}
			continue
		}
		if isInArgs {
			l++
		} else {
			l = 0
			_ = b.WriteByte(m.pattern[i])
		}
	}
	if l = len(keys); l < 1 {
		m.start, m.end = "", ""
		m.priority = float64(len(m.pattern) * multiplierExactPattern)
		m.parse = func(s string) (url.Values, bool) { return nil, s == m.pattern }
		return true
	}
	if m.priority == 0 {
		m.priority = float64((len(b.String()) * multiplierExactPattern) + (l * multiplierNKeys))
	}
	sp, u, match := strings.Split(b.String(), `%s`), url.Values(nil), false
	format := strings.TrimSpace(strings.Repeat(`%s `, l))
	vs, ps := make([]string, l), make([]interface{}, l)
	for i := range ps {
		ps[i] = &vs[i]
	}
	m.parse = func(s string) (url.Values, bool) {
		for i := range sp {
			s = strings.Replace(s, sp[i], ` `, 1)
		}
		if n, err := fmt.Sscanf(s, format, ps...); err == nil && n == l {
			u, match = make(url.Values), true
			for i := range ps {
				u.Add(keys[i], vs[i])
			}
		}
		return u, match
	}
	return format != ``
}
func (m *muxMatcherPattern) Match(r *http.Request) bool {
	defer func() {
	}()
	u, match := m.parse(r.URL.Path)
	if match && len(u) > 0 {
		*r = *(set(r, ctxKeyNamedArgs{}, u))
	}
	return match
}
func (m *muxMatcherPattern) Priority() float64 { return m.priority }

func (m *muxMatcherPattern) GoString() string {
	s := fmt.Sprintf("Pattern:{Priority:%.2f, Pattern:%s, Start:%s, End:%s}", m.priority, m.pattern, m.start, m.end)
	if m.start == "" && m.end == "" {
		s = fmt.Sprintf("Pattern:{Priority:%.2f, Pattern:%s}", m.priority, m.pattern)
	}
	return s
}
func (m *muxMatcherPattern) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"pattern":{"priority":%.2f,"pattern":%q,"start":%q,"end":%q}}`,
		m.priority, m.pattern, m.start, m.end)), nil
}
