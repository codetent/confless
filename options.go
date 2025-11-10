package confless

import "github.com/spf13/afero"

type loaderOption func(l *loader)

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
