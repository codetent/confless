<br>
<div align="center">
  <img height="50" src="doc/confless.svg">
  <p align="center">
    <b>configuration with less effort</b>
  </p>
</div>
<br>

![GitHub Check](https://github.com/codetent/confless/workflows/Check/badge.svg)
![GitHub License](https://img.shields.io/github/license/codetent/confless)

A simple, flexible Go library for loading configuration from multiple sources with minimal boilerplate.

- **Less Boilerplate**: No need to manually parse flags, read files, and merge configurations
- **Less Complexity**: Simple API that does one thing well
- **Less Configuration**: Reasonable defaults with optional customization
- **More Flexibility**: Support for multiple sources with automatic precedence

## 📦 Installation

```bash
go get github.com/codetent/confless
```

## 🚀 Quick Start

```go
type Config struct {
	  Port int
}

config := &Config{Port: 8080}

confless.RegisterFile("config.json")
confless.RegisterEnv("APP")
confless.RegisterFlags(flag.CommandLine)

flag.Parse()

err := confless.Load(config)
if err != nil {
    log.Fatal(err)
}
```

## 📚 Basics

### Keys

Field names are taken from struct fields.
Tag annotations like `json` and `yaml` can be used to override the field name.

### Values

Types are taken from struct fields.

The following basic types can be set from all sources:
- string
- bool
- int (and all variants: int8, int16, int32, int64)
- uint (and all variants: uint8, uint16, uint32, uint64)
- float32, float64

Complex types like slices and maps can only be set directly in the struct or by loading values from files.

Default values for fields can be set when initializing the struct.
They will be overridden by values from sources if set.

## 📁 Sources

Sources are applied in the following order (later sources override earlier ones):

1. **Files** (in registration order)
2. **Command-line flags**
3. **Environment variables** (highest precedence)
4. **Dynamically registered files**

### Files

Load configuration from JSON or YAML files. Files are loaded in registration order and merged together. Missing files are silently skipped.

The file format is automatically detected from the file extension.
If the extension is not supported, the file format defaults to JSON.

You can also explicitly specify the format using file options:

```go
// Register a JSON file (format detected automatically from .json extension)
confless.RegisterFile("config.json")

// Register a YAML file (format detected automatically from .yaml extension)
confless.RegisterFile("config.yaml")

// Register a file with explicit format override
confless.RegisterFile("config.txt", confless.WithFileFormat(confless.FileFormatYAML))
```

**Example `config.json`:**
```json
{
    "name": "MyApp",
    "port": 3000,
    "database": {
        "host": "localhost",
        "port": 5432
    }
}
```

**Example `config.yaml`:**
```yaml
name: MyApp
port: 3000
database:
  host: localhost
  port: 5432
```

#### Dynamic File Paths

You can mark a field in your configuration with the `confless:"file"` tag to automatically load it as a configuration file. This is useful for environment-specific configurations.

The format is automatically detected from the file extension (same rules as static file registration). You can also specify the format explicitly in the tag (`confless:"file,format=yaml"`) to override the automatic detection.

Note that dynamically registered files are loaded at the end, while statically registered files are loaded first.

```go
type Config struct {
    ConfigFile string `json:"config_file" confless:"file"` // Path to additional config
}

config := &Config{ConfigFile: "production.json"}

// The field is automatically detected via the confless:"file" tag
confless.Load(config)
```

### Environment Variables

Load configuration from environment variables with a specified prefix.

```go
// Register environment variables with prefix "APP"
confless.RegisterEnv("APP")
```

**Key Naming Convention:**
- Environment variables use underscores: `APP_DATABASE_HOST`
- They start with the specified prefix: `APP_`
- Array items are represented by their index: `APP_ITEMS_0`
- They are converted to dot notation to represent a path to fields: `database.host`

**Example:**
```go
// APP_NAME=Production
// APP_DATABASE_HOST=db.example.com
// APP_DATABASE_PORT=5432

confless.RegisterEnv("APP")
confless.Load(&config)
```

### Command-Line Flags

Load configuration from Go's standard `flag` package.
Matching flags, that have been defined beforehand, are automatically detected and if set, their values will be used to populate the struct.

```go
flag.String("name", "", "Application name")
flag.String("database-host", "", "Database host")

confless.RegisterFlags(flag.CommandLine)
flag.Parse()
confless.Load(&config)
```

**Key Naming Convention:**
- Flags use dashes: `--database-host`
- Array items are represented by their index: `--items-0`
- They are converted to dot notation to represent a path to fields: `database.host`

**Example:**
```bash
./app --name=MyApp --database-host=localhost
```

## 📝 Example

```go
type Config struct {
    Name string
    Port int
    Config string `confless:"file"`
}

// Register flags to load
flag.String("name", "", "the name of the object")

// Set default values for fields
config := &Config{
    Name: "DefaultApp",
    Port: 8080,
    Config: "production.json",
}

// Register sources to load
confless.RegisterFile("config.json")
// Config field is automatically detected via confless:"file" tag
confless.RegisterEnv("APP")
confless.RegisterFlags(flag.CommandLine)

// Parse flags before loading
flag.Parse()

// Load configuration
err := confless.Load(config)
if err != nil {
    log.Fatal(err)
}
```

**Usage:**
```bash
APP_PORT=9000 ./app --name=MyApp

# Sets port to 9000 instead of 8080 (default)
# Sets name to "MyApp" instead of "DefaultApp" (default)
# All other fields are loaded from the config.json file (if exists) or the default values are taken
```

For more examples, see the [examples](examples) directory.

## 💡 Tips & Tricks

### Validation

Since confless just populates a struct, you can use any validation library to validate it after loading.

One of the most popular is [validator](https://github.com/go-playground/validator).
Just add the `validate` tag to the struct fields you want to validate and validate it using the library.

```go
validate := validator.New(validator.WithRequiredStructEnabled())
err = validate.Struct(config)
if err != nil {
    log.Fatal(err)
}
```

### Multiple Loaders

If you need to load multiple configurations differently in one application, you can create multiple loaders instead of using the default global loader.

```go
loader := confless.NewLoader()
loader.RegisterEnv("APP")
```
