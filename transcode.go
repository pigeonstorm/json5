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
	// Pre-allocate buffer with some headroom for JSON5->JSON conversion
	t.out.Grow(len(data) + len(data)/10)

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
		// Fast path for ASCII identifiers
		for t.pos < len(t.data) {
			c := t.data[t.pos]
			// Fast ASCII check
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '$' || c == '_' {
				t.pos++
				continue
			}
			// Slow path for unicode
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

	start := t.pos
	for t.pos < len(t.data) {
		c := t.data[t.pos]
		
		// Fast path: check for quote (end of string)
		if c == quote {
			// Write any accumulated bytes before the quote
			if t.pos > start {
				t.out.Write(t.data[start:t.pos])
			}
			t.pos++
			t.out.WriteByte('"')
			return nil
		}

		// Fast path: check for escape
		if c == '\\' {
			// Write any accumulated bytes before the escape
			if t.pos > start {
				t.out.Write(t.data[start:t.pos])
			}
			if t.pos+1 >= len(t.data) {
				return fmt.Errorf("unexpected EOF in string escape")
			}
			next := t.data[t.pos+1]
			if next == '\n' || next == '\r' {
				// Line continuation, skip backslash and newline
				t.pos++ // skip backslash
				if next == '\r' && t.pos+1 < len(t.data) && t.data[t.pos+1] == '\n' {
					t.pos += 2
				} else {
					t.pos++
				}
				start = t.pos
				continue
			} else if next == '\'' && quote == '"' {
				// Escaped single quote in double quoted string
				t.pos += 2
				t.out.WriteByte('\'')
				start = t.pos
				continue
			} else if next == '\'' {
				t.pos += 2
				t.out.WriteByte('\'')
				start = t.pos
				continue
			}

			t.out.WriteByte('\\')
			t.out.WriteByte(next)
			t.pos += 2
			start = t.pos
			continue
		}

		// Fast path: unescaped " inside ' string
		if c == '"' && quote == '\'' {
			// Write any accumulated bytes before the quote
			if t.pos > start {
				t.out.Write(t.data[start:t.pos])
			}
			t.out.WriteByte('\\')
			t.out.WriteByte('"')
			t.pos++
			start = t.pos
			continue
		}

		// Fast path: ASCII character (most common case)
		if c < 0x80 {
			t.pos++
			continue
		}

		// Slow path: multi-byte UTF-8 character
		r, width := utf8.DecodeRune(t.data[t.pos:])
		if r == utf8.RuneError {
			return fmt.Errorf("invalid UTF-8 in string")
		}
		// Write accumulated bytes and the rune
		if t.pos > start {
			t.out.Write(t.data[start:t.pos])
		}
		t.out.WriteRune(r)
		t.pos += width
		start = t.pos
	}
	
	// Write any remaining bytes
	if t.pos > start {
		t.out.Write(t.data[start:t.pos])
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
			c := t.data[t.pos]
			// Fast path for ASCII hex digits
			if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
				t.pos++
				continue
			}
			// Slow path for unicode (shouldn't happen for hex, but be safe)
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
		c := t.data[t.pos]
		// Fast path for ASCII digits
		if c >= '0' && c <= '9' {
			t.pos++
		} else if c == '.' {
			if hasDecimal {
				break // Second dot?
			}
			hasDecimal = true
			t.pos++
		} else if c == 'e' || c == 'E' {
			t.pos++
			if t.pos < len(t.data) && (t.data[t.pos] == '+' || t.data[t.pos] == '-') {
				t.pos++
			}
			// Expect digits (fast path for ASCII)
			for t.pos < len(t.data) {
				c2 := t.data[t.pos]
				if c2 >= '0' && c2 <= '9' {
					t.pos++
				} else {
					// Check for unicode digits (slow path)
					r2, w2 := utf8.DecodeRune(t.data[t.pos:])
					if !unicode.IsDigit(r2) {
						break
					}
					t.pos += w2
				}
			}
			break // End of number
		} else {
			// Check for unicode digits (slow path)
			r, width := utf8.DecodeRune(t.data[t.pos:])
			if unicode.IsDigit(r) {
				t.pos += width
			} else {
				break
			}
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
		c := t.data[t.pos]
		// Fast path for ASCII whitespace
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			t.pos++
			continue
		}
		// Check for comments before doing expensive rune decode
		if c == '/' {
			if t.pos+1 < len(t.data) {
				if t.data[t.pos+1] == '/' {
					// Single line comment
					t.pos += 2
					for t.pos < len(t.data) {
						if t.data[t.pos] == '\n' || t.data[t.pos] == '\r' {
							break
						}
						t.pos++
					}
					continue
				} else if t.data[t.pos+1] == '*' {
					// Multi line comment
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
		// Check for other unicode whitespace (slower path)
		r, width := utf8.DecodeRune(t.data[t.pos:])
		if isWhitespace(r) {
			t.pos += width
			continue
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
