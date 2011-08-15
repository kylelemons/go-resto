// Package rest is a RESTful Object framework.
//
// When a variable is Mapped to a given path, if it is a pointer type the
// variable will be mutable via the REST framework.  If it is not a pointer
// type or SetReadOnly() is caled on the returned *Resource, the variable will
// be accessible, but not modifiable.
//
// Below are the types understood as objects mapped through the REST interface,
// and what the various methods do when performed on an object of that type. If
// a method is not described below, it is not suported.
//
// All Types:
//   HEAD requests act the same as GET, but with headers only (the body is not
//     sent as part of the reply).
//   OPTIONS requests respond with the acceptable methods for that object in
//     the response Allow header.
//
// Basic Types: (int, float, string, etc)
//   GET requests will return the value (as described below) of the variable.
//   PUT requests will set the value (as described below) of the variable.
//
// Collection Types: (map, slice, etc)
//   GET requests will return all of the values stored in the collection.
//   PUT will replace the collection with the given set of values.
//   POST will add a new element to the collection.
//   - Subelements of a map or slice (by string key or numeric index):
//     GET requests return the value of the element
//     PUT requests create or replace the element
//
// Object Types: (interfaces, structs, etc)
//   GET requests will return the entire value
//   PUT will modify the corresponding parts of the value
//   - Fields of a structure are mapped below the object in the same way
//     they would be if the field were Mapped directly.
package rest
