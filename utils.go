package confless

import (
	"reflect"
)

// Creates a new object of the given type.
func MakeNewObject(t reflect.Type) any {
	// Unwrap the type until it is not a pointer.
	objType := t
	depth := 0
	for objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
		depth++
	}

	// Create a new object of the given type (returns a pointer).
	obj := reflect.New(objType)

	// Add the necessary number of pointers to the object.
	for i := 0; i < (depth - 1); i++ {
		obj = obj.Addr()
	}

	return obj.Interface()
}
