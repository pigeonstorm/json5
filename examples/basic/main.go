package main

import (
	"fmt"
	"log"
	"os"

	"github.com/victor/json5"
)

func main() {
	// Read the JSON5 file
	data, err := os.ReadFile("example.json5")
	if err != nil {
		log.Fatal(err)
	}

	var result map[string]interface{}
	if err := json5.Unmarshal(data, &result); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully parsed example.json5:\n")
	for k, v := range result {
		fmt.Printf("- %s: %v\n", k, v)
	}

	// Marshal back to JSON to show it works
	output, err := json5.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nMarshalled JSON:\n%s\n", output)
}
