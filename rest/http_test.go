package rest

import (
	"bytes"
	"http"
	"http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

type testObjectType struct{
	String  string
	Numbers []int
	Map     map[string]bool
}
var testObject = testObjectType{
	String: "teststr",
	Numbers: []int{6,9,42},
	Map:     map[string]bool{
		"true": true,
		"false": false,
	},
}

var mapTests = []struct {
	InPath  string
	InObj   interface{}
	OutPath string
	OutType reflect.Type
	OutKind reflect.Kind
	OutRO   bool

	// Set in TestMap for later use
	*Resource
}{
	{
		InPath:  "/mutable",
		InObj:   &testObject,
		OutPath: "/mutable/",
		OutType: reflect.TypeOf(testObjectType{}),
		OutKind: reflect.Struct,
		OutRO:   false,
	},
	{
		InPath:  "/readonly",
		InObj:   testObject,
		OutPath: "/readonly/",
		OutType: reflect.TypeOf(testObjectType{}),
		OutKind: reflect.Struct,
		OutRO:   true,
	},
	{
		InPath:  "/int",
		InObj:   new(int),
		OutPath: "/int/",
		OutType: reflect.TypeOf(int(0)),
		OutKind: reflect.Int,
		OutRO:   false,
	},
	{
		InPath:  "/str",
		InObj:   "test",
		OutPath: "/str/",
		OutType: reflect.TypeOf(""),
		OutKind: reflect.String,
		OutRO:   true,
	},
}

func TestMap(t *testing.T) {
	for i, test := range mapTests {
		desc := test.OutPath

		res, err := Map(test.InPath, test.InObj)
		if err != nil {
			t.Errorf("map(%q): %s", test.InPath, err)
		}
		mapTests[i].Resource = res

		if got, want := res.path, test.OutPath; got != want {
			t.Errorf("%s - path = %q, want %q", desc, got, want)
		}
		if got, want := res.value.Type(), test.OutType; got != want {
			t.Errorf("%s - type = %v, want %v", desc, got, want)
		}
		if got, want := res.kind, test.OutKind; got != want {
			t.Errorf("%s - kind = %v, want %v", desc, got, want)
		}
		if got, want := res.ro, test.OutRO; got != want {
			t.Errorf("%s - readonly = %v, want %v", desc, got, want)
		}
	}
}

type errorResponder int
func (e errorResponder) ServeREST(w http.ResponseWriter, r *http.Request) os.Error { return e }
func (e errorResponder) String() string { return "errorResponder" }
func (e errorResponder) ErrorCode() int { return int(e) }

var errorCoderPaths = []struct{
	Path string
	Code int
}{
	{"/error/auth", http.StatusUnauthorized},
}

func TestErrorCoder(t *testing.T) {
	for _, pathcode := range errorCoderPaths {
		Handle(pathcode.Path, errorResponder(pathcode.Code))
	}
}

var handleTests = []struct {
	Path     string
	Method   string
	Body     string
	ErrCode  int
	Contains string
}{
	{"/missing/", "GET", "", http.StatusNotFound, ""},
	{"/mutable/", "GET", "", http.StatusOK, ""},
	{"/mutable/", "HEAD", "", http.StatusOK, ""},
	{"/mutable/", "DELETE", "", http.StatusOK, ""},
	{"/mutable/", "PATCH", "", http.StatusOK, ""},
	{"/mutable/", "POST", "", http.StatusOK, ""},
	{"/mutable/", "PUT", "", http.StatusOK, ""},
	{"/readonly/", "GET", "", http.StatusOK, ""},
	{"/readonly/", "HEAD", "", http.StatusOK, ""},
	{"/readonly/", "DELETE", "", http.StatusForbidden, ""},
	{"/readonly/", "PATCH", "", http.StatusForbidden, ""},
	{"/readonly/", "POST", "", http.StatusForbidden, ""},
	{"/readonly/", "PUT", "", http.StatusForbidden, ""},
	{"/mutable/", "CONNECT", "", http.StatusMethodNotAllowed, ""},
	{"/mutable/", "UNKNOWN", "", http.StatusNotImplemented, ""},
	{"/error/auth", "GET", "", http.StatusUnauthorized, ""},
	{"/int/", "GET", "", http.StatusOK, "0"},
	{"/str/", "GET", "", http.StatusOK, "test"},
	{"/mutable/", "GET", "", http.StatusOK,
		`{"String":"teststr","Numbers":[6,9,42],"Map":{"false":false,"true":true}}`},
	{"/mutable/string", "GET", "", http.StatusOK, "teststr"},
	{"/mutable/nUMbErs", "GET", "", http.StatusOK, "[6,9,42]"},
	{"/mutable/numbers/2", "GET", "", http.StatusOK, "42"},
	{"/mutable/numbers/true", "GET", "", http.StatusNotFound, ""},
	{"/mutable/map", "GET", "", http.StatusOK, `{"false":false,"true":true}`},
	{"/mutable/map/true", "GET", "", http.StatusOK, `true`},
	{"/mutable/map/2", "GET", "", http.StatusNotFound, ""},
	{"/str/blah", "GET", "", http.StatusNotFound, ""},
}

func TestHandle(t *testing.T) {
	for _, test := range handleTests {
		desc := test.Method + " " + test.Path
		r, err := http.NewRequest(test.Method, test.Path, bytes.NewBufferString(test.Body))
		if err != nil {
			t.Errorf("%s - newrequest: %s", desc, err)
			continue
		}
		w := httptest.NewRecorder()
		r.RemoteAddr = "unittest"

		DefaultServeMux.ServeHTTP(w, r)
		if got, want := w.Code, test.ErrCode; got != want {
			t.Errorf("%s - code = %v, want %v", desc, got, want)
		}
		if bytes.Index(w.Body.Bytes(), []byte(test.Contains)) < 0 {
			t.Errorf("%s - body does not contain %q:", desc, test.Contains)
			t.Errorf("%s", w.Body.String())
		}
	}
}

var optionsTests = []struct {
	Path  string
	Allow string
}{
	{"/mutable/", "OPTIONS, HEAD, GET, POST, PATCH, PUT, DELETE"},
	{"/readonly/", "OPTIONS, HEAD, GET"},
}

func TestOptions(t *testing.T) {
	for _, test := range optionsTests {
		desc := test.Path
		r, err := http.NewRequest("OPTIONS", test.Path, nil)
		if err != nil {
			t.Errorf("%s - newrequest: %s", desc, err)
			continue
		}
		w := httptest.NewRecorder()

		DefaultServeMux.ServeHTTP(w, r)
		if got, want := strings.Join(w.HeaderMap["Allow"], ", "), test.Allow; got != want {
			t.Errorf("%s - allow = %q, want %q", desc, got, want)
		}
	}
}
