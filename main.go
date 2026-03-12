package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	// Parse flags
	format := flag.String("f", "json", "output format: json or yaml")
	query := flag.String("q", "", "dot-notation path (e.g., .user.name)")
	flag.Parse()

	// Read input
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}

	// Detect input format
	var data interface{}
	inputStr := strings.TrimSpace(string(input))
	if strings.HasPrefix(inputStr, "{") || strings.HasPrefix(inputStr, "[") {
		if err := json.Unmarshal(input, &data); err != nil {
			fmt.Fprintf(os.Stderr, "invalid JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := yaml.Unmarshal(input, &data); err != nil {
			fmt.Fprintf(os.Stderr, "invalid YAML: %v\n", err)
			os.Exit(1)
		}
	}

	// Apply query if provided
	if *query != "" {
		parts := strings.Split((*query)[1:], ".")
		cur := data
		for _, part := range parts {
			if m, ok := cur.(map[string]interface{}); ok {
				if v, exists := m[part]; exists {
					cur = v
				} else {
					fmt.Fprintf(os.Stderr, "key not found: %s\n", part)
					os.Exit(1)
				}
			} else if arr, ok := cur.([]interface{}); ok {
				idx := 0
				fmt.Sscanf(part, "%d", &idx)
				if idx >= 0 && idx < len(arr) {
					cur = arr[idx]
				} else {
					fmt.Fprintf(os.Stderr, "array index out of range: %d\n", idx)
					os.Exit(1)
				}
			} else {
				fmt.Fprintf(os.Stderr, "cannot traverse: %v\n", cur)
				os.Exit(1)
			}
		}
		data = cur
	}

	// Output in requested format
	if *format == "json" {
		out, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error marshaling JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
	} else {
		out, err := yaml.Marshal(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error marshaling YAML: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
	}
}
