package dotpath

import (
	"encoding/json"
	"reflect"
	"testing"
)

// CustomUnmarshaler is a custom type that implements json.Unmarshaler for testing
type CustomUnmarshaler struct {
	Value string
}

func (c *CustomUnmarshaler) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	c.Value = "unmarshaled:" + s
	return nil
}

func Test_structField(t *testing.T) {
	type TestStruct struct {
		Name     string
		Age      int    `json:"age"`
		Email    string `json:"email_address,omitempty"`
		IsActive bool   `json:"isActive,omitempty"`
	}

	tests := []struct {
		name     string
		s        reflect.Value
		n        string
		wantErr  bool
		validate func(t *testing.T, got reflect.Value)
	}{
		{
			name: "find field by struct name (exact match)",
			s: reflect.ValueOf(TestStruct{
				Name: "John",
			}),
			n: "Name",
			validate: func(t *testing.T, got reflect.Value) {
				if got.String() != "John" {
					t.Errorf("got %v, want John", got.String())
				}
			},
		},
		{
			name: "find field by struct name (case-insensitive)",
			s: reflect.ValueOf(TestStruct{
				Name: "John",
			}),
			n: "name",
			validate: func(t *testing.T, got reflect.Value) {
				if got.String() != "John" {
					t.Errorf("got %v, want John", got.String())
				}
			},
		},
		{
			name: "find field by JSON tag",
			s: reflect.ValueOf(TestStruct{
				Email: "test@example.com",
			}),
			n: "email_address",
			validate: func(t *testing.T, got reflect.Value) {
				if got.String() != "test@example.com" {
					t.Errorf("got %v, want test@example.com", got.String())
				}
			},
		},
		{
			name: "find field by JSON tag (case-insensitive)",
			s: reflect.ValueOf(TestStruct{
				Email: "test@example.com",
			}),
			n: "EMAIL_ADDRESS",
			validate: func(t *testing.T, got reflect.Value) {
				if got.String() != "test@example.com" {
					t.Errorf("got %v, want test@example.com", got.String())
				}
			},
		},
		{
			name:    "field not found",
			s:       reflect.ValueOf(TestStruct{}),
			n:       "NotFound",
			wantErr: true,
		},
		{
			name:    "empty struct",
			s:       reflect.ValueOf(struct{}{}),
			n:       "anything",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := structField(tt.s, tt.n)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("structField() failed: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("structField() succeeded unexpectedly")
			}
			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func Test_getValue(t *testing.T) {
	type Nested struct {
		Value string
	}

	type TestStruct struct {
		Name      string
		Age       int
		Nested    Nested
		PtrNested *Nested
		Items     []int
		Items2    []Nested
		Array     [3]string
		MapStr    map[string]string
		MapNested map[string]Nested
	}

	tests := []struct {
		name     string
		v        reflect.Value
		p        string
		wantErr  bool
		validate func(t *testing.T, got reflect.Value)
	}{
		{
			name: "simple struct field",
			v: reflect.ValueOf(TestStruct{
				Name: "John",
			}),
			p: "Name",
			validate: func(t *testing.T, got reflect.Value) {
				if got.String() != "John" {
					t.Errorf("got %v, want John", got.String())
				}
			},
		},
		{
			name: "nested struct field",
			v: reflect.ValueOf(TestStruct{
				Nested: Nested{
					Value: "test",
				},
			}),
			p: "Nested.Value",
			validate: func(t *testing.T, got reflect.Value) {
				if got.String() != "test" {
					t.Errorf("got %v, want test", got.String())
				}
			},
		},
		{
			name: "slice index",
			v: reflect.ValueOf(TestStruct{
				Items: []int{10, 20, 30},
			}),
			p: "Items.1",
			validate: func(t *testing.T, got reflect.Value) {
				if got.Int() != 20 {
					t.Errorf("got %v, want 20", got.Int())
				}
			},
		},
		{
			name: "array index",
			v: reflect.ValueOf(TestStruct{
				Array: [3]string{"a", "b", "c"},
			}),
			p: "Array.2",
			validate: func(t *testing.T, got reflect.Value) {
				if got.String() != "c" {
					t.Errorf("got %v, want c", got.String())
				}
			},
		},
		{
			name: "slice of structs",
			v: reflect.ValueOf(TestStruct{
				Items2: []Nested{
					{
						Value: "first",
					},
					{
						Value: "second",
					},
				},
			}),
			p: "Items2.1.Value",
			validate: func(t *testing.T, got reflect.Value) {
				if got.String() != "second" {
					t.Errorf("got %v, want second", got.String())
				}
			},
		},
		{
			name: "pointer dereference",
			v: reflect.ValueOf(TestStruct{
				PtrNested: &Nested{
					Value: "ptr",
				},
			}),
			p: "PtrNested.Value",
			validate: func(t *testing.T, got reflect.Value) {
				if got.String() != "ptr" {
					t.Errorf("got %v, want ptr", got.String())
				}
			},
		},
		{
			name: "nil pointer",
			v: reflect.ValueOf(TestStruct{
				PtrNested: nil,
			}),
			p:       "PtrNested.Value",
			wantErr: true,
		},
		{
			name:    "invalid field",
			v:       reflect.ValueOf(TestStruct{}),
			p:       "NotFound",
			wantErr: true,
		},
		{
			name: "invalid index (non-numeric)",
			v: reflect.ValueOf(TestStruct{
				Items: []int{1, 2, 3},
			}),
			p:       "Items.invalid",
			wantErr: true,
		},
		{
			name:    "empty path",
			v:       reflect.ValueOf(TestStruct{}),
			p:       "",
			wantErr: true,
		},
		{
			name: "simple map access (string key)",
			v: reflect.ValueOf(TestStruct{
				MapStr: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			}),
			p: "MapStr.key1",
			validate: func(t *testing.T, got reflect.Value) {
				if got.String() != "value1" {
					t.Errorf("got %v, want value1", got.String())
				}
			},
		},
		{
			name: "nested map access (map value field)",
			v: reflect.ValueOf(TestStruct{
				MapNested: map[string]Nested{
					"first":  {Value: "first_value"},
					"second": {Value: "second_value"},
				},
			}),
			p: "MapNested.second.Value",
			validate: func(t *testing.T, got reflect.Value) {
				if got.String() != "second_value" {
					t.Errorf("got %v, want second_value", got.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValue(tt.v, tt.p)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("getValue() failed: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("getValue() succeeded unexpectedly")
			}
			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func Test_setValue(t *testing.T) {
	tests := []struct {
		name     string
		v        reflect.Value
		value    any
		wantErr  bool
		validate func(t *testing.T, v reflect.Value)
	}{
		{
			name:  "set string",
			v:     reflect.ValueOf(new(string)).Elem(),
			value: "test",
			validate: func(t *testing.T, v reflect.Value) {
				if v.String() != "test" {
					t.Errorf("got %v, want test", v.String())
				}
			},
		},
		{
			name:  "set bool",
			v:     reflect.ValueOf(new(bool)).Elem(),
			value: true,
			validate: func(t *testing.T, v reflect.Value) {
				if v.Bool() != true {
					t.Errorf("got %v, want true", v.Bool())
				}
			},
		},
		{
			name:  "set bool from string",
			v:     reflect.ValueOf(new(bool)).Elem(),
			value: "true",
			validate: func(t *testing.T, v reflect.Value) {
				if v.Bool() != true {
					t.Errorf("got %v, want true", v.Bool())
				}
			},
		},
		{
			name:    "set bool (invalid)",
			v:       reflect.ValueOf(new(bool)).Elem(),
			value:   "invalid",
			wantErr: true,
		},
		{
			name:  "set int",
			v:     reflect.ValueOf(new(int)).Elem(),
			value: 42,
			validate: func(t *testing.T, v reflect.Value) {
				if v.Int() != 42 {
					t.Errorf("got %v, want 42", v.Int())
				}
			},
		},
		{
			name:  "set int from string",
			v:     reflect.ValueOf(new(int)).Elem(),
			value: "42",
			validate: func(t *testing.T, v reflect.Value) {
				if v.Int() != 42 {
					t.Errorf("got %v, want 42", v.Int())
				}
			},
		},
		{
			name:    "set int (invalid)",
			v:       reflect.ValueOf(new(int)).Elem(),
			value:   "not a number",
			wantErr: true,
		},
		{
			name:  "set uint",
			v:     reflect.ValueOf(new(uint)).Elem(),
			value: 42,
			validate: func(t *testing.T, v reflect.Value) {
				if v.Uint() != 42 {
					t.Errorf("got %v, want 42", v.Uint())
				}
			},
		},
		{
			name:  "set uint from string",
			v:     reflect.ValueOf(new(uint)).Elem(),
			value: "42",
			validate: func(t *testing.T, v reflect.Value) {
				if v.Uint() != 42 {
					t.Errorf("got %v, want 42", v.Uint())
				}
			},
		},
		{
			name:    "set uint (invalid)",
			v:       reflect.ValueOf(new(uint)).Elem(),
			value:   "not a number",
			wantErr: true,
		},
		{
			name:  "set float",
			v:     reflect.ValueOf(new(float64)).Elem(),
			value: 3.14159,
			validate: func(t *testing.T, v reflect.Value) {
				if v.Float() != 3.14159 {
					t.Errorf("got %v, want 3.14159", v.Float())
				}
			},
		},
		{
			name:  "set float from string",
			v:     reflect.ValueOf(new(float64)).Elem(),
			value: "3.14159",
			validate: func(t *testing.T, v reflect.Value) {
				if v.Float() != 3.14159 {
					t.Errorf("got %v, want 3.14159", v.Float())
				}
			},
		},
		{
			name:    "set float (invalid)",
			v:       reflect.ValueOf(new(float64)).Elem(),
			value:   "not a number",
			wantErr: true,
		},
		{
			name:  "json.Unmarshaler",
			v:     reflect.ValueOf(new(CustomUnmarshaler)).Elem(),
			value: "test",
			validate: func(t *testing.T, v reflect.Value) {
				c := v.Interface().(CustomUnmarshaler)
				if c.Value != "unmarshaled:test" {
					t.Errorf("got %v, want unmarshaled:test", c.Value)
				}
			},
		},
		{
			name:    "unsettable value",
			v:       reflect.ValueOf("not settable"),
			value:   "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setValue(tt.v, tt.value)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("setValue() failed: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("setValue() succeeded unexpectedly")
			}
			if tt.validate != nil {
				tt.validate(t, tt.v)
			}
		})
	}
}
