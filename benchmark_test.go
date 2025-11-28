package json5

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// LargeData represents a complex nested structure for benchmarking
type LargeData struct {
	Users      []User      `json:"users"`
	Products   []Product   `json:"products"`
	Orders     []Order     `json:"orders"`
	Metadata   Metadata    `json:"metadata"`
	Config     BenchConfig `json:"config"`
	Statistics Statistics  `json:"statistics"`
}

type User struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Age         int      `json:"age"`
	Active      bool     `json:"active"`
	Tags        []string `json:"tags"`
	Preferences map[string]interface{} `json:"preferences"`
}

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	InStock     bool    `json:"inStock"`
	Description string  `json:"description"`
	Attributes  map[string]interface{} `json:"attributes"`
}

type Order struct {
	ID        int    `json:"id"`
	UserID    int    `json:"userId"`
	ProductID int    `json:"productId"`
	Quantity  int    `json:"quantity"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

type Metadata struct {
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Tags        []string          `json:"tags"`
	Properties  map[string]string `json:"properties"`
}

type BenchConfig struct {
	Timeout    int    `json:"timeout"`
	Retries    int    `json:"retries"`
	Enabled    bool   `json:"enabled"`
	LogLevel   string `json:"logLevel"`
	Endpoints  []string `json:"endpoints"`
}

type Statistics struct {
	TotalUsers    int64   `json:"totalUsers"`
	TotalProducts int64   `json:"totalProducts"`
	TotalOrders   int64   `json:"totalOrders"`
	Revenue       float64 `json:"revenue"`
	AverageOrder  float64 `json:"averageOrder"`
}

var (
	largeDataJSON  []byte
	largeDataJSON5 []byte
	largeDataValue interface{}
	targetSizeMB   = 12 // Target > 10MB
)

func init() {
	// Generate synthetic data
	data := generateLargeData(targetSizeMB)
	
	// Marshal to JSON
	var err error
	largeDataJSON, err = json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal JSON: %v", err))
	}
	
	// Create JSON5 version (manually construct with JSON5 features)
	largeDataJSON5 = generateJSON5(data)
	
	// Unmarshal JSON to get the value for unmarshal benchmarks
	if err := json.Unmarshal(largeDataJSON, &largeDataValue); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal JSON: %v", err))
	}
	
	// Verify we have > 10MB
	jsonSizeMB := float64(len(largeDataJSON)) / (1024 * 1024)
	json5SizeMB := float64(len(largeDataJSON5)) / (1024 * 1024)
	fmt.Printf("Generated data sizes:\n")
	fmt.Printf("  JSON:  %.2f MB (%d bytes)\n", jsonSizeMB, len(largeDataJSON))
	fmt.Printf("  JSON5: %.2f MB (%d bytes)\n", json5SizeMB, len(largeDataJSON5))
	fmt.Printf("  Size difference: %.2f%%\n", 
		(float64(len(largeDataJSON5)-len(largeDataJSON))/float64(len(largeDataJSON)))*100)
}

func generateLargeData(targetMB int) *LargeData {
	// Calculate approximate items needed for target size
	// Rough estimate: each user ~500 bytes, each product ~400 bytes, each order ~200 bytes
	// Plus overhead from structure
	usersCount := (targetMB * 1024 * 1024) / 500
	productsCount := usersCount / 2
	ordersCount := usersCount * 3
	
	// Cap at reasonable numbers
	if usersCount > 50000 {
		usersCount = 50000
		productsCount = 25000
		ordersCount = 150000
	}
	
	data := &LargeData{
		Users:    make([]User, usersCount),
		Products: make([]Product, productsCount),
		Orders:   make([]Order, ordersCount),
		Metadata: Metadata{
			Version:     "1.0.0",
			Environment: "production",
			Tags:        []string{"benchmark", "test", "large", "data"},
			Properties: map[string]string{
				"region":    "us-east-1",
				"datacenter": "dc1",
				"cluster":   "prod-cluster-01",
			},
		},
		Config: BenchConfig{
			Timeout:  30,
			Retries:  3,
			Enabled:  true,
			LogLevel: "info",
			Endpoints: []string{
				"https://api.example.com/v1",
				"https://api.example.com/v2",
				"https://api-backup.example.com/v1",
			},
		},
		Statistics: Statistics{
			TotalUsers:    int64(usersCount),
			TotalProducts: int64(productsCount),
			TotalOrders:   int64(ordersCount),
			Revenue:       1234567.89,
			AverageOrder:  99.99,
		},
	}
	
	// Generate users
	names := []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry"}
	categories := []string{"Electronics", "Clothing", "Food", "Books", "Toys", "Home", "Sports", "Automotive"}
	statuses := []string{"pending", "completed", "shipped", "cancelled"}
	
	for i := range data.Users {
		theme := "light"
		if i%2 == 0 {
			theme = "dark"
		}
		language := "es"
		if i%3 == 0 {
			language = "en"
		}
		data.Users[i] = User{
			ID:     i + 1,
			Name:   fmt.Sprintf("%s %s%d", names[i%len(names)], "User", i),
			Email:  fmt.Sprintf("user%d@example.com", i),
			Age:    20 + (i % 50),
			Active: i%3 != 0,
			Tags:   []string{fmt.Sprintf("tag%d", i%10), fmt.Sprintf("category%d", i%5)},
			Preferences: map[string]interface{}{
				"theme":         theme,
				"notifications": i%2 == 0,
				"language":      language,
			},
		}
	}
	
	// Generate products
	for i := range data.Products {
		data.Products[i] = Product{
			ID:          i + 1,
			Name:        fmt.Sprintf("Product %d: %s Item", i+1, categories[i%len(categories)]),
			Price:       9.99 + float64(i%1000),
			Category:    categories[i%len(categories)],
			InStock:     i%4 != 0,
			Description: fmt.Sprintf("This is a detailed description for product %d with many features and specifications.", i+1),
			Attributes: map[string]interface{}{
				"weight":    1.5 + float64(i%10),
				"dimensions": fmt.Sprintf("%dx%dx%d", 10+i%20, 10+i%20, 10+i%20),
				"color":     []string{"red", "blue", "green"}[i%3],
			},
		}
	}
	
	// Generate orders
	for i := range data.Orders {
		data.Orders[i] = Order{
			ID:        i + 1,
			UserID:    (i % usersCount) + 1,
			ProductID: (i % productsCount) + 1,
			Quantity:  1 + (i % 10),
			Status:    statuses[i%len(statuses)],
			Timestamp: 1609459200 + int64(i*3600), // Start from 2021-01-01
		}
	}
	
	return data
}

func generateJSON5(data *LargeData) []byte {
	// This is a simplified JSON5 generator for benchmarking
	// In practice, you'd use a proper JSON5 encoder
	// For now, we'll create a compact JSON5-like version
	
	var b strings.Builder
	b.WriteString("{\n")
	
	// Users array
	b.WriteString("  users: [\n")
	for i, user := range data.Users {
		if i > 0 {
			b.WriteString(",\n")
		}
		b.WriteString(fmt.Sprintf("    {id: %d, name: '%s', email: '%s', age: %d, active: %t, tags: [", 
			user.ID, user.Name, user.Email, user.Age, user.Active))
		for j, tag := range user.Tags {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("'%s'", tag))
		}
		b.WriteString("], preferences: {")
		first := true
		for k, v := range user.Preferences {
			if !first {
				b.WriteString(", ")
			}
			first = false
			if str, ok := v.(string); ok {
				b.WriteString(fmt.Sprintf("%s: '%s'", k, str))
			} else if bval, ok := v.(bool); ok {
				b.WriteString(fmt.Sprintf("%s: %t", k, bval))
			}
		}
		b.WriteString("}}")
	}
	b.WriteString("\n  ],\n")
	
	// Products array
	b.WriteString("  products: [\n")
	for i, product := range data.Products {
		if i > 0 {
			b.WriteString(",\n")
		}
		b.WriteString(fmt.Sprintf("    {id: %d, name: '%s', price: %.2f, category: '%s', inStock: %t, description: '%s', attributes: {", 
			product.ID, product.Name, product.Price, product.Category, product.InStock, product.Description))
		first := true
		for k, v := range product.Attributes {
			if !first {
				b.WriteString(", ")
			}
			first = false
			if str, ok := v.(string); ok {
				b.WriteString(fmt.Sprintf("%s: '%s'", k, str))
			} else if fval, ok := v.(float64); ok {
				b.WriteString(fmt.Sprintf("%s: %.2f", k, fval))
			}
		}
		b.WriteString("}}")
	}
	b.WriteString("\n  ],\n")
	
	// Orders array (compact)
	b.WriteString("  orders: [\n")
	for i, order := range data.Orders {
		if i > 0 {
			b.WriteString(",\n")
		}
		b.WriteString(fmt.Sprintf("    {id: %d, userId: %d, productId: %d, quantity: %d, status: '%s', timestamp: %d}", 
			order.ID, order.UserID, order.ProductID, order.Quantity, order.Status, order.Timestamp))
	}
	b.WriteString("\n  ],\n")
	
	// Metadata
	b.WriteString("  metadata: {version: '1.0.0', environment: 'production', tags: ['benchmark', 'test', 'large', 'data'], properties: {region: 'us-east-1', datacenter: 'dc1', cluster: 'prod-cluster-01'}},\n")
	
	// Config
	b.WriteString("  config: {timeout: 30, retries: 3, enabled: true, logLevel: 'info', endpoints: ['https://api.example.com/v1', 'https://api.example.com/v2', 'https://api-backup.example.com/v1']},\n")
	
	// Statistics
	b.WriteString(fmt.Sprintf("  statistics: {totalUsers: %d, totalProducts: %d, totalOrders: %d, revenue: %.2f, averageOrder: %.2f}\n", 
		data.Statistics.TotalUsers, data.Statistics.TotalProducts, data.Statistics.TotalOrders, 
		data.Statistics.Revenue, data.Statistics.AverageOrder))
	
	b.WriteString("}\n")
	return []byte(b.String())
}

// Benchmark Marshal operations
func BenchmarkMarshal_JSON_Comprehensive(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(largeDataValue); err != nil {
			b.Fatalf("json marshal error: %v", err)
		}
	}
}

func BenchmarkMarshal_JSON5_Comprehensive(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := Marshal(largeDataValue); err != nil {
			b.Fatalf("json5 marshal error: %v", err)
		}
	}
}

// Benchmark Unmarshal operations
func BenchmarkUnmarshal_JSON_Comprehensive(b *testing.B) {
	var v interface{}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(largeDataJSON, &v); err != nil {
			b.Fatalf("json unmarshal error: %v", err)
		}
	}
}

func BenchmarkUnmarshal_JSON5_Comprehensive(b *testing.B) {
	var v interface{}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := Unmarshal(largeDataJSON5, &v); err != nil {
			b.Fatalf("json5 unmarshal error: %v", err)
		}
	}
}

// Test function to generate benchmark summary
func TestBenchmarkSummary(t *testing.T) {
	jsonSize := len(largeDataJSON)
	json5Size := len(largeDataJSON5)
	sizeDiff := jsonSize - json5Size
	sizeDiffPercent := (float64(sizeDiff) / float64(jsonSize)) * 100
	
	fmt.Printf("\n=== Benchmark Data Summary ===\n")
	fmt.Printf("JSON Size:  %d bytes (%.2f MB)\n", jsonSize, float64(jsonSize)/(1024*1024))
	fmt.Printf("JSON5 Size: %d bytes (%.2f MB)\n", json5Size, float64(json5Size)/(1024*1024))
	fmt.Printf("Size Difference: %d bytes (%.2f%%)\n", sizeDiff, sizeDiffPercent)
	fmt.Printf("\nRun benchmarks with: go test -bench=. -benchmem\n")
}

