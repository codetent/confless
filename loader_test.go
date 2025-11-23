package confless

import (
	"flag"
	"reflect"
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
		fileOpts []fileOption
		content string
		obj     any
		wantErr bool
		verify  func(t *testing.T, obj any)
	}{
		{
			name: "load from JSON file with automatic format detection",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "config.json", []byte(`{"name": "MyApp", "database": {"host": "localhost", "port": 5432}}`), 0644)
					return fs
				}()),
			},
			path:    "config.json",
			fileOpts: []fileOption{},
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
			name: "load from YAML file with automatic format detection",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "config.yaml", []byte("name: MyApp\nport: 8080"), 0644)
					return fs
				}()),
			},
			path:    "config.yaml",
			fileOpts: []fileOption{},
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
			name: "load from YML file with automatic format detection",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "config.yml", []byte("name: MyApp\nport: 8080"), 0644)
					return fs
				}()),
			},
			path:    "config.yml",
			fileOpts: []fileOption{},
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
			name: "override format with WithFileFormat option",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "config.txt", []byte("name: MyApp\nport: 8080"), 0644)
					return fs
				}()),
			},
			path:    "config.txt",
			fileOpts: []fileOption{WithFileFormat(FileFormatYAML)},
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
			path:    "config.json",
			fileOpts: []fileOption{},
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
			path:    "nonexistent.json",
			fileOpts: []fileOption{},
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
		{
			name: "file without extension defaults to JSON",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "config", []byte(`{"name": "MyApp", "port": 8080}`), 0644)
					return fs
				}()),
			},
			path:    "config",
			fileOpts: []fileOption{},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLoader(tt.opts...)
			l.RegisterFile(tt.path, tt.fileOpts...)
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

func Test_loader_ConfigTag(t *testing.T) {
	tests := []struct {
		name    string
		opts    []loaderOption
		obj     any
		wantErr bool
		verify  func(t *testing.T, obj any)
	}{
		{
			name: "load from file path in tagged field with automatic format detection",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "production.json", []byte(`{"name": "ProductionApp", "port": 9000}`), 0644)
					return fs
				}()),
			},
			obj: &struct {
				ConfigFile string `confless:"file"`
				Name       string
				Port       int
			}{
				ConfigFile: "production.json",
			},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				// Use reflection to get values to avoid type assertion issues with tags
				v := reflect.ValueOf(obj).Elem()
				name := v.FieldByName("Name").String()
				port := int(v.FieldByName("Port").Int())
				if name != "ProductionApp" {
					t.Errorf("expected Name to be 'ProductionApp', got '%s'", name)
				}
				if port != 9000 {
					t.Errorf("expected Port to be 9000, got %d", port)
				}
			},
		},
		{
			name: "load from YAML file path in tagged field with automatic format detection",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "production.yaml", []byte("name: ProductionApp\nport: 9000"), 0644)
					return fs
				}()),
			},
			obj: &struct {
				ConfigFile string `confless:"file"`
				Name       string
				Port       int
			}{
				ConfigFile: "production.yaml",
			},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				// Use reflection to get values to avoid type assertion issues with tags
				v := reflect.ValueOf(obj).Elem()
				name := v.FieldByName("Name").String()
				port := int(v.FieldByName("Port").Int())
				if name != "ProductionApp" {
					t.Errorf("expected Name to be 'ProductionApp', got '%s'", name)
				}
				if port != 9000 {
					t.Errorf("expected Port to be 9000, got %d", port)
				}
			},
		},
		{
			name: "load from tagged field with explicit format",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "config.yaml", []byte("name: YAMLApp\nport: 8080"), 0644)
					return fs
				}()),
			},
			obj: &struct {
				ConfigFile string `confless:"file,format=yaml"`
				Name       string
				Port       int
			}{
				ConfigFile: "config.yaml",
			},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				// Use reflection to get values to avoid type assertion issues with tags
				v := reflect.ValueOf(obj).Elem()
				name := v.FieldByName("Name").String()
				port := int(v.FieldByName("Port").Int())
				if name != "YAMLApp" {
					t.Errorf("expected Name to be 'YAMLApp', got '%s'", name)
				}
				if port != 8080 {
					t.Errorf("expected Port to be 8080, got %d", port)
				}
			},
		},
		{
			name: "load from nested tagged field",
			opts: []loaderOption{
				WithFS(func() afero.Fs {
					fs := afero.NewMemMapFs()
					_ = afero.WriteFile(fs, "nested.json", []byte(`{"name": "NestedApp", "port": 7000}`), 0644)
					return fs
				}()),
			},
			obj: &struct {
				Settings struct {
					ConfigPath string `confless:"file"`
				}
				Name string
				Port int
			}{
				Settings: struct {
					ConfigPath string `confless:"file"`
				}{
					ConfigPath: "nested.json",
				},
			},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				// Use reflection to get values to avoid type assertion issues with tags
				v := reflect.ValueOf(obj).Elem()
				name := v.FieldByName("Name").String()
				port := int(v.FieldByName("Port").Int())
				if name != "NestedApp" {
					t.Errorf("expected Name to be 'NestedApp', got '%s'", name)
				}
				if port != 7000 {
					t.Errorf("expected Port to be 7000, got %d", port)
				}
			},
		},
		{
			name: "skip empty tagged field",
			opts: []loaderOption{
				WithFS(afero.NewMemMapFs()),
			},
			obj: &struct {
				ConfigFile string `confless:"file"`
				Name       string
			}{
				ConfigFile: "",
			},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				// Use reflection to get values to avoid type assertion issues with tags
				v := reflect.ValueOf(obj).Elem()
				name := v.FieldByName("Name").String()
				if name != "" {
					t.Errorf("expected Name to be empty, got '%s'", name)
				}
			},
		},
		{
			name: "skip missing file from tagged field",
			opts: []loaderOption{
				WithFS(afero.NewMemMapFs()),
			},
			obj: &struct {
				ConfigFile string `confless:"file"`
				Name       string
			}{
				ConfigFile: "nonexistent.json",
			},
			wantErr: false,
			verify: func(t *testing.T, obj any) {
				// Use reflection to get values to avoid type assertion issues with tags
				v := reflect.ValueOf(obj).Elem()
				name := v.FieldByName("Name").String()
				if name != "" {
					t.Errorf("expected Name to be empty, got '%s'", name)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLoader(tt.opts...)
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
