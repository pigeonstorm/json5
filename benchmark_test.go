package json5

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"
)

// BenchData mirrors the small‑data benchmark struct used elsewhere.
type BenchData struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Count       int                    `json:"count"`
	Enabled     bool                   `json:"enabled"`
	Tags        []string               `json:"tags"`
	Meta        map[string]interface{} `json:"meta"`
}

func getBenchData() BenchData {
	return BenchData{
		Name:        "Benchmark",
		Description: "Testing the performance of JSON5 vs JSON",
		Count:       1000,
		Enabled:     true,
		Tags:        []string{"go", "json", "json5", "benchmark"},
		Meta: map[string]interface{}{
			"author":  "Victor",
			"version": 1.0,
		},
	}
}

// Flag to specify a JSON file for the large‑data benchmarks.
var jsonFilePath = flag.String("jsonfile", "", "path to a JSON file (10 MB+).")

var (
	largeJSONData []byte // raw JSON (also valid JSON5)
	largeValue    interface{}
)

func TestMain(m *testing.M) {
	flag.Parse()
	if *jsonFilePath == "" {
		fmt.Fprintln(os.Stderr, "jsonfile flag is required for large-data benchmarks")
		os.Exit(1)
	}
	data, err := os.ReadFile(*jsonFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", *jsonFilePath, err)
		os.Exit(1)
	}
	largeJSONData = data
	if err := json.Unmarshal(largeJSONData, &largeValue); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal %s: %v\n", *jsonFilePath, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func BenchmarkUnmarshal_JSON_Large(b *testing.B) {
	var v interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(largeJSONData, &v); err != nil {
			b.Fatalf("json unmarshal error: %v", err)
		}
	}
}

func BenchmarkUnmarshal_JSON5_Large(b *testing.B) {
	var v interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Unmarshal(largeJSONData, &v); err != nil {
			b.Fatalf("json5 unmarshal error: %v", err)
		}
	}
}

func BenchmarkMarshal_JSON_Large(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(largeValue); err != nil {
			b.Fatalf("json marshal error: %v", err)
		}
	}
}

func BenchmarkMarshal_JSON5_Large(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Marshal(largeValue); err != nil {
			b.Fatalf("json5 marshal error: %v", err)
		}
	}
}

// Existing small‑data benchmarks retained for reference.
func BenchmarkMarshal_JSON(b *testing.B) {
	data := getBenchData()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_JSON5(b *testing.B) {
	data := getBenchData()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Marshal(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_JSON(b *testing.B) {
	data := getBenchData()
	bytes, _ := json.Marshal(data)
	var v BenchData
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(bytes, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_JSON5_StandardInput(b *testing.B) {
	data := getBenchData()
	bytes, _ := json.Marshal(data)
	var v BenchData
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Unmarshal(bytes, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_JSON5_JSON5Input(b *testing.B) {
	// JSON5 input with unquoted keys and single quotes
	input := `{\n    name: 'Benchmark',\n    description: 'Testing the performance of JSON5 vs JSON',\n    count: 1000,\n    enabled: true,\n    tags: ['go', 'json', 'json5', 'benchmark'],\n    meta: {\n        author: 'Victor',\n        version: 1.0\n    }\n}`
	bytes := []byte(input)
	var v BenchData
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Unmarshal(bytes, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func TestDataUsageComparison(t *testing.T) {
	data := getBenchData()
	// Standard JSON
	jsonBytes, _ := json.Marshal(data)
	// JSON5 (Manually constructed compact version)
	json5String := `{name:\"Benchmark\",description:\"Testing the performance of JSON5 vs JSON\",count:1000,enabled:true,tags:[\"go\",\"json\",\"json5\",\"benchmark\"],meta:{author:\"Victor\",version:1}}`
	fmt.Printf("\nData Usage Comparison:\n")
	fmt.Printf("Standard JSON Size: %d bytes\n", len(jsonBytes))
	fmt.Printf("Compact JSON5 Size: %d bytes\n", len(json5String))
	fmt.Printf("Savings: %d bytes (%.2f%%)\n", len(jsonBytes)-len(json5String), float64(len(jsonBytes)-len(json5String))/float64(len(jsonBytes))*100)
}
