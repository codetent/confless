package confless

import (
	"flag"
	"fmt"
	"os"

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

type loader struct {
	fs        afero.Fs
	envReader func() []string

	envPrefix string
	flagSets  []*flag.FlagSet
	files     []*configFile
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

// Register the flags to load.
// Names are converted to dot-separated paths (e.g. "my-flag" -> "my.flag").
// Note that flags must be parsed before loading.
func (l *loader) RegisterFlags(f *flag.FlagSet) {
	l.flagSets = append(l.flagSets, f)
}

// Populate the object by applying the registered sources.
func (l *loader) Load(obj any) error {
	files := make([]*configFile, 0, len(l.files))
	files = append(files, l.files...)

	// Add file paths from tagged fields.
	for field, format := range findFileFields(obj) {
		path := field.String()
		if path == "" {
			continue
		}

		format := fileFormat(format)
		if format == "" {
			format = FileFormatJSON
		}

		files = append(files, &configFile{
			path:   path,
			format: format,
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

		// Populate the object by the file.
		err = populateByFile(f, string(file.format), obj)
		_ = f.Close()
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
	if l.envPrefix != "" {
		err := populateByEnv(l.envReader(), l.envPrefix, obj)
		if err != nil {
			return fmt.Errorf("failed to load env: %w", err)
		}
	}

	return nil
}
