package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/pigeonstorm/json5"
)

// Load config from JSON5 file

// Configuration object
type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func main() {
	// Simulating a config file
	configData := `
	{
		host: "localhost",
		port: 8080, // Default port
	}
	`

	dec := json5.NewDecoder(strings.NewReader(configData))
	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Config: %+v\n", cfg)

	// Marshal back to JSON
	output, err := json5.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Marshalled JSON:\n%s\n", output)
}
