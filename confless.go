package confless

import "flag"

var (
	defaultLoader = NewLoader()
)

// Register an environment variable prefix to load.
// Names are converted to dot-separated paths (e.g. "MY_FLAG" -> "my.flag").
func RegisterEnv(pre string) {
	defaultLoader.RegisterEnv(pre)
}

// Register a file to load.
func RegisterFile(path string, format fileFormat) {
	defaultLoader.RegisterFile(path, format)
}

// Register a field in the config that contains the path to a file to load.
func RegisterFileField(field string, format fileFormat) {
	defaultLoader.RegisterFileField(field, format)
}

// Register the flags to load.
// Names are converted to dot-separated paths (e.g. "my-flag" -> "my.flag").
// Note that flags must be parsed before loading.
func RegisterFlags(f *flag.FlagSet) {
	defaultLoader.RegisterFlags(f)
}

// Populate the given object by applying the registered sources.
func Load(obj any) error {
	return defaultLoader.Load(obj)
}
