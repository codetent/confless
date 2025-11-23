package reflectutil

import (
	"reflect"
)

// Returns a pointer to the given value.
func PtrTo[T any](v T) *T {
	return &v
}

// Creates a new object of the given type.
func MakeNewObject(t reflect.Type) any {
	// Unwrap the type until it is not a pointer.
	objType := t
	depth := 0
	for objType.Kind() == reflect.Pointer {
		objType = objType.Elem()
		depth++
	}

	// Create a new object of the given type (returns a pointer).
	obj := reflect.New(objType)

	// Add the necessary number of pointers to the object.
	// For each additional pointer level, create a new pointer to the current pointer.
	for i := 0; i < (depth - 1); i++ {
		// Create a new pointer to the current pointer type
		ptrToPtr := reflect.New(obj.Type())
		// Set the value of the new pointer to point to the current object
		ptrToPtr.Elem().Set(obj)
		obj = ptrToPtr
	}

	return obj.Interface()
}

// Unpacks the value if it is a pointer.
func UnpackValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return reflect.Value{}
		}

		v = v.Elem()
	}

	return v
}
