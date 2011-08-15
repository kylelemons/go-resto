package rest

import (
	"log"
	"os"
	"reflect"
)

// Map makes the object available through a RESTful interface at path.
func Map(path string, object interface{}) (*Resource, os.Error) {
	if path[len(path)-1] != '/' {
		path = path + "/"
	}

	return mapValue(path, reflect.ValueOf(object))
}

func mapValue(path string, value reflect.Value) (*Resource, os.Error) {
	if k := value.Kind(); k == reflect.Ptr || k == reflect.Interface {
		value = value.Elem()
	}

	r := &Resource{
		ro:    !value.CanSet(),
		path:  path,
		kind:  value.Kind(),
		value: value,
	}

	Handle(path, r)
	log.Printf("rest: added mapping %s for %T object", path, value.Interface())

	return r, nil
}
