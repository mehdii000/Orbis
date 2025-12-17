package prettifier

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Project represents the top-level structure of your input data.
type Project struct {
	Language    string                 `json:"language"`
	Framework   string                 `json:"framework"`
	Description string                 `json:"description"`
	Structure   map[string]interface{} `json:"structure"`
}

// PrintProjectTree takes a JSON byte slice and prints the formatted tree.
func PrintProjectTree(jsonData []byte) error {
	var p Project
	if err := json.Unmarshal(jsonData, &p); err != nil {
		return fmt.Errorf("failed to parse project JSON: %w", err)
	}

	fmt.Printf("🚀 %s (%s Project)\n", p.Framework, p.Language)
	fmt.Printf("📝 %s\n\n", p.Description)

	// Start recursion with the Structure map
	renderTree(p.Structure, "")
	return nil
}

// renderTree is unexported (private) to keep the package API clean.
func renderTree(m map[string]interface{}, indent string) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, key := range keys {
		isLast := i == len(keys)-1
		connector := "├── "
		newIndent := indent + "│   "

		if isLast {
			connector = "└── "
			newIndent = indent + "    "
		}

		// UI Logic: Choose icon based on naming convention
		symbol := "📄 "
		if strings.HasSuffix(key, "/") {
			symbol = "📁 "
		}

		fmt.Printf("%s%s%s%s\n", indent, connector, symbol, key)

		// Recursive step: if value is a map, it's a directory
		if subMap, ok := m[key].(map[string]interface{}); ok {
			renderTree(subMap, newIndent)
		}
	}
}
