package json5

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// Transcode converts JSON5 input to standard JSON.
func Transcode(data []byte) ([]byte, error) {
	t := &transcoder{
		data: data,
		pos:  0,
	}
	// Pre-allocate buffer
	t.out.Grow(len(data))

	if err := t.scanValue(); err != nil {
		return nil, err
	}

	// Check for trailing non-whitespace garbage
	t.skipWhitespace()
	if t.pos < len(t.data) {
		return nil, fmt.Errorf("unexpected data after top-level value at position %d", t.pos)
	}

	return t.out.Bytes(), nil
}

type transcoder struct {
	data []byte
	pos  int
	out  bytes.Buffer
}

func (t *transcoder) scanValue() error {
	t.skipWhitespace()
	if t.pos >= len(t.data) {
		return fmt.Errorf("unexpected EOF")
	}

	r, _ := utf8.DecodeRune(t.data[t.pos:])
	switch r {
	case '{':
		return t.scanObject()
	case '[':
		return t.scanArray()
	case '"', '\'':
		return t.scanString()
	default:
		if isNumberStart(r) {
			return t.scanNumber()
		}
		// Try literal (true, false, null, Infinity, NaN)
		return t.scanLiteral()
	}
}

func (t *transcoder) scanObject() error {
	t.out.WriteByte('{')
	t.pos++ // consume '{'

	first := true
	for {
		t.skipWhitespace()
		if t.pos >= len(t.data) {
			return fmt.Errorf("unexpected EOF in object")
		}
		if t.data[t.pos] == '}' {
			t.pos++
			t.out.WriteByte('}')
			return nil
		}

		if !first {
			if t.data[t.pos] != ',' {
				return fmt.Errorf("expected comma in object, got %c", t.data[t.pos])
			}
			t.pos++ // consume comma

			// Check for trailing comma
			t.skipWhitespace()
			if t.pos >= len(t.data) {
				return fmt.Errorf("unexpected EOF in object")
			}
			if t.data[t.pos] == '}' {
				t.pos++
				t.out.WriteByte('}')
				return nil
			}

			t.out.WriteByte(',')
		}

		// Scan Key
		if err := t.scanKey(); err != nil {
			return err
		}

		t.skipWhitespace()
		if t.pos >= len(t.data) || t.data[t.pos] != ':' {
			return fmt.Errorf("expected colon after key")
		}
		t.out.WriteByte(':')
		t.pos++

		if err := t.scanValue(); err != nil {
			return err
		}

		first = false
	}
}

func (t *transcoder) scanArray() error {
	t.out.WriteByte('[')
	t.pos++ // consume '['

	first := true
	for {
		t.skipWhitespace()
		if t.pos >= len(t.data) {
			return fmt.Errorf("unexpected EOF in array")
		}
		if t.data[t.pos] == ']' {
			t.pos++
			t.out.WriteByte(']')
			return nil
		}

		if !first {
			if t.data[t.pos] != ',' {
				return fmt.Errorf("expected comma in array")
			}
			t.pos++ // consume comma

			// Check for trailing comma
			t.skipWhitespace()
			if t.pos >= len(t.data) {
				return fmt.Errorf("unexpected EOF in array")
			}
			if t.data[t.pos] == ']' {
				t.pos++
				t.out.WriteByte(']')
				return nil
			}

			t.out.WriteByte(',')
		}

		if err := t.scanValue(); err != nil {
			return err
		}

		first = false
	}
}

func (t *transcoder) scanKey() error {
	t.skipWhitespace()
	if t.pos >= len(t.data) {
		return fmt.Errorf("unexpected EOF expecting key")
	}

	r, width := utf8.DecodeRune(t.data[t.pos:])
	if r == '"' || r == '\'' {
		return t.scanString()
	}

	// Unquoted key
	if isIdentifierStart(r) {
		start := t.pos
		t.pos += width
		for t.pos < len(t.data) {
			r, width = utf8.DecodeRune(t.data[t.pos:])
			if !isIdentifierPart(r) {
				break
			}
			t.pos += width
		}
		// Quote the key
		t.out.WriteByte('"')
		t.out.Write(t.data[start:t.pos])
		t.out.WriteByte('"')
		return nil
	}

	return fmt.Errorf("invalid key start: %c", r)
}

func (t *transcoder) scanString() error {
	quote := t.data[t.pos] // ' or "
	t.pos++
	t.out.WriteByte('"') // Always output double quote

	for t.pos < len(t.data) {
		r, width := utf8.DecodeRune(t.data[t.pos:])
		if r == rune(quote) {
			t.pos++
			t.out.WriteByte('"')
			return nil
		}

		if r == '\\' {
			if t.pos+1 >= len(t.data) {
				return fmt.Errorf("unexpected EOF in string escape")
			}
			// Handle escapes
			// JSON5 allows escaped newlines (continuation)
			next := t.data[t.pos+1]
			if next == '\n' || next == '\r' {
				// Line continuation, skip backslash and newline
				t.pos++ // skip backslash
				if next == '\r' && t.pos+1 < len(t.data) && t.data[t.pos+1] == '\n' {
					t.pos += 2
				} else {
					t.pos++
				}
				continue
			} else if next == '\'' && quote == '"' {
				// Escaped single quote in double quoted string?
				// JSON5: \ is escape. \' is valid.
				// JSON: \' is INVALID.
				// So if we have \', we must output ' without backslash if we are outputting double quotes.
				t.pos += 2
				t.out.WriteByte('\'')
				continue
			} else if next == '"' && quote == '\'' {
				// Escaped double quote in single quoted string?
				// JSON5: \" is valid.
				// JSON: \" is valid.
				// But wait, if input is 'foo\"bar', output is "foo\"bar". Correct.
				// If input is 'foo"bar', output is "foo\"bar". We need to escape the double quote!
				// Handled below in default case? No.
			}

			// If we are converting ' to ", we need to ensure " inside is escaped.
			// And ' inside is NOT escaped.

			// Let's handle generic char copy, but check for specific issues.
			// Actually, it's easier to just decode the escape and re-encode if needed?
			// No, that's slow.

			// Simple approach:
			// Copy backslash and next char, UNLESS:
			// 1. It's \' and we are outputting ". Then just output '.
			// 2. It's line continuation. Skip.

			if next == '\'' {
				t.pos += 2
				t.out.WriteByte('\'')
				continue
			}

			t.out.WriteByte('\\')
			t.out.WriteByte(next)
			t.pos += 2
			continue
		}

		if r == '"' && quote == '\'' {
			// Unescaped " inside ' string. Must escape it for JSON output.
			t.out.WriteByte('\\')
			t.out.WriteByte('"')
			t.pos++
			continue
		}

		// Regular char
		// JSON strings cannot contain unescaped control characters.
		// JSON5 allows some? No, mostly same.
		// Just copy.
		t.out.WriteRune(r)
		t.pos += width
	}
	return fmt.Errorf("unexpected EOF in string")
}

func (t *transcoder) scanNumber() error {
	// JSON5 numbers:
	// - Hex: 0x...
	// - Leading decimal: .5
	// - Trailing decimal: 5.
	// - Explicit plus: +5
	// - Infinity, NaN (handled in scanLiteral or here?)
	//   Actually, Infinity/NaN are identifiers usually, but -Infinity is - then identifier.

	start := t.pos

	// Check for hex
	if t.pos+2 <= len(t.data) && t.data[t.pos] == '0' && (t.data[t.pos+1] == 'x' || t.data[t.pos+1] == 'X') {
		t.pos += 2
		for t.pos < len(t.data) {
			r, width := utf8.DecodeRune(t.data[t.pos:])
			if !isHexDigit(r) {
				break
			}
			t.pos += width
		}
		hexStr := string(t.data[start:t.pos])
		// Parse hex
		n, err := strconv.ParseInt(hexStr, 0, 64)
		if err != nil {
			// Try uint
			un, err2 := strconv.ParseUint(hexStr, 0, 64)
			if err2 != nil {
				return fmt.Errorf("invalid hex number: %s", hexStr)
			}
			t.out.WriteString(strconv.FormatUint(un, 10))
			return nil
		}
		t.out.WriteString(strconv.FormatInt(n, 10))
		return nil
	}

	// Scan standard number chars
	// We need to handle + and leading .

	// Just consume all number-like characters and then try to parse/fix?
	// Or parse strictly?
	// Strict is better.

	if t.data[t.pos] == '+' {
		t.pos++ // Skip +
	}

	// Capture the rest
	numStart := t.pos
	hasDecimal := false

	if t.pos < len(t.data) && t.data[t.pos] == '.' {
		hasDecimal = true
		t.pos++
		// Must have digits after . for JSON?
		// JSON5: .5 is valid. JSON: 0.5
		// We will prepend 0 if needed.
	}

	for t.pos < len(t.data) {
		r, width := utf8.DecodeRune(t.data[t.pos:])
		if unicode.IsDigit(r) {
			t.pos += width
		} else if r == '.' {
			if hasDecimal {
				break // Second dot?
			}
			hasDecimal = true
			t.pos += width
		} else if r == 'e' || r == 'E' {
			t.pos += width
			if t.pos < len(t.data) && (t.data[t.pos] == '+' || t.data[t.pos] == '-') {
				t.pos++
			}
			// Expect digits
			for t.pos < len(t.data) {
				r2, w2 := utf8.DecodeRune(t.data[t.pos:])
				if !unicode.IsDigit(r2) {
					break
				}
				t.pos += w2
			}
			break // End of number
		} else {
			break
		}
	}

	numStr := string(t.data[numStart:t.pos])

	// Fixups
	if numStr == "" {
		// Could happen if we had just "+" or "."?
		if hasDecimal {
			// It was just "."?
			// JSON5 ". " is not valid number. ".5" is.
			// If we had "." but no digits, it's invalid.
			// But wait, we consumed digits.
			// If numStr is ".", that means we had "." and no digits after.
			// JSON5 requires digits after leading dot? No, `5.` is valid. `.5` is valid. `.` is NOT valid.
			if len(numStr) == 1 && numStr[0] == '.' {
				return fmt.Errorf("invalid number: dot")
			}
		} else {
			return fmt.Errorf("invalid number")
		}
	}

	// If it starts with ., prepend 0
	if numStr[0] == '.' {
		t.out.WriteByte('0')
	}
	t.out.WriteString(numStr)

	// If it ends with ., append 0?
	// JSON requires fractional part to have digits. "5." is invalid. "5.0" is valid.
	if numStr[len(numStr)-1] == '.' {
		t.out.WriteByte('0')
	}

	return nil
}

func (t *transcoder) scanLiteral() error {
	// true, false, null, Infinity, NaN
	// -Infinity is handled by scanNumber? No, - is number start.
	// Wait, scanNumber handles -?
	// isNumberStart includes -.
	// But -Infinity starts with -.
	// So scanNumber needs to handle Infinity if it sees -I...

	// Actually, let's just peek.
	start := t.pos
	for t.pos < len(t.data) {
		r, width := utf8.DecodeRune(t.data[t.pos:])
		if !isIdentifierPart(r) && r != '-' && r != '+' {
			break
		}
		t.pos += width
	}
	lit := string(t.data[start:t.pos])

	switch lit {
	case "true", "false", "null":
		t.out.WriteString(lit)
		return nil
	case "Infinity", "+Infinity":
		return fmt.Errorf("infinity not supported in standard JSON")
	case "-Infinity":
		return fmt.Errorf("-infinity not supported in standard JSON")
	case "NaN", "+NaN", "-NaN":
		return fmt.Errorf("unsupported value: NaN")
	}

	return fmt.Errorf("invalid literal or identifier: %s", lit)
}

func (t *transcoder) skipWhitespace() {
	for t.pos < len(t.data) {
		r, width := utf8.DecodeRune(t.data[t.pos:])
		if isWhitespace(r) {
			t.pos += width
			continue
		}

		// Comments
		if r == '/' {
			if t.pos+1 < len(t.data) {
				if t.data[t.pos+1] == '/' {
					// Single line
					t.pos += 2
					for t.pos < len(t.data) {
						if t.data[t.pos] == '\n' || t.data[t.pos] == '\r' {
							break
						}
						t.pos++
					}
					continue
				} else if t.data[t.pos+1] == '*' {
					// Multi line
					t.pos += 2
					for t.pos < len(t.data) {
						if t.data[t.pos] == '*' && t.pos+1 < len(t.data) && t.data[t.pos+1] == '/' {
							t.pos += 2
							break
						}
						t.pos++
					}
					continue
				}
			}
		}
		break
	}
}

func isWhitespace(r rune) bool {
	// JSON5 whitespace includes additional unicode characters
	// For simplicity, we use unicode.IsSpace which covers most.
	// Spec: \t, \v, \f, \u0020, \u00A0, \uFEFF, and LineTerminator
	return unicode.IsSpace(r) || r == '\uFEFF'
}

func isIdentifierStart(r rune) bool {
	return unicode.IsLetter(r) || r == '$' || r == '_'
}

func isIdentifierPart(r rune) bool {
	return isIdentifierStart(r) || unicode.IsDigit(r)
}

func isNumberStart(r rune) bool {
	return unicode.IsDigit(r) || r == '+' || r == '-' || r == '.'
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
