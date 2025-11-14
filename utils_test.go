package confless_test

import (
	"reflect"
	"testing"

	"github.com/codetent/confless"
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
			got := confless.MakeNewObject(tt.t)
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
