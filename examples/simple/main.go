package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/codetent/confless"
)

type Config struct {
	Name  string `json:"name"`
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Debug bool   `json:"debug"`

	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"database"`

	Config string `confless:"file"`
}

func main() {
	// Define command-line flags
	flag.String("host", "", "Server host")
	flag.Int("port", 0, "Server port")
	flag.Bool("debug", false, "Enable debug mode")
	flag.String("database-host", "", "Database host")
	flag.Int("database-port", 0, "Database port")
	flag.Bool("database-ssl", false, "Enable SSL for database")
	flag.String("config", "", "Path to configuration file")

	// Set default values
	config := &Config{
		Name:  "default",
		Host:  "localhost",
		Port:  8080,
		Debug: false,
	}

	// Read configuration from command-line flags
	confless.RegisterFlags(flag.CommandLine)
	// Read configuration from environment variables starting with "APP_"
	confless.RegisterEnv("APP")
	// Read configuration from config.json
	confless.RegisterFile("files/config.json")

	// Parse flags before loading
	flag.Parse()

	// Load configuration from all registered sources
	err := confless.Load(config)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Print the final configuration to stdout
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(config); err != nil {
		log.Fatalf("Failed to encode config as JSON: %v", err)
	}
}
