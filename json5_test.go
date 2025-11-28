package json5

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

type Config struct {
	Name    string `json:"name"`
	Count   int    `json:"count"`
	Enabled bool   `json:"enabled"`
	List    []int  `json:"list"`
}

func TestUnmarshal(t *testing.T) {
	input := `{
		name: 'Test',
		count: 0x10,
		enabled: true,
		list: [1, 2, 3,], // Trailing comma
	}`

	var cfg Config
	if err := Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	want := Config{
		Name:    "Test",
		Count:   16,
		Enabled: true,
		List:    []int{1, 2, 3},
	}

	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("Unmarshal got %+v, want %+v", cfg, want)
	}
}

func TestNewDecoder(t *testing.T) {
	input := `{
		name: "Stream",
		count: 10
	}`
	r := strings.NewReader(input)
	dec := NewDecoder(r)

	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if cfg.Name != "Stream" || cfg.Count != 10 {
		t.Errorf("Decode got %+v", cfg)
	}
}

func TestMarshal(t *testing.T) {
	cfg := Config{
		Name:    "Marshal",
		Count:   20,
		Enabled: true,
		List:    []int{4, 5, 6},
	}

	data, err := Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Should be standard JSON
	want := `{"name":"Marshal","count":20,"enabled":true,"list":[4,5,6]}`
	if string(data) != want {
		t.Errorf("Marshal got %s, want %s", string(data), want)
	}
}

func TestMarshalIndent(t *testing.T) {
	cfg := Config{
		Name:  "Indent",
		Count: 30,
	}

	data, err := MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent failed: %v", err)
	}

	want := `{
  "name": "Indent",
  "count": 30,
  "enabled": false,
  "list": null
}`
	if string(data) != want {
		t.Errorf("MarshalIndent got:\n%s\nwant:\n%s", string(data), want)
	}
}

func TestFunctional(t *testing.T) {
	data, err := os.ReadFile("examples/basic/example.json5")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	var result map[string]interface{}
	if err := Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify specific values
	tests := []struct {
		key  string
		want interface{}
	}{
		{"unquoted", "and you can quote me on that"},
		{"singleQuotes", "I can use \"double quotes\" here"},
		{"lineBreaks", "Look, Mom! No \\n's!"},
		{"hexadecimal", float64(0xdecaf)}, // JSON numbers are float64
		{"leadingDecimalPoint", 0.8675309},
		{"andTrailing", 8675309.0},
		{"positiveSign", 1.0},
		{"trailingComma", "in objects"},
		{"backwardsCompatible", "with JSON"},
	}

	for _, tt := range tests {
		got, ok := result[tt.key]
		if !ok {
			t.Errorf("Key %s not found", tt.key)
			continue
		}
		if got != tt.want {
			t.Errorf("Key %s = %v, want %v", tt.key, got, tt.want)
		}
	}

	// Verify array
	arr, ok := result["andIn"].([]interface{})
	if !ok {
		t.Fatalf("andIn is not an array")
	}
	if len(arr) != 1 || arr[0] != "arrays" {
		t.Errorf("andIn = %v, want ['arrays']", arr)
	}

	// Marshal back and check validity
	out, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if !json.Valid(out) {
		t.Errorf("Marshaled output is not valid JSON: %s", out)
	}
}
