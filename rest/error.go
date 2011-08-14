package rest

import (
	"fmt"
)

type ReadOnlyError struct {
	Path   string
	Object interface{}
}

func (e *ReadOnlyError) String() string {
	return fmt.Sprintf("rest: unable to map %s: %T object is read-only", e.Path, e.Object)
}
