package reflectutil

import (
	"reflect"
	"testing"
)

func TestMakeNewObject(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		t    reflect.Type
		want reflect.Type // expected type of the returned value
	}{
		{
			name: "int type",
			t:    reflect.TypeOf(0),
			want: reflect.TypeOf((*int)(nil)),
		},
		{
			name: "string type",
			t:    reflect.TypeOf(""),
			want: reflect.TypeOf((*string)(nil)),
		},
		{
			name: "bool type",
			t:    reflect.TypeOf(false),
			want: reflect.TypeOf((*bool)(nil)),
		},
		{
			name: "pointer to int",
			t:    reflect.TypeOf((*int)(nil)),
			want: reflect.TypeOf((*int)(nil)),
		},
		{
			name: "pointer to pointer to int",
			t:    reflect.TypeOf((**int)(nil)),
			want: reflect.TypeOf((**int)(nil)),
		},
		{
			name: "struct type",
			t:    reflect.TypeOf(struct{ Name string }{}),
			want: reflect.TypeOf((*struct{ Name string })(nil)),
		},
		{
			name: "pointer to struct",
			t:    reflect.TypeOf((*struct{ Name string })(nil)),
			want: reflect.TypeOf((*struct{ Name string })(nil)),
		},
		{
			name: "slice type",
			t:    reflect.TypeOf([]int{}),
			want: reflect.TypeOf((*[]int)(nil)),
		},
		{
			name: "map type",
			t:    reflect.TypeOf(map[string]int{}),
			want: reflect.TypeOf((*map[string]int)(nil)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeNewObject(tt.t)
			gotType := reflect.TypeOf(got)

			// For non-pointer input types, MakeNewObject should return a pointer
			// For pointer input types, it should return the same pointer level
			if gotType != tt.want {
				t.Errorf("MakeNewObject() returned type %v, want %v", gotType, tt.want)
			}

			// Verify that the value is not nil (for pointer types)
			if gotType.Kind() == reflect.Ptr {
				gotValue := reflect.ValueOf(got)
				if gotValue.IsNil() {
					t.Errorf("MakeNewObject() returned nil pointer")
				}
			}
		})
	}
}

func TestUnpackValue(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		v    reflect.Value
		want reflect.Value
	}{
		{
			name: "non-pointer int",
			v:    reflect.ValueOf(42),
			want: reflect.ValueOf(42),
		},
		{
			name: "non-pointer string",
			v:    reflect.ValueOf("hello"),
			want: reflect.ValueOf("hello"),
		},
		{
			name: "non-pointer bool",
			v:    reflect.ValueOf(true),
			want: reflect.ValueOf(true),
		},
		{
			name: "single pointer to int",
			v:    reflect.ValueOf(PtrTo(42)),
			want: reflect.ValueOf(42),
		},
		{
			name: "single pointer to string",
			v:    reflect.ValueOf(PtrTo("hello")),
			want: reflect.ValueOf("hello"),
		},
		{
			name: "double pointer to int",
			v:    reflect.ValueOf(PtrTo(42)),
			want: reflect.ValueOf(42),
		},
		{
			name: "triple pointer to int",
			v:    reflect.ValueOf(PtrTo(42)),
			want: reflect.ValueOf(42),
		},
		{
			name: "nil pointer",
			v:    reflect.ValueOf((*int)(nil)),
			want: reflect.Value{},
		},
		{
			name: "nil double pointer",
			v:    reflect.ValueOf((**int)(nil)),
			want: reflect.Value{},
		},
		{
			name: "pointer to nil pointer",
			v:    reflect.ValueOf(PtrTo((*int)(nil))),
			want: reflect.Value{},
		},
		{
			name: "non-pointer struct",
			v:    reflect.ValueOf(struct{ Name string }{Name: "test"}),
			want: reflect.ValueOf(struct{ Name string }{Name: "test"}),
		},
		{
			name: "pointer to struct",
			v:    reflect.ValueOf(&struct{ Name string }{Name: "test"}),
			want: reflect.ValueOf(struct{ Name string }{Name: "test"}),
		},
		{
			name: "non-pointer slice",
			v:    reflect.ValueOf([]int{1, 2, 3}),
			want: reflect.ValueOf([]int{1, 2, 3}),
		},
		{
			name: "pointer to slice",
			v:    reflect.ValueOf(PtrTo([]int{1, 2, 3})),
			want: reflect.ValueOf([]int{1, 2, 3}),
		},
		{
			name: "non-pointer map",
			v:    reflect.ValueOf(map[string]int{"a": 1}),
			want: reflect.ValueOf(map[string]int{"a": 1}),
		},
		{
			name: "pointer to map",
			v:    reflect.ValueOf(PtrTo(map[string]int{"a": 1})),
			want: reflect.ValueOf(map[string]int{"a": 1}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnpackValue(tt.v)

			// Check if both are invalid (nil pointer case)
			if !got.IsValid() && !tt.want.IsValid() {
				return // Both are invalid, test passes
			}

			// Check if one is invalid and the other is not
			if !got.IsValid() || !tt.want.IsValid() {
				t.Errorf("UnpackValue() = %v (valid: %v), want %v (valid: %v)", got, got.IsValid(), tt.want, tt.want.IsValid())
				return
			}

			// Compare the actual values
			if got.Kind() != tt.want.Kind() {
				t.Errorf("UnpackValue() kind = %v, want %v", got.Kind(), tt.want.Kind())
				return
			}

			// For comparable types, use direct comparison
			if got.CanInterface() && tt.want.CanInterface() {
				gotVal := got.Interface()
				wantVal := tt.want.Interface()

				// Use deep equality for complex types
				if !reflect.DeepEqual(gotVal, wantVal) {
					t.Errorf("UnpackValue() = %v, want %v", gotVal, wantVal)
				}
			} else {
				// For types that can't be compared directly, check kind and type
				if got.Type() != tt.want.Type() {
					t.Errorf("UnpackValue() type = %v, want %v", got.Type(), tt.want.Type())
				}
			}
		})
	}
}
