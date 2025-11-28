package main

import (
	"fmt"
	"log"

	"github.com/victor/json5"
)

func main() {
	data := `{
// comments
unquoted: 'and you can quote me on that',
singleQuotes: 'I can use "double quotes" here',
lineBreaks: "Look, Mom! \
No \\n's!",
hexadecimal: 0xdecaf,
leadingDecimalPoint: .8675309, andTrailing: 8675309.,
positiveSign: +1,
trailingComma: 'in objects', andIn: ['arrays',],
"backwardsCompatible": "with JSON",
}

`

	var result map[string]interface{}
	if err := json5.Unmarshal([]byte(data), &result); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully parsed example.json5:\n")
	for k, v := range result {
		fmt.Printf("- %s: %v\n", k, v)
	}
}
