package confless

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"reflect"
	"strings"

	"dario.cat/mergo"
	"github.com/goccy/go-yaml"

	"github.com/codetent/confless/pkg/dotpath"
)

var (
	ErrObjectNotAPointer = errors.New("object is not a pointer")
)

// Populate the object by the given flags.
// Names are converted to dot-separated paths (e.g. "my-flag" -> "my.flag").
func populateByFlags(fset *flag.FlagSet, obj any) error {
	// Check if the object is a pointer.
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return ErrObjectNotAPointer
	}

	fset.Visit(func(f *flag.Flag) {
		// Replace the dash in the key with a dot.
		key := strings.ReplaceAll(f.Name, "-", ".")

		// Set the value at the given path.
		_ = dotpath.Set(obj, key, f.Value.String())
	})

	return nil
}

// Populate the object by environment variables with the given prefix.
// Overrides existing values only if set in the environment variables.
// Names are converted to dot-separated paths (e.g. "MY_FLAG" -> "my.flag").
func populateByEnv(env []string, pre string, obj any) error {
	// Check if the object is a pointer.
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return ErrObjectNotAPointer
	}

	prefix := strings.ToLower(pre) + "_"

	for _, env := range env {
		// Split the environment variable into key and value.
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		// Remove the prefix from the key.
		key := strings.ToLower(parts[0])
		key, ok := strings.CutPrefix(key, prefix)
		if !ok {
			continue
		}

		// Replace the underscore in the key with a dot.
		path := strings.ReplaceAll(key, "_", ".")

		// Set the value at the given path.
		err := dotpath.Set(obj, path, parts[1])
		if err != nil {
			return fmt.Errorf("failed to set path: %w", err)
		}
	}

	return nil
}

// Populate the object by a file with the given path and format.
// Overrides existing values only if set in the file.
func populateByFile(r io.Reader, format string, obj any) error {
	// Check if the object is a pointer.
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return ErrObjectNotAPointer
	}

	// Create a new object of the same type as the given object.
	decoded := MakeNewObject(reflect.TypeOf(obj))

	// Unmarshal the file based on the format.
	switch format {
	case "json":
		err := json.NewDecoder(r).Decode(decoded)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json: %w", err)
		}
	case "yaml":
		err := yaml.NewDecoder(r).Decode(decoded)
		if err != nil {
			return fmt.Errorf("failed to unmarshal yaml: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Merge the decoded object into the given object.
	err := mergo.Merge(obj, decoded, mergo.WithOverride)
	if err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}
