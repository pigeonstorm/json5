package json5

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestTranscode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string // Expected JSON (optional, if empty we just check validity)
		wantErr bool
	}{
		{
			name:  "Basic Object",
			input: `{key: "value"}`,
			want:  `{"key":"value"}`,
		},
		{
			name:  "Trailing Comma Object",
			input: `{key: "value",}`,
			want:  `{"key":"value"}`,
		},
		{
			name:  "Basic Array",
			input: `[1, 2, 3]`,
			want:  `[1,2,3]`,
		},
		{
			name:  "Trailing Comma Array",
			input: `[1, 2, 3,]`,
			want:  `[1,2,3]`,
		},
		{
			name: "Comments",
			input: `{
				// Single line
				key: /* Multi
				line */ "value"
			}`,
			want: `{"key":"value"}`,
		},
		{
			name:  "Single Quotes",
			input: `{'key': 'value'}`,
			want:  `{"key":"value"}`,
		},
		{
			name:  "Hex Numbers",
			input: `[0x1, 0x10, 0xFF]`,
			want:  `[1,16,255]`,
		},
		{
			name:  "Leading Decimal",
			input: `[.5, .123]`,
			want:  `[0.5,0.123]`,
		},
		{
			name:  "Trailing Decimal",
			input: `[5., 10.]`,
			want:  `[5.0,10.0]`,
		},
		{
			name:  "Explicit Plus",
			input: `[+5, +1.5]`,
			want:  `[5,1.5]`,
		},
		{
			name: "Escaped Newline in String",
			input: `{"key": "line\
continued"}`,
			want: `{"key":"linecontinued"}`,
		},
		{
			name: "Complex Nested",
			input: `{
                foo: 'bar',
                arr: [
                    1,
                    // comment
                    2,
                ],
                nested: {
                    a: .5
                }
            }`,
			want: `{"foo":"bar","arr":[1,2],"nested":{"a":0.5}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Transcode([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Transcode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				// Compact expected JSON for comparison if provided
				if tt.want != "" {
					var buf bytes.Buffer
					if err := json.Compact(&buf, []byte(tt.want)); err != nil {
						t.Fatalf("Invalid expected JSON: %v", err)
					}
					wantCompact := buf.String()

					// Compact actual JSON
					buf.Reset()
					if err := json.Compact(&buf, got); err != nil {
						t.Errorf("Transcode() produced invalid JSON: %s, error: %v", got, err)
						return
					}
					gotCompact := buf.String()

					if gotCompact != wantCompact {
						t.Errorf("Transcode() = %s, want %s", gotCompact, wantCompact)
					}
				} else {
					// Just check if it is valid JSON
					if !json.Valid(got) {
						t.Errorf("Transcode() produced invalid JSON: %s", got)
					}
				}
			}
		})
	}
}
