package confless

import (
	"iter"
	"reflect"
	"strings"

	"github.com/codetent/confless/pkg/reflectutil"
)

// Parses the tag into a map of key-value pairs.
// For example, the tag "file,format=yaml" will be parsed into:
//
//	{
//	  "file": "true",
//	  "format": "yaml",
//	}
func parseTag(t reflect.StructTag) map[string]string {
	tag := t.Get("confless")

	kvs := make(map[string]string)
	if tag == "" {
		return kvs
	}

	// Iterate over the parts of the tag separated by commas.
	for part := range strings.SplitSeq(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split the part into key and value.
		kv := strings.SplitN(part, "=", 2)
		if len(kv) < 2 {
			// If no value is provided, use "true" as the value.
			kv = append(kv, "true")
		}

		kvs[kv[0]] = kv[1]
	}

	return kvs
}

// Returns a sequence of file paths and formats found in the given object.
func findFileFields(o any) iter.Seq2[reflect.Value, string] {
	v := reflect.ValueOf(o)

	return func(yield func(reflect.Value, string) bool) {
		// If the value is not a struct, skip.
		v = reflectutil.UnpackValue(v)
		if v.Kind() != reflect.Struct {
			return
		}

		// Iterate over the fields of the struct.
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			value := reflectutil.UnpackValue(v.Field(i))

			// Parse field tag.
			kvs := parseTag(field.Tag)

			switch value.Kind() {
			case reflect.String:
				// If the field has a file tag, yield it.
				if kvs["file"] != "" {
					if !yield(value, kvs["format"]) {
						return
					}
				}
			case reflect.Struct:
				// Recursively find file fields in the nested struct.
				findFileFields(value.Interface())(yield)
			default:
				// Skip other field types.
				continue
			}
		}
	}
}
