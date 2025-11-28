# JSON5 for Go

A Go package that implements the [JSON5](https://spec.json5.org/) data interchange format.
This package wraps the standard `encoding/json` package, preserving existing `Marshal` and `Unmarshal` behavior while adding support for JSON5 features.

## Features

- **Comments**: Single (`//`) and multi-line (`/* ... */`) comments.
- **Trailing Commas**: Allowed in objects and arrays.
- **Single Quotes**: Strings can be single-quoted.
- **Unquoted Keys**: Object keys can be unquoted if they are valid identifiers.
- **Numbers**: Hexadecimal (`0x...`), leading/trailing decimals, explicit plus sign.
- **Line Continuations**: Escaped newlines in strings.

## Installation

```bash
go get github.com/jvmvik/json5
```

## Usage

### Unmarshal
 Read a JSON5 file and unmarshal it into a map.	

```go
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
```

### Mashal

Write a map to a JSON5 file.	
```go
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
```

## TODO
 - [ ] Improve benchmarking

## Performance Benchmark
Basic benchmarking results are shown below. 

### Marshal

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| BenchmarkMarshal_JSON | 383.0 | 416 | 7 |
| BenchmarkMarshal_JSON5 | 386.1 | 416 | 7 |

### Unmarshal (standard JSON input)

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| BenchmarkUnmarshal_JSON | 1168 | 416 | 19 |
| BenchmarkUnmarshal_JSON5_StandardInput | 1927 | 608 | 21 |

### Unmarshal (JSON5 input)

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| BenchmarkUnmarshal_JSON5_JSON5Input | 2103 | 640 | 22 |

## Data Usage Comparison

| Format | Size (bytes) |
|---|---|
| Standard JSON | 185 |
| Compact JSON5 | 169 |

**Savings:** 16 bytes (8.65%)

## License

MIT