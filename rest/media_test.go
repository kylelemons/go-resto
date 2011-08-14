package rest

import (
	"reflect"
	"strings"
	"testing"
)

var parseMediaTypeTests = []struct{
	Desc   string
	String []string
	Parsed MediaTypeList
}{
	{
		Desc: "Any",
		String: []string{"*/*"},
		Parsed: []*MediaType{
			&MediaType{"*","*",1.0,map[string]string{}, 0},
		},
	},
	{
		Desc: "Any Text",
		String: []string{"text/*"},
		Parsed: []*MediaType{
			&MediaType{"text","*",1.0,map[string]string{}, 0},
		},
	},
	{
		Desc: "Plain Text",
		String: []string{"text/plain"},
		Parsed: []*MediaType{
			&MediaType{"text","plain",1.0,map[string]string{}, 0},
		},
	},
	{
		Desc: "HTML 50%",
		String: []string{"text/html; q=0.5"},
		Parsed: []*MediaType{
			&MediaType{"text","html",0.5,map[string]string{}, 0},
		},
	},
	{
		Desc: "HTML Level 1 90%",
		String: []string{"text/html;level=3; q=0.9"},
		Parsed: []*MediaType{
			&MediaType{"text","html",0.9,map[string]string{
				"level": "3",
			}, 0},
		},
	},
	{
		Desc: "English",
		String: []string{"en-US"},
		Parsed: []*MediaType{
			&MediaType{"en-US","*",1.0,map[string]string{}, 0},
		},
	},
	{
		Desc: "List of values",
		String: []string{"text/*", "text/html;q=.5 ,text/html;level=1; q=0.9", "*/*;q=.1"},
		Parsed: []*MediaType{
			&MediaType{"text","*",1.0,map[string]string{}, 0},
			&MediaType{"text","html",0.5,map[string]string{}, 1},
			&MediaType{"text","html",0.9,map[string]string{
				"level": "1",
			}, 2},
			&MediaType{"*","*",0.1,map[string]string{}, 3},
		},
	},
}

func TestParseMediaTypes(t *testing.T) {
	for _, test := range parseMediaTypeTests {
		desc := test.Desc
		list := ParseMediaTypes(test.String)
		if got, want := list, test.Parsed; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %q, want %q", desc, got, want)
		}
	}
}

var testMatch = []struct{
	Requested, Known string
	Match bool
}{
	{"*/*", "text/html", true},
	{"text/*", "text/html", true},
	{"text/*", "text/html;level=1", true},
	{"text/*", "application/json", false},
	{"text/html", "text/html", true},
	{"text/html", "application/json", false},
	{"text/html", "text/html;level=1", true},
	{"text/html;level=1", "text/html", false},
}

func TestMatch(t *testing.T) {
	for _, test := range testMatch {
		r := ParseMediaTypes([]string{test.Requested})[0]
		k := ParseMediaTypes([]string{test.Known})[0]
		match := r.Match(k)
		if got, want := match, test.Match; got != want {
			t.Errorf("%q ~= %q is %v, want %v", r, k, got, want)
		}
	}
}

var testSort = []struct{
	Desc string
	In   string
	Out  string
}{
	{
		Desc: "Specific to nonspecific",
		In:   "*/*, text/*, text/html, text/html;level=1",
		Out:  "text/html;level=1, text/html, text/*, */*",
	},
	{
		Desc: "Based on initial order",
		In:   "text/html, text/plain, application/json, application/yaml",
		Out:  "text/html, text/plain, application/json, application/yaml",
	},
	{
		Desc: "Based on quality",
		In:   "text/html;q=.25, text/plain;q=.75, application/json, application/yaml;q=.5",
		Out:  "application/json, text/plain;q=0.75, application/yaml;q=0.5, text/html;q=0.25",
	},
	{
		Desc: "Mixed",
		In:   "text/html;q=0.5, text/html;level=1;q=0.5, */*;q=0.5, application/json",
		Out:  "application/json, text/html;level=1;q=0.5, text/html;q=0.5, */*;q=0.5",
	},
}

func TestSortMediaTypes(t *testing.T) {
	for _, test := range testSort {
		desc := test.Desc
		list := ParseMediaTypes([]string{test.In})
		list.Sort()
		if got, want := list.String(), test.Out; got != want {
			t.Errorf("%s: got %q, want %q", desc, got, want)
		}
	}
}

var testFilter = []struct{
	Desc string
	Available string
	Requested string
	Filtered  string
}{
	{
		Desc:      "Filter only",
		Available: "application/json",
		Requested: "text/html, application/json",
		Filtered:  "application/json",
	},
	{
		Desc:      "Filter doublewild",
		Available: "application/json",
		Requested: "text/html, */*",
		Filtered:  "application/json",
	},
	{
		Desc:      "Filter wild",
		Available: "text/plain",
		Requested: "text/html, text/*",
		Filtered:  "text/plain",
	},
	{
		Desc:      "Prefer Available Order",
		Available: "application/json, application/yaml, text/html",
		Requested: "text/html, application/json",
		Filtered:  "application/json, text/html",
	},
	{
		Desc:      "Requested Q",
		Available: "application/json, application/yaml, text/html",
		Requested: "text/html, application/json;q=0.5",
		Filtered:  "text/html, application/json;q=0.5",
	},
	{
		Desc:      "Available Q",
		Available: "application/json, application/yaml;q=0.8",
		Requested: "text/html, */*",
		Filtered:  "application/json, application/yaml;q=0.8",
	},
	{
		Desc:      "Both Q",
		Available: "application/json, application/yaml;q=0.8",
		Requested: "text/html, application/json;q=0.5, */*;q=0.25",
		Filtered:  "application/json;q=0.5, application/json;q=0.25, application/yaml;q=0.2",
	},
}

func TestFilterMediaTypes(t *testing.T) {
	for _, test := range testFilter {
		desc := test.Desc
		avail := ParseMediaTypes([]string{test.Available})
		req := ParseMediaTypes([]string{test.Requested})
		filt := avail.Filter(req)
		if got, want := filt.String(), test.Filtered; got != want {
			t.Errorf("%s: filtered %q, want %q", desc, got, want)
		}

		chosen := avail.Choose(req)
		first := test.Filtered
		if idx := strings.IndexRune(first, ','); idx >= 0 {
			first = first[:idx]
		}
		if got, want := chosen.String(), first; got != want {
			t.Errorf("%s: chose %q, want %q", desc, got, want)
		}
	}
}

var benchFilterChoose = struct{
	Available, Requested string
}{
	"application/json, application/yaml;q=.8, text/plain;q=.5, text/html;q=.5",
	"text/html, application/json, text/*;q=.5, application/*;q=.25, */*;q=.1",
}

func BenchmarkFilter(b *testing.B) {
	avail := ParseMediaTypes([]string{benchFilterChoose.Available})
	req := ParseMediaTypes([]string{benchFilterChoose.Requested})
	for i := 0; i < b.N; i++ {
		_ = avail.Filter(req)
	}
}

func BenchmarkChoose(b *testing.B) {
	avail := ParseMediaTypes([]string{benchFilterChoose.Available})
	req := ParseMediaTypes([]string{benchFilterChoose.Requested})
	for i := 0; i < b.N; i++ {
		_ = avail.Choose(req)
	}
}
