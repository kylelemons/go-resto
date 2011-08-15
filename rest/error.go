package rest

import (
	"fmt"
	"http"
	"os"
)

type UnhandledType struct {
	Path   string
	Object interface{}
}
func (e *UnhandledType) String() string {
	return fmt.Sprintf("rest: unhandled type %T", e.Object)
}
func (e *UnhandledType) ErrorCode() int {
	return http.StatusNotImplemented
}

type BadMethod struct {
	Path   string
	Method string
	Object interface{}
}
func (e *BadMethod) String() string {
	return fmt.Sprintf("rest: %s unsupported for %T", e.Method, e.Object)
}
func (e *BadMethod) ErrorCode() int {
	return http.StatusMethodNotAllowed
}

type BadSub struct {
	ResURI string
	SubURI string
	Object interface{}
}
func (e *BadSub) String() string {
	return fmt.Sprintf("rest: %s (%T) has no sub-entity %q", e.ResURI, e.Object, e.SubURI)
}
func (e *BadSub) ErrorCode() int {
	return http.StatusNotFound
}

type FailedEncode struct {
	Err    os.Error
	Media  string
	Object interface{}
}
func (e *FailedEncode) String() string {
	return fmt.Sprintf("rest: encoding %T as %s: %s", e.Object, e.Media, e.Err)
}
