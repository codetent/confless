package confless

import (
	"flag"
	"testing"

	"github.com/spf13/afero"
)

func Test_loader_RegisterEnv(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		opts    []loaderOption
		pre     string
		env     []string
		obj     any
		wantErr bool
		verify  func(t *testing.T, obj any)
	}{
		{
			name: "load from environment variables with prefix",
			opts: []loaderOption{
				WithEnvReader(func() []string {
					return []string{
						"APP_NAME=MyApp",
						"APP_DATABASE_HOST=localhost",
						"APP_DATABASE_PORT=5432",
						"OTHER_VAR=ignored",
					}
				}),
			},
			pre: "APP",
			obj: &struct {
				Name     string
				Database struct {
					Host string
					Port int
				}
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Name     string
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
			name: "empty prefix loads nothing",
			opts: []loaderOption{
				WithEnvReader(func() []string {
					return []string{"APP_NAME=MyApp"}
				}),
			},
			pre: "",
			obj: &struct {
				Name string
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Name string })
				if cfg.Name != "" {
					t.Errorf("expected Name to be empty, got '%s'", cfg.Name)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLoader(tt.opts...)
			l.RegisterEnv(tt.pre)
			err := l.Load(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.verify != nil {
				tt.verify(t, tt.obj)
			}
		})
	}
}

func Test_loader_RegisterFile(t *testing.T) {
	tests := []struct {
		name    string
		opts    []loaderOption
		path    string
		format  fileFormat
		content string
		obj     any
		wantErr bool
		verify  func(t *testing.T, obj any)
	}{
		{
			name: "load from JSON file",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "config.json", []byte(`{"name": "MyApp", "database": {"host": "localhost", "port": 5432}}`), 0644)
					return fs
				}()),
			},
			path:   "config.json",
			format: FileFormatJSON,
			obj: &struct {
				Name     string
				Database struct {
					Host string
					Port int
				}
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Name     string
					Database struct {
						Host string
						Port int
					}
				})
				if cfg.Name != "MyApp" {
					t.Errorf("expected Name to be 'MyApp', got '%s'", cfg.Name)
				}
				if cfg.Database.Host != "localhost" {
					t.Errorf("expected Database.Host to be 'localhost', got '%s'", cfg.Database.Host)
				}
				if cfg.Database.Port != 5432 {
					t.Errorf("expected Database.Port to be 5432, got %d", cfg.Database.Port)
				}
			},
		},
		{
			name: "load from YAML file",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "config.yaml", []byte("name: MyApp\nport: 8080"), 0644)
					return fs
				}()),
			},
			path:   "config.yaml",
			format: FileFormatYAML,
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
			name: "merge with existing values",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "config.json", []byte(`{"port": 9000}`), 0644)
					return fs
				}()),
			},
			path:   "config.json",
			format: FileFormatJSON,
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
			name: "skip missing file",
			opts: []loaderOption{
				WithFS(afero.NewMemMapFs()),
			},
			path:   "nonexistent.json",
			format: FileFormatJSON,
			obj: &struct {
				Name string
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Name string })
				if cfg.Name != "" {
					t.Errorf("expected Name to be empty, got '%s'", cfg.Name)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLoader(tt.opts...)
			l.RegisterFile(tt.path, tt.format)
			err := l.Load(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.verify != nil {
				tt.verify(t, tt.obj)
			}
		})
	}
}

func Test_loader_RegisterFileField(t *testing.T) {
	tests := []struct {
		name    string
		opts    []loaderOption
		field   string
		format  fileFormat
		obj     any
		wantErr bool
		verify  func(t *testing.T, obj any)
	}{
		{
			name: "load from file path in field",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "production.json", []byte(`{"name": "ProductionApp", "port": 9000}`), 0644)
					return fs
				}()),
			},
			field:  "ConfigFile",
			format: FileFormatJSON,
			obj: &struct {
				ConfigFile string
				Name       string
				Port       int
			}{
				ConfigFile: "production.json",
			},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					ConfigFile string
					Name       string
					Port       int
				})
				if cfg.Name != "ProductionApp" {
					t.Errorf("expected Name to be 'ProductionApp', got '%s'", cfg.Name)
				}
				if cfg.Port != 9000 {
					t.Errorf("expected Port to be 9000, got %d", cfg.Port)
				}
			},
		},
		{
			name: "error when field is not a string",
			opts: []loaderOption{
				WithFS(afero.NewMemMapFs()),
			},
			field:  "ConfigFile",
			format: FileFormatJSON,
			obj: &struct {
				ConfigFile int
			}{
				ConfigFile: 123,
			},
			wantErr: true,
		},
		{
			name: "skip missing file from field",
			opts: []loaderOption{
				WithFS(afero.NewMemMapFs()),
			},
			field:  "ConfigFile",
			format: FileFormatJSON,
			obj: &struct {
				ConfigFile string
				Name       string
			}{
				ConfigFile: "nonexistent.json",
			},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					ConfigFile string
					Name       string
				})
				if cfg.Name != "" {
					t.Errorf("expected Name to be empty, got '%s'", cfg.Name)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLoader(tt.opts...)
			l.RegisterFileField(tt.field, tt.format)
			err := l.Load(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.verify != nil {
				tt.verify(t, tt.obj)
			}
		})
	}
}

func Test_loader_RegisterFlags(t *testing.T) {
	tests := []struct {
		name    string
		opts    []loaderOption
		f       *flag.FlagSet
		obj     any
		wantErr bool
		verify  func(t *testing.T, obj any)
	}{
		{
			name: "load from flags",
			opts: []loaderOption{},
			f: func() *flag.FlagSet {
				fset := flag.NewFlagSet("test", flag.ContinueOnError)
				fset.String("name", "", "name flag")
				fset.String("database-host", "", "database host flag")
				fset.String("database-port", "", "database port flag")
				_ = fset.Parse([]string{"--name=MyApp", "--database-host=localhost", "--database-port=5432"})
				return fset
			}(),
			obj: &struct {
				Name     string
				Database struct {
					Host string
					Port int
				}
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct {
					Name     string
					Database struct {
						Host string
						Port int
					}
				})
				if cfg.Name != "MyApp" {
					t.Errorf("expected Name to be 'MyApp', got '%s'", cfg.Name)
				}
				if cfg.Database.Host != "localhost" {
					t.Errorf("expected Database.Host to be 'localhost', got '%s'", cfg.Database.Host)
				}
				if cfg.Database.Port != 5432 {
					t.Errorf("expected Database.Port to be 5432, got %d", cfg.Database.Port)
				}
			},
		},
		{
			name: "load bool field from flags",
			opts: []loaderOption{},
			f: func() *flag.FlagSet {
				fset := flag.NewFlagSet("test", flag.ContinueOnError)
				fset.Bool("debug", false, "debug flag")
				_ = fset.Parse([]string{"--debug"})
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
			name: "unparsed flags are ignored",
			opts: []loaderOption{},
			f: func() *flag.FlagSet {
				fset := flag.NewFlagSet("test", flag.ContinueOnError)
				fset.String("name", "", "name flag")
				// Don't parse flags
				return fset
			}(),
			obj: &struct {
				Name string
			}{},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				cfg := obj.(*struct{ Name string })
				if cfg.Name != "" {
					t.Errorf("expected Name to be empty, got '%s'", cfg.Name)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLoader(tt.opts...)
			l.RegisterFlags(tt.f)
			err := l.Load(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.verify != nil {
				tt.verify(t, tt.obj)
			}
		})
	}
}
