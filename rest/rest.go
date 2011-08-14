// Package rest is a RESTful Object framework.
package rest

import (
	"log"
	"os"
	"reflect"
)

func Map(path string, object interface{}) (*Resource, os.Error) {
	value := reflect.ValueOf(object)

	if k := value.Kind(); k == reflect.Ptr || k == reflect.Interface {
		return Map(path, value.Elem())
	}

	if path[len(path)-1] != '/' {
		path = path + "/"
	}

	r := &Resource{
		ro:    !value.CanSet(),
		path:  path,
		kind:  value.Kind(),
		value: value,
	}

	Handle(path, r)
	log.Printf("rest: added mapping %s for %T object", path, object)

	return r, nil
}
