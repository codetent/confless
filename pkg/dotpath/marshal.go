package dotpath

import (
	"encoding"
	"fmt"
	"reflect"
)

// Unmarshal the value using a custom unmarshaler.
func unmarshalText(v reflect.Value, value any) error {
	// Convert the value to bytes.
	valueBytes, ok := value.([]byte)
	if !ok {
		valueString, ok := value.(string)
		if !ok {
			return fmt.Errorf("%w: %s", ErrUnsupportedType, v.Kind())
		}

		valueBytes = []byte(valueString)
	}

	// Check if the value implements TextUnmarshaler.
	unmarshaler, ok := v.Addr().Interface().(encoding.TextUnmarshaler)
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnsupportedType, v.Kind())
	}

	// Unmarshal the value.
	err := unmarshaler.UnmarshalText(valueBytes)
	if err != nil {
		return fmt.Errorf("%w: failed to unmarshal value: %w", ErrInvalidValue, err)
	}

	return nil
}
