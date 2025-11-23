package confless

import "flag"

var (
	defaultLoader = NewLoader()
)

// Configure the default loader with the given options.
// Note that reconfiguring the default loader will reset the registered sources.
func Configure(opts ...loaderOption) {
	defaultLoader = NewLoader(opts...)
}

// Register an environment variable prefix to load.
// Names are converted to dot-separated paths (e.g. "MY_FLAG" -> "my.flag").
func RegisterEnv(pre string) {
	defaultLoader.RegisterEnv(pre)
}

// Register a file to load.
func RegisterFile(path string, opts ...fileOption) {
	defaultLoader.RegisterFile(path, opts...)
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
