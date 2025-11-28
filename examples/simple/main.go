package main

import (
	"fmt"
	"log"

	"github.com/pigeonstorm/json5"
)

func main() {
	input := `{
		// This is a JSON5 object
		hello: 'world',
		count: 42,
	}`

	var result map[string]interface{}
	if err := json5.Unmarshal([]byte(input), &result); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Hello: %s\n", result["hello"])
	fmt.Printf("Count: %v\n", result["count"])

	// Marshal back to JSON
	output, err := json5.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Marshalled JSON:\n%s\n", output)
}
