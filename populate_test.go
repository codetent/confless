package confless

import (
	"flag"
	"io"
	"strings"
	"testing"
)

func Test_populateByFlags(t *testing.T) {
	tests := []struct {
		name    string
		fset    *flag.FlagSet
		obj     any
		wantErr bool
		verify  func(t *testing.T, obj any)
	}{
		{
			name: "error when object is not a pointer",
			fset: flag.NewFlagSet("test", flag.ContinueOnError),
			obj: struct {
				Name string
			}{},
			wantErr: true,
		},
		{
			name: "populate string field",
			fset: func() *flag.FlagSet {
				fset := flag.NewFlagSet("test", flag.ContinueOnError)
				fset.String("name", "", "name flag")
				_ = fset.Parse([]string{"--name=MyApp"})
				return fset
			}(),
			obj: &struct {
				Name string
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Name string })
				if cfg.Name != "MyApp" {
					t.Errorf("expected Name to be 'MyApp', got '%s'", cfg.Name)
				}
			},
		},
		{
			name: "populate int field",
			fset: func() *flag.FlagSet {
				fset := flag.NewFlagSet("test", flag.ContinueOnError)
				fset.String("port", "", "port flag")
				_ = fset.Parse([]string{"--port=8080"})
				return fset
			}(),
			obj: &struct {
				Port int
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Port int })
				if cfg.Port != 8080 {
					t.Errorf("expected Port to be 8080, got %d", cfg.Port)
				}
			},
		},
		{
			name: "populate bool field",
			fset: func() *flag.FlagSet {
				fset := flag.NewFlagSet("test", flag.ContinueOnError)
				fset.String("debug", "", "debug flag")
				_ = fset.Parse([]string{"--debug=true"})
				return fset
			}(),
			obj: &struct {
				Debug bool
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Debug bool })
				if !cfg.Debug {
					t.Errorf("expected Debug to be true, got %v", cfg.Debug)
				}
			},
		},
		{
			name: "populate nested field with dash notation",
			fset: func() *flag.FlagSet {
				fset := flag.NewFlagSet("test", flag.ContinueOnError)
				fset.String("database-host", "", "database host flag")
				_ = fset.Parse([]string{"--database-host=localhost"})
				return fset
			}(),
			obj: &struct {
				Database struct {
					Host string
				}
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Database struct {
						Host string
					}
				})
				if cfg.Database.Host != "localhost" {
					t.Errorf("expected Database.Host to be 'localhost', got '%s'", cfg.Database.Host)
				}
			},
		},
		{
			name: "populate array index with dash notation",
			fset: func() *flag.FlagSet {
				fset := flag.NewFlagSet("test", flag.ContinueOnError)
				fset.String("items-0", "", "items[0] flag")
				_ = fset.Parse([]string{"--items-0=42"})
				return fset
			}(),
			obj: &struct {
				Items []int
			}{
				Items: []int{0, 0},
			},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Items []int })
				if len(cfg.Items) != 2 {
					t.Errorf("expected Items length to be 2, got %d", len(cfg.Items))
				}
				if cfg.Items[0] != 42 {
					t.Errorf("expected Items[0] to be 42, got %d", cfg.Items[0])
				}
			},
		},
		{
			name: "only visited flags are processed",
			fset: func() *flag.FlagSet {
				fset := flag.NewFlagSet("test", flag.ContinueOnError)
				fset.String("name", "default", "name flag")
				fset.String("port", "8080", "port flag")
				// Parse but don't set any flags
				_ = fset.Parse([]string{})
				return fset
			}(),
			obj: &struct {
				Name string
				Port int
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Name string
					Port int
				})
				if cfg.Name != "" {
					t.Errorf("expected Name to be empty (flag not visited), got '%s'", cfg.Name)
				}
				if cfg.Port != 0 {
					t.Errorf("expected Port to be 0 (flag not visited), got %d", cfg.Port)
				}
			},
		},
		{
			name: "populate uint field",
			fset: func() *flag.FlagSet {
				fset := flag.NewFlagSet("test", flag.ContinueOnError)
				fset.String("count", "", "count flag")
				_ = fset.Parse([]string{"--count=100"})
				return fset
			}(),
			obj: &struct {
				Count uint
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Count uint })
				if cfg.Count != 100 {
					t.Errorf("expected Count to be 100, got %d", cfg.Count)
				}
			},
		},
		{
			name: "populate float field",
			fset: func() *flag.FlagSet {
				fset := flag.NewFlagSet("test", flag.ContinueOnError)
				fset.String("ratio", "", "ratio flag")
				_ = fset.Parse([]string{"--ratio=3.14"})
				return fset
			}(),
			obj: &struct {
				Ratio float64
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Ratio float64 })
				if cfg.Ratio != 3.14 {
					t.Errorf("expected Ratio to be 3.14, got %f", cfg.Ratio)
				}
			},
		},
		{
			name: "populate multiple nested fields",
			fset: func() *flag.FlagSet {
				fset := flag.NewFlagSet("test", flag.ContinueOnError)
				fset.String("database-host", "", "database host")
				fset.String("database-port", "", "database port")
				_ = fset.Parse([]string{"--database-host=localhost", "--database-port=5432"})
				return fset
			}(),
			obj: &struct {
				Database struct {
					Host string
					Port int
				}
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Database struct {
						Host string
						Port int
					}
				})
				if cfg.Database.Host != "localhost" {
					t.Errorf("expected Database.Host to be 'localhost', got '%s'", cfg.Database.Host)
				}
				if cfg.Database.Port != 5432 {
					t.Errorf("expected Database.Port to be 5432, got %d", cfg.Database.Port)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := populateByFlags(tt.fset, tt.obj)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("populateByFlags() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("populateByFlags() succeeded unexpectedly")
			}
			if tt.verify != nil {
				tt.verify(t, tt.obj)
			}
		})
	}
}

func Test_populateByEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     []string
		pre     string
		obj     any
		wantErr bool
		verify  func(t *testing.T, obj any)
	}{
		{
			name: "error when object is not a pointer",
			env:  []string{"APP_NAME=MyApp"},
			pre:  "APP",
			obj: struct {
				Name string
			}{},
			wantErr: true,
		},
		{
			name: "populate string field with prefix",
			env:  []string{"APP_NAME=MyApp"},
			pre:  "APP",
			obj: &struct {
				Name string
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Name string })
				if cfg.Name != "MyApp" {
					t.Errorf("expected Name to be 'MyApp', got '%s'", cfg.Name)
				}
			},
		},
		{
			name: "populate int field with prefix",
			env:  []string{"APP_PORT=8080"},
			pre:  "APP",
			obj: &struct {
				Port int
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Port int })
				if cfg.Port != 8080 {
					t.Errorf("expected Port to be 8080, got %d", cfg.Port)
				}
			},
		},
		{
			name: "populate bool field with prefix",
			env:  []string{"APP_DEBUG=true"},
			pre:  "APP",
			obj: &struct {
				Debug bool
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Debug bool })
				if !cfg.Debug {
					t.Errorf("expected Debug to be true, got %v", cfg.Debug)
				}
			},
		},
		{
			name: "populate nested field with underscore notation",
			env:  []string{"APP_DATABASE_HOST=localhost"},
			pre:  "APP",
			obj: &struct {
				Database struct {
					Host string
				}
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Database struct {
						Host string
					}
				})
				if cfg.Database.Host != "localhost" {
					t.Errorf("expected Database.Host to be 'localhost', got '%s'", cfg.Database.Host)
				}
			},
		},
		{
			name: "populate array index with underscore notation",
			env:  []string{"APP_ITEMS_0=42"},
			pre:  "APP",
			obj: &struct {
				Items []int
			}{
				Items: []int{0, 0},
			},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Items []int })
				if len(cfg.Items) != 2 {
					t.Errorf("expected Items length to be 2, got %d", len(cfg.Items))
				}
				if cfg.Items[0] != 42 {
					t.Errorf("expected Items[0] to be 42, got %d", cfg.Items[0])
				}
			},
		},
		{
			name: "ignore env vars without prefix",
			env:  []string{"OTHER_NAME=Other", "APP_NAME=MyApp"},
			pre:  "APP",
			obj: &struct {
				Name string
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Name string })
				if cfg.Name != "MyApp" {
					t.Errorf("expected Name to be 'MyApp', got '%s'", cfg.Name)
				}
			},
		},
		{
			name: "case-insensitive prefix matching",
			env:  []string{"app_name=MyApp"},
			pre:  "APP",
			obj: &struct {
				Name string
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Name string })
				if cfg.Name != "MyApp" {
					t.Errorf("expected Name to be 'MyApp', got '%s'", cfg.Name)
				}
			},
		},
		{
			name: "populate multiple nested fields",
			env:  []string{"APP_DATABASE_HOST=localhost", "APP_DATABASE_PORT=5432"},
			pre:  "APP",
			obj: &struct {
				Database struct {
					Host string
					Port int
				}
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Database struct {
						Host string
						Port int
					}
				})
				if cfg.Database.Host != "localhost" {
					t.Errorf("expected Database.Host to be 'localhost', got '%s'", cfg.Database.Host)
				}
				if cfg.Database.Port != 5432 {
					t.Errorf("expected Database.Port to be 5432, got %d", cfg.Database.Port)
				}
			},
		},
		{
			name: "populate uint field",
			env:  []string{"APP_COUNT=100"},
			pre:  "APP",
			obj: &struct {
				Count uint
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Count uint })
				if cfg.Count != 100 {
					t.Errorf("expected Count to be 100, got %d", cfg.Count)
				}
			},
		},
		{
			name: "populate float field",
			env:  []string{"APP_RATIO=3.14"},
			pre:  "APP",
			obj: &struct {
				Ratio float64
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Ratio float64 })
				if cfg.Ratio != 3.14 {
					t.Errorf("expected Ratio to be 3.14, got %f", cfg.Ratio)
				}
			},
		},
		{
			name: "ignore invalid env var format",
			env:  []string{"APP_NAME", "APP_PORT=8080"},
			pre:  "APP",
			obj: &struct {
				Name string
				Port int
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Name string
					Port int
				})
				if cfg.Name != "" {
					t.Errorf("expected Name to be empty (invalid env var), got '%s'", cfg.Name)
				}
				if cfg.Port != 8080 {
					t.Errorf("expected Port to be 8080, got %d", cfg.Port)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := populateByEnv(tt.env, tt.pre, tt.obj)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("populateByEnv() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("populateByEnv() succeeded unexpectedly")
			}
			if tt.verify != nil {
				tt.verify(t, tt.obj)
			}
		})
	}
}

func Test_populateByFile(t *testing.T) {
	tests := []struct {
		name    string
		r       io.Reader
		format  string
		obj     any
		wantErr bool
		verify  func(t *testing.T, obj any)
	}{
		{
			name:   "error when object is not a pointer",
			r:      strings.NewReader(`{"name": "MyApp"}`),
			format: "json",
			obj: struct {
				Name string
			}{},
			wantErr: true,
		},
		{
			name:   "populate from JSON file",
			r:      strings.NewReader(`{"name": "MyApp", "port": 8080}`),
			format: "json",
			obj: &struct {
				Name string
				Port int
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Name string
					Port int
				})
				if cfg.Name != "MyApp" {
					t.Errorf("expected Name to be 'MyApp', got '%s'", cfg.Name)
				}
				if cfg.Port != 8080 {
					t.Errorf("expected Port to be 8080, got %d", cfg.Port)
				}
			},
		},
		{
			name:   "populate from YAML file",
			r:      strings.NewReader("name: MyApp\nport: 8080"),
			format: "yaml",
			obj: &struct {
				Name string
				Port int
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Name string
					Port int
				})
				if cfg.Name != "MyApp" {
					t.Errorf("expected Name to be 'MyApp', got '%s'", cfg.Name)
				}
				if cfg.Port != 8080 {
					t.Errorf("expected Port to be 8080, got %d", cfg.Port)
				}
			},
		},
		{
			name:   "merge with existing values",
			r:      strings.NewReader(`{"port": 9000}`),
			format: "json",
			obj: &struct {
				Name string
				Port int
			}{
				Name: "DefaultApp",
				Port: 8080,
			},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Name string
					Port int
				})
				if cfg.Name != "DefaultApp" {
					t.Errorf("expected Name to remain 'DefaultApp', got '%s'", cfg.Name)
				}
				if cfg.Port != 9000 {
					t.Errorf("expected Port to be 9000 (overridden), got %d", cfg.Port)
				}
			},
		},
		{
			name:   "populate nested structure from JSON",
			r:      strings.NewReader(`{"database": {"host": "localhost", "port": 5432}}`),
			format: "json",
			obj: &struct {
				Database struct {
					Host string
					Port int
				}
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Database struct {
						Host string
						Port int
					}
				})
				if cfg.Database.Host != "localhost" {
					t.Errorf("expected Database.Host to be 'localhost', got '%s'", cfg.Database.Host)
				}
				if cfg.Database.Port != 5432 {
					t.Errorf("expected Database.Port to be 5432, got %d", cfg.Database.Port)
				}
			},
		},
		{
			name:   "populate nested structure from YAML",
			r:      strings.NewReader("database:\n  host: localhost\n  port: 5432"),
			format: "yaml",
			obj: &struct {
				Database struct {
					Host string
					Port int
				}
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Database struct {
						Host string
						Port int
					}
				})
				if cfg.Database.Host != "localhost" {
					t.Errorf("expected Database.Host to be 'localhost', got '%s'", cfg.Database.Host)
				}
				if cfg.Database.Port != 5432 {
					t.Errorf("expected Database.Port to be 5432, got %d", cfg.Database.Port)
				}
			},
		},
		{
			name:   "error for unsupported format",
			r:      strings.NewReader(`{"name": "MyApp"}`),
			format: "xml",
			obj: &struct {
				Name string
			}{},
			wantErr: true,
		},
		{
			name:   "error for invalid JSON",
			r:      strings.NewReader(`{"name": "MyApp"`),
			format: "json",
			obj: &struct {
				Name string
			}{},
			wantErr: true,
		},
		{
			name:   "error for invalid YAML",
			r:      strings.NewReader("name: [invalid"),
			format: "yaml",
			obj: &struct {
				Name string
			}{},
			wantErr: true,
		},
		{
			name:   "populate array from JSON",
			r:      strings.NewReader(`{"items": [1, 2, 3]}`),
			format: "json",
			obj: &struct {
				Items []int
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Items []int })
				if len(cfg.Items) != 3 {
					t.Errorf("expected Items length to be 3, got %d", len(cfg.Items))
				}
				if cfg.Items[0] != 1 || cfg.Items[1] != 2 || cfg.Items[2] != 3 {
					t.Errorf("expected Items to be [1, 2, 3], got %v", cfg.Items)
				}
			},
		},
		{
			name:   "populate with json tag",
			r:      strings.NewReader(`{"config_file": "production.json"}`),
			format: "json",
			obj: &struct {
				ConfigFile string `json:"config_file"`
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					ConfigFile string `json:"config_file"`
				})
				if cfg.ConfigFile != "production.json" {
					t.Errorf("expected ConfigFile to be 'production.json', got '%s'", cfg.ConfigFile)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := populateByFile(tt.r, tt.format, tt.obj)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("populateByFile() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("populateByFile() succeeded unexpectedly")
			}
			if tt.verify != nil {
				tt.verify(t, tt.obj)
			}
		})
	}
}
