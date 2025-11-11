package dotpath

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

// Extract names from tags.
func namesFromTags(f reflect.StructField) []string {
	names := make([]string, 0, 2)

	for _, tag := range []string{"json", "yaml"} {
		tag := strings.SplitN(f.Tag.Get(tag), ",", 2)
		if len(tag) > 0 && tag[0] != "" {
			names = append(names, tag[0])
		}
	}

	return names
}

// Returns the field with the given name (case-insensitive).
func structField(s reflect.Value, n string) (reflect.Value, error) {
	for i := 0; i < s.NumField(); i++ {
		// Take the name from the struct field.
		fieldType := s.Type().Field(i)
		names := []string{fieldType.Name}

		// Extract names from tags.
		tagNames := namesFromTags(fieldType)
		names = append(names, tagNames...)

		// Compare the names with the given name.
		for _, name := range names {
			if strings.EqualFold(name, n) {
				return s.Field(i), nil
			}
		}
	}

	return reflect.Value{}, fmt.Errorf("field not found: %s", n)
}

// Returns the value at the given path.
func getValue(v reflect.Value, p string) (reflect.Value, error) {
	parts := strings.Split(p, ".")

	// Traverse the path.
	for len(parts) > 0 {
		// If the value is a pointer, dereference it.
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Struct:
			var err error
			v, err = structField(v, parts[0])
			if err != nil {
				return reflect.Value{}, fmt.Errorf("failed to get field: %w", err)
			}
		case reflect.Array, reflect.Slice:
			index, err := strconv.Atoi(parts[0])
			if err != nil {
				return reflect.Value{}, fmt.Errorf("invalid index: %s", parts[0])
			}

			v = v.Index(index)
		default:
			return reflect.Value{}, fmt.Errorf("unsupported type: %s", v.Kind())
		}

		// Pop the first part of the path.
		parts = parts[1:]
	}

	return v, nil
}

func setValue(v reflect.Value, value any) error {
	// If the value is a pointer, dereference it.
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// If the value is not settable, return an error.
	if !v.CanSet() {
		return fmt.Errorf("value is not settable")
	}

	// If the value is a json.Unmarshaler, use it to unmarshal the value.
	unmarshaler, ok := v.Addr().Interface().(json.Unmarshaler)
	if ok {
		b, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}

		err = unmarshaler.UnmarshalJSON(b)
		if err != nil {
			return fmt.Errorf("failed to unmarshal value: %w", err)
		}
		return nil
	}

	// Handle basic types.
	switch v.Kind() {
	case reflect.String:
		c, err := cast.ToStringE(value)
		if err != nil {
			return fmt.Errorf("failed to cast value: %w", err)
		}

		v.SetString(c)
	case reflect.Bool:
		c, err := cast.ToBoolE(value)
		if err != nil {
			return fmt.Errorf("failed to cast value: %w", err)
		}

		v.SetBool(c)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		c, err := cast.ToInt64E(value)
		if err != nil {
			return fmt.Errorf("failed to cast value: %w", err)
		}

		v.SetInt(c)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		c, err := cast.ToUint64E(value)
		if err != nil {
			return fmt.Errorf("failed to cast value: %w", err)
		}

		v.SetUint(c)
	case reflect.Float32, reflect.Float64:
		c, err := cast.ToFloat64E(value)
		if err != nil {
			return fmt.Errorf("failed to cast value: %w", err)
		}

		v.SetFloat(c)
	default:
		return fmt.Errorf("unsupported type: %s", v.Kind())
	}

	return nil
}
