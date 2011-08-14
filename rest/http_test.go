package rest

import (
	"reflect"
	"testing"
)

type testObjectType struct{}
var testObject testObjectType

var mapTests = []struct{
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
		InPath: "/mutable",
		InObj:  &testObject,
		OutPath: "/mutable/",
		OutType: reflect.TypeOf(testObjectType{}),
		OutKind: reflect.Struct,
		OutRO:   false,
	},
	{
		InPath: "/readonly",
		InObj:  testObject,
		OutPath: "/readonly/",
		OutType: reflect.TypeOf(testObjectType{}),
		OutKind: reflect.Struct,
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

var handleTests = []struct{
	Method string
	Path   string
	Body   string
}{
}

func TestHandle(t *testing.T) {
}
