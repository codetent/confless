package dotpath

import (
	"fmt"
	"reflect"
)

// Get the value at the given path of the object.
func Get(obj any, p string) (any, error) {
	refField, err := getValue(reflect.ValueOf(obj), p)
	if err != nil {
		return nil, fmt.Errorf("failed to get field: %w", err)
	}

	return refField.Interface(), nil
}

// Set the value at the given path of the object.
func Set(obj any, p string, v any) error {
	refField, err := getValue(reflect.ValueOf(obj), p)
	if err != nil {
		return fmt.Errorf("failed to get field: %w", err)
	}

	err = setValue(refField, v)
	if err != nil {
		return fmt.Errorf("failed to set field: %w", err)
	}

	return nil
}
