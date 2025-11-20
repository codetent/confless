package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/codetent/confless"
)

type demo struct {
	Name    string
	Config  string `json:"config" confless:"file"`
	Age     int    `json:"age"`
	Objects struct {
		Apple string
	}
	Items     []int
	CreatedAt time.Time
}

func init() {
	flag.String("name", "", "the name of the object")
}

func main() {
	obj := &demo{
		Name:   "Alice",
		Items:  []int{0, 0},
		Config: "other.json",
	}

	confless.RegisterEnv("example")
	confless.RegisterFile("config.json", confless.FileFormatJSON)
	confless.RegisterFlags(flag.CommandLine)

	flag.Parse()

	err := confless.Load(obj)
	if err != nil {
		fmt.Println("failed to load:", err)
	}

	fmt.Println(obj)
}
