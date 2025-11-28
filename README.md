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
go get github.com/victor/json5
```

## Usage

### Unmarshal

```go
package main

import (
	"fmt"
	"github.com/victor/json5"
)

func main() {
	input := `{
		// Comments are allowed
		key: 'value', // Unquoted keys and single quotes
	}`
	
	var result map[string]interface{}
	json5.Unmarshal([]byte(input), &result)
	fmt.Println(result)
}
```

### Decoder

```go
f, _ := os.Open("config.json5")
defer f.Close()

var cfg Config
json5.NewDecoder(f).Decode(&cfg)
```

## License

MIT
