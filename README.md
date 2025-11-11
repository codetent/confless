# confless

**configuration with less effort** - A simple, flexible Go library for loading configuration from multiple sources with minimal boilerplate.

## Features

- **Less Boilerplate**: No need to manually parse flags, read files, and merge configurations
- **Less Complexity**: Simple API that does one thing well
- **Less Configuration**: Reasonable defaults with optional customization
- **More Flexibility**: Support for multiple sources with automatic precedence

## Installation

```bash
go get github.com/codetent/confless
```

## Quick Start

```go
config := &Config{Port: 8080}

confless.RegisterFile("config.json", confless.FileFormatJSON)
confless.RegisterEnv("APP")
confless.RegisterFlags(flag.CommandLine)

flag.Parse()
confless.Load(config)
```

## Configuration Sources

### Files (JSON/YAML)

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

The keys either have to match the field names in the struct or be annotated with tags.

#### Dynamic File Paths

You can also register a field in your configuration that contains the path to another configuration file. This is useful for environment-specific configurations.

The file field must be a dot-separated path to the field in the struct.
Either the key has to match the actual field name or be annotated with a tag.

```go
type Config struct {
    ConfigFile string `json:"config_file"` // Path to additional config
}

config := &Config{ConfigFile: "production.json"}

confless.RegisterFileField("config_file", confless.FileFormatJSON)
confless.Load(config)
```

### Environment Variables

Load configuration from environment variables with a specified prefix. Environment variable names are converted to dot-separated paths.

```go
// Register environment variables with prefix "APP"
confless.RegisterEnv("APP")
```

**Naming Convention:**
- Environment variables use underscores: `APP_DATABASE_HOST`
- They are converted to dot notation to represent a path: `database.host`
- Arrays are represented as a dot-separated list of indices: `items.0`
- Tag annotations can be used to override the key
- The prefix is removed: `APP_` → removed

**Example:**
```go
// APP_NAME=Production
// APP_DATABASE_HOST=db.example.com
// APP_DATABASE_PORT=5432

confless.RegisterEnv("APP")
confless.Load(&config)
```

### Command-Line Flags

Load configuration from Go's standard `flag` package. Flag names are converted to dot-separated paths.

```go
flag.String("name", "", "Application name")
flag.String("database-host", "", "Database host")

confless.RegisterFlags(flag.CommandLine)
flag.Parse()
confless.Load(&config)
```

**Naming Convention:**
- Flags use dashes: `--database-host`
- They are converted to dot notation: `database.host`
- Tag annotations can be used to override the key

**Example:**
```bash
./app --name=MyApp --database-host=localhost
```

## Precedence

Configuration sources are applied in the following order (later sources override earlier ones):

1. **Files** (in registration order)
2. **Command-line flags**
3. **Environment variables** (highest precedence)

This means environment variables will always override file and flag values, making them perfect for deployment-specific overrides.

## Example

```go
flag.String("name", "", "the name of the object")

config := &Config{
    Name: "DefaultApp",
    Port: 8080,
    Config: "production.json",
}

confless.RegisterFile("config.json", confless.FileFormatJSON)
confless.RegisterFileField("config", confless.FileFormatJSON)
confless.RegisterEnv("APP")
confless.RegisterFlags(flag.CommandLine)

flag.Parse()
confless.Load(config)
```

**Usage:**
```bash
APP_PORT=9000 ./app --name=MyApp
```
