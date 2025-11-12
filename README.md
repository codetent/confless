<br>
<div align="center">
  <img height="50" src="doc/confless.svg">
  <p align="center">
    <b>configuration with less effort</b>
  </p>
</div>
<br>

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
config := &Config{Port: 8080}

confless.RegisterFile("config.json", confless.FileFormatJSON)
confless.RegisterEnv("APP")
confless.RegisterFlags(flag.CommandLine)

flag.Parse()
confless.Load(config)
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
- int
- uint
- float

Complex types like slices and maps can only be set directly in the struct or by loading values from files.

Default values for fields can be set when initializing the struct.
They will be overridden by values from sources if set.

## 📁 Sources

Sources are applied in the following order (later sources override earlier ones):

1. **Files** (in registration order)
2. **Command-line flags**
3. **Environment variables** (highest precedence)

### Files

Load configuration from JSON or YAML files. Files are loaded in registration order and merged together. Missing files are silently skipped.

```go
// Register a JSON file
confless.RegisterFile("config.json", confless.FileFormatJSON)

// Register a YAML file
confless.RegisterFile("config.yaml", confless.FileFormatYAML)
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

You can also register a field in your configuration that contains the path to another configuration file. This is useful for environment-specific configurations.

The file field must be a dot-separated path to the field in the struct.

```go
type Config struct {
    ConfigFile string `json:"config_file"` // Path to additional config
}

config := &Config{ConfigFile: "production.json"}

confless.RegisterFileField("config_file", confless.FileFormatJSON)
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
Matching flags are automatically detected and if set, their values will be used to populate the struct.

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

## 💡 Example

```go
// Register flags to load
flag.String("name", "", "the name of the object")

// Set default values for fields
config := &Config{
    Name: "DefaultApp",
    Port: 8080,
    Config: "production.json",
}

// Register sources to load
confless.RegisterFile("config.json", confless.FileFormatJSON)
confless.RegisterFileField("config", confless.FileFormatJSON)
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
