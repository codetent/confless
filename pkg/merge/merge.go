package merge

import (
	"fmt"
	"reflect"

	"dario.cat/mergo"
)

// Transformer that detects zero values using the IsZero method (if defined).
// An example of a type that implements the IsZero method is time.Time.
type isZeroTransformer struct{}

func (t *isZeroTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	// Check if the type implements the IsZero method.
	isZero, hasIsZero := typ.MethodByName("IsZero")
	if !hasIsZero {
		return nil
	}

	return func(dst, src reflect.Value) error {
		// Check if the destination value is settable.
		if !dst.CanSet() {
			return nil
		}

		// Check if is zero and set the source value if it is.
		result := isZero.Func.Call([]reflect.Value{src})
		if result[0].Bool() {
			return nil
		}

		// Set the source value.
		dst.Set(src)

		return nil
	}
}

func Merge(dst any, src any) error {
	err := mergo.Merge(
		dst,
		src,
		mergo.WithOverride,
		mergo.WithTypeCheck,
		mergo.WithTransformers(&isZeroTransformer{}),
	)
	if err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}
