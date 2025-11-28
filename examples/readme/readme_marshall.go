package main

import (
	"fmt"
	"log"

	"github.com/pigeonstorm/json5"
)

func main() {
	data := map[string]interface{}{
		// comments
		"unquoted":     "and you can quote me on that",
		"singleQuotes": "I can use \"double quotes\" here",
		"lineBreaks": `Look, Mom! \
		No \n's!`,
		"hexadecimal":         0xdecaf,
		"leadingDecimalPoint": .8675309,
		"andTrailing":         8675309.,
		"positiveSign":        +1,
		"trailingComma":       "in objects",
		"andIn":               []string{"arrays"},
		"backwardsCompatible": "with JSON",
	}
	// Marshal back to JSON to show it works
	output, err := json5.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nMarshalled JSON:\n%s\n", output)
}
