package rest

import (
	"http"
	"os"
	"reflect"
	"sync"
)

// A Resource represents an object that is accessible (and possibly editable)
// via HTTP.
type Resource struct {
	ro    bool
	path  string
	kind  reflect.Kind
	value reflect.Value
	lock  sync.RWMutex
}

// ReadOnly returns true if the Resource is read-only.
func (res *Resource) ReadOnly() bool { return res.ro }

// SetReadOnly makes the object immutable via the REST framework.
func (res *Resource) SetReadOnly() { res.ro = true }

// ServeREST handles a RESTful HTTP request.
func (res *Resource) ServeREST(w http.ResponseWriter, r *http.Request) os.Error {
	return nil
}
