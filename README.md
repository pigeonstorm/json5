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

## Performance Benchmark

Benchmarking results comparing native Go `encoding/json` with this JSON5 library on **~15 MB** of synthetic data (nested objects, arrays, and various data types).

### Marshal Performance

| Library | ns/op | B/op | allocs/op | Overhead |
|---|---|---|---|---|
| `encoding/json` | 80.3M | 77.3 MB | 1.81M | baseline |
| `json5` | 80.1M | 77.3 MB | 1.81M | -0.28% |

**Note:** `json5.Marshal` delegates to `encoding/json.Marshal`, so performance is identical.

### Unmarshal Performance

| Library | ns/op | B/op | allocs/op | Overhead |
|---|---|---|---|---|
| `encoding/json` | 90.9M | 85.1 MB | 2.65M | baseline |
| `json5` | 138.4M | 103.5 MB | 2.71M | +52.3% |

**Note:** `json5.Unmarshal` performs JSON5-to-JSON transcoding before unmarshaling, which adds overhead. The transcoding step has been optimized with fast-path ASCII handling, reducing overhead from ~79.5% to ~52.3%.

### Data Usage Comparison

| Format | Size | Overhead |
|---|---|---|
| Standard JSON | 15.00 MB (15,730,018 bytes) | baseline |
| JSON5 | 15.43 MB (16,180,306 bytes) | +2.86% |

**Note:** The JSON5 format in this benchmark uses unquoted keys and single quotes, which can be more compact in some cases, but the formatting adds some overhead. Actual size savings depend on the specific JSON5 features used (unquoted keys, trailing commas, etc.).

### Running Benchmarks

To run the comprehensive benchmarks yourself:

```bash
go test -bench="Benchmark.*Comprehensive" -benchmem
```

## License

MIT