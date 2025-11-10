package confless

import (
	"flag"
	"fmt"
	"os"

	"github.com/codetent/confless/pkg/dotpath"
	"github.com/spf13/afero"
)

type fileFormat string

const (
	FileFormatJSON fileFormat = "json"
	FileFormatYAML fileFormat = "yaml"
)

type configFile struct {
	path   string
	format fileFormat
}

type configFileField struct {
	field  string
	format fileFormat
}

type loader struct {
	fs        afero.Fs
	envReader func() []string

	envPrefix  string
	flagSets   []*flag.FlagSet
	files      []*configFile
	fileFields []*configFileField
}

// Creates a new loader with the given options.
func NewLoader(opts ...loaderOption) *loader {
	l := &loader{
		fs:        afero.NewOsFs(),
		envReader: os.Environ,
		flagSets:  make([]*flag.FlagSet, 0),
		files:     make([]*configFile, 0),
	}

	// Apply the given options.
	for _, opt := range opts {
		opt(l)
	}

	return l
}

// Register an environment variable prefix to load.
// Names are converted to dot-separated paths (e.g. "MY_FLAG" -> "my.flag").
func (l *loader) RegisterEnv(pre string) {
	l.envPrefix = pre
}

// Register a file to load.
func (l *loader) RegisterFile(path string, format fileFormat) {
	l.files = append(l.files, &configFile{
		path:   path,
		format: format,
	})
}

// Register a field in the config that contains the path to a file to load.
func (l *loader) RegisterFileField(field string, format fileFormat) {
	l.fileFields = append(l.fileFields, &configFileField{
		field:  field,
		format: format,
	})
}

// Register the flags to load.
// Names are converted to dot-separated paths (e.g. "my-flag" -> "my.flag").
// Note that flags must be parsed before loading.
func (l *loader) RegisterFlags(f *flag.FlagSet) {
	l.flagSets = append(l.flagSets, f)
}

// Populate the object by applying the registered sources.
func (l *loader) Load(obj any) error {
	files := make([]*configFile, 0, len(l.fileFields)+len(l.files))
	files = append(files, l.files...)

	// Collect files from file fields.
	for _, field := range l.fileFields {
		// Get the value of the field.
		value, err := dotpath.Get(obj, field.field)
		if err != nil {
			return fmt.Errorf("failed to get field: %w", err)
		}

		path, ok := value.(string)
		if !ok {
			return fmt.Errorf("field is not a string: %s", field.field)
		}

		files = append(files, &configFile{
			path:   path,
			format: field.format,
		})
	}

	// Load the files.
	for _, file := range files {
		// Open the file.
		f, err := l.fs.Open(file.path)
		if err != nil {
			if os.IsNotExist(err) {
				// Skip if file does not exist.
				continue
			}

			return fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()

		// Populate the object by the file.
		err = populateByFile(f, string(file.format), obj)
		if err != nil {
			return fmt.Errorf("failed to load file: %w", err)
		}
	}

	// Load the flags.
	for _, fset := range l.flagSets {
		err := populateByFlags(fset, obj)
		if err != nil {
			return fmt.Errorf("failed to load flags: %w", err)
		}
	}

	// Load the environment variables.
	err := populateByEnv(l.envReader(), l.envPrefix, obj)
	if err != nil {
		return fmt.Errorf("failed to load env: %w", err)
	}

	return nil
}
