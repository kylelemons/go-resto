package rest

import (
	"http"
	"log"
	"os"
	"strings"
)

var DefaultServeMux = http.DefaultServeMux

func ListenAndServe(addr string) os.Error {
	// TODO(kevlar): add instrumentation for examining modifications

	server := http.Server{
		Addr: addr,
		Handler: DefaultServeMux,
	}
	return server.ListenAndServe()
}

type Handler interface{
	ServeREST(http.ResponseWriter, *http.Request) os.Error
}

type ErrorCoder interface {
	ErrorCode() int
}

// Handle maps the given handler (typically a *Resource) at the given path and
// provides a first level of logging, locking, and access control for the
// resource.
//
// A request for POST, PUT, DELETE, or PATCH on a ReadOnly resource will result
// in an HTTP Forbidden response.  A CONNECT or other request will also
// generate the necessary HTTP error code.
//
// If the method is a "safe" method (e.g. GET), the resource is locked for reading.
// If the method is an "unsafe" method (e.g. PUT), the resource is locked for writing.
// The resource is unlocked when the request handling completes.
//
// Errors returned by rhe ServeREST function of the handler are sent to the
// client.  By default, these are sent with an HTTP Internal Server Error
// response, but if the error has an ErrorCode() int method, the return value
// of that method will be used is the status code instead.
func Handle(path string, handler Handler) {
	DefaultServeMux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		log := func(message string) {
			log.Printf("rest: %s: %s", r.RemoteAddr, message)
		}

		var res *Resource
		if r, ok := handler.(*Resource); ok {
			res = r
		}

		switch r.Method {
		case "POST", "PUT", "DELETE", "PATCH":
			if res != nil {
				if res.ro {
					log("attempt to modify read-only resource blocked")
					http.Error(w, "Read-Only Resource", http.StatusForbidden)
					return
				}
				res.lock.Lock()
				defer res.lock.Unlock()
			}
		case "CONNECT":
			log("CONNECT attempt blocked")
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		case "GET", "HEAD":
			if res != nil {
				res.lock.RLock()
				defer res.lock.Unlock()
			}
		case "OPTIONS":
			allow := []string{"OPTIONS", "HEAD", "GET", "POST", "PATCH", "PUT", "DELETE"}
			if res != nil && res.ro {
				allow = allow[:3]
			}
			w.Header().Set("Allow", strings.Join(allow, ", "))
		default:
			log("unknown method: " + r.Method)
			http.Error(w, "Not Implemented", http.StatusNotImplemented)
			return
		}

		err := handler.ServeREST(w, r)
		if err == nil {
			return
		}

		log(err.String())

		status := http.StatusInternalServerError
		if err, ok := err.(ErrorCoder); ok {
			status = err.ErrorCode()
		}

		http.Error(w, err.String(), status)
	})
}
