package json5

import (
	"bytes"
	"encoding/json"
	"unicode/utf8"
)

// Marshal returns the JSON5 encoding of v.
// Keys that are valid identifiers will be unquoted.
func Marshal(v interface{}) ([]byte, error) {
	// First marshal to JSON
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	
	// Convert JSON to JSON5 format (unquote valid identifier keys)
	return convertJSONToJSON5(jsonBytes)
}

// MarshalIndent is like Marshal but applies Indent to format the output.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	// First marshal to JSON with indentation
	jsonBytes, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, err
	}
	
	// Convert JSON to JSON5 format (unquote valid identifier keys)
	return convertJSONToJSON5(jsonBytes)
}

// convertJSONToJSON5 converts JSON bytes to JSON5 format by unquoting valid identifier keys
func convertJSONToJSON5(jsonBytes []byte) ([]byte, error) {
	var buf bytes.Buffer
	pos := 0
	
	for pos < len(jsonBytes) {
		c := jsonBytes[pos]
		
		// Look for quoted strings that might be keys
		if c == '"' {
			start := pos
			pos++
			escaped := false
			foundClosing := false
			
			// Find the end of the string
			for pos < len(jsonBytes) && !foundClosing {
				if escaped {
					escaped = false
					pos++
					continue
				}
				if jsonBytes[pos] == '\\' {
					escaped = true
					pos++
					continue
				}
				if jsonBytes[pos] == '"' {
					// Found closing quote
					keyEnd := pos
					pos++ // consume closing quote
					foundClosing = true
					
					// Check if this is followed by ':' (indicating it's a key)
					colonPos := pos
					// Skip whitespace
					for colonPos < len(jsonBytes) {
						ws := jsonBytes[colonPos]
						if ws != ' ' && ws != '\t' && ws != '\n' && ws != '\r' {
							break
						}
						colonPos++
					}
					
					// If followed by ':', it's a key
					if colonPos < len(jsonBytes) && jsonBytes[colonPos] == ':' {
						keyBytes := jsonBytes[start+1 : keyEnd]
						
						// Check if it's a valid identifier (can be unquoted)
						if isValidIdentifier(keyBytes) {
							// Write unquoted key
							buf.Write(keyBytes)
							// Write whitespace and colon
							buf.Write(jsonBytes[pos:colonPos+1])
							pos = colonPos + 1
							// Break out to continue outer loop
							break
						}
					}
					
					// Not a key or not valid identifier, write as-is
					buf.Write(jsonBytes[start:pos])
					break
				}
				pos++
			}
			
			if !foundClosing {
				// Shouldn't reach here with valid JSON, but handle it
				buf.WriteByte(c)
				pos++
			}
			continue
		}
		
		// Regular character, copy as-is
		buf.WriteByte(c)
		pos++
	}
	
	return buf.Bytes(), nil
}

// isValidIdentifier checks if a byte slice represents a valid JSON5 identifier
func isValidIdentifier(key []byte) bool {
	if len(key) == 0 {
		return false
	}
	
	// Check first character (must be identifier start)
	r, width := utf8.DecodeRune(key)
	if !isIdentifierStart(r) {
		return false
	}
	
	pos := width
	for pos < len(key) {
		r, width := utf8.DecodeRune(key[pos:])
		if !isIdentifierPart(r) {
			return false
		}
		pos += width
	}
	
	return true
}


