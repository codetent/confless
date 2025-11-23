package confless

import "github.com/spf13/afero"

const (
	FileFormatJSON fileFormat = "json"
	FileFormatYAML fileFormat = "yaml"
)

type loaderOption func(l *loader)
type fileOption func(f *configFile)
type fileFormat string

// Set the file system to use.
func WithFS(fs afero.Fs) loaderOption {
	return func(l *loader) {
		l.fs = fs
	}
}

// Set the environment reader to use.
func WithEnvReader(reader func() []string) loaderOption {
	return func(l *loader) {
		l.envReader = reader
	}
}

// Set the file format to use.
func WithFileFormat(format fileFormat) fileOption {
	return func(f *configFile) {
		f.format = format
	}
}
