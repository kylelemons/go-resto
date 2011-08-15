package rest

import (
	"http"
	"json"
	"os"
	"reflect"
	"strconv"
	"strings"
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
	value := res.value
	path := r.URL.Path

	// Make sure the path has the proper prefix
	if !strings.HasPrefix(path, res.path) {
		return os.NewError("rest: misdirected request")
	}
	path = path[len(res.path):]

	// Strip the entity path of a trailing /
	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	for curr, next := path, ""; len(curr) > 0; curr, next = next, "" {
		if idx := strings.IndexRune(curr, '/'); idx >= 0 {
			curr, next = curr[:idx], curr[idx+1:]
		}

		// TODO(kevlar): check function
		for value.Kind() == reflect.Ptr {
			// TODO(kevlar): avoid nil dereference
			value = value.Elem()
		}

		switch value.Kind() {
		case reflect.Array, reflect.Slice:
			idx, err := strconv.Atoi(curr)
			if err != nil || idx < 0 || idx > value.Len() {
				break
			}
			value = value.Index(idx)
			continue
		case reflect.Map:
			if value.Type().Key().Kind() != reflect.String {
				break
			}
			elem := value.MapIndex(reflect.ValueOf(curr))
			if !elem.IsValid() {
				break
			}
			value = elem
			continue
		case reflect.Struct:
			lower := strings.ToLower(curr)
			field := value.FieldByNameFunc(func(name string) bool {
				return strings.ToLower(name) == lower
			})
			if !field.IsValid() {
				break
			}
			value = field
			continue
		}
		return &BadSub{res.path, path, res.value.Interface()}
	}

	for value.Kind() == reflect.Ptr {
		// TODO(kevlar): avoid nil dereference
		value = value.Elem()
	}

	switch res.kind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.String:
		return serveSimple(value, w, r)
	case reflect.Array, reflect.Slice:
		return serveCollection(value, w, r)
	case reflect.Map:
		return serveCollection(value, w, r)
	case reflect.Struct:
		return serveObject(value, w, r)
	default:
		return &UnhandledType{r.URL.Path, res.value.Interface()}
	}

	panic("unreachable")
}

func serveSimple(val reflect.Value, w http.ResponseWriter, r *http.Request) os.Error {
	switch r.Method {
	case "HEAD", "GET", "PUT":
	default:
		return &BadMethod{r.URL.Path, r.Method, val.Interface()}
	}

	ctype := "application/json"
	// TODO(kevlar): Content type negotiation
	w.Header().Set("Content-Type", ctype)

	if r.Method == "HEAD" {
		return nil
	}

	js, err := json.Marshal(val.Interface())
	if err != nil {
		return &FailedEncode{err,ctype,val.Interface()}
	}

	if _, err := w.Write(js); err != nil {
		return err
	}

	return nil
}

func serveCollection(val reflect.Value, w http.ResponseWriter, r *http.Request) os.Error {
	switch r.Method {
	case "HEAD", "GET", "PUT", "POST":
	default:
		return &BadMethod{r.URL.Path, r.Method, val.Interface()}
	}

	ctype := "application/json"
	// TODO(kevlar): Content type negotiation
	w.Header().Set("Content-Type", ctype)

	if r.Method == "HEAD" {
		return nil
	}

	js, err := json.Marshal(val.Interface())
	if err != nil {
		return &FailedEncode{err,ctype,val.Interface()}
	}

	if _, err := w.Write(js); err != nil {
		return err
	}

	return nil
}

func serveObject(val reflect.Value, w http.ResponseWriter, r *http.Request) os.Error {
	switch r.Method {
	case "HEAD", "GET", "PUT":
	default:
		return nil
		return &BadMethod{r.URL.Path, r.Method, val.Interface()}
	}

	ctype := "application/json"
	// TODO(kevlar): Content type negotiation
	w.Header().Set("Content-Type", ctype)

	if r.Method == "HEAD" {
		return nil
	}

	js, err := json.Marshal(val.Interface())
	if err != nil {
		return &FailedEncode{err,ctype,val.Interface()}
	}

	if _, err := w.Write(js); err != nil {
		return err
	}

	return nil
}
