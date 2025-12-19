package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"orbis/abstractions/ollama"
	"orbis/abstractions/prettifier"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const ROOT_FOLDER = "project"

func walkThroughDirs(m map[string]interface{}, currentPath string) {
	// 1. Sort keys for deterministic folder creation
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := m[key]
		// Join the current path with the new key (e.g., "root" + "subdir")
		fullPath := filepath.Join(currentPath, key)

		// 2. Check if the value is a nested map (a directory)
		subMap, isDirectory := value.(map[string]interface{})

		if isDirectory {
			// Create the directory
			err := os.MkdirAll(fullPath, os.ModePerm)
			if err != nil {
				fmt.Printf("Error creating directory %s: %v\n", fullPath, err)
				continue
			}
			// 3. Recursive call: Pass the new fullPath as the base for the next level
			walkThroughDirs(subMap, fullPath)
		} else {
			// 4. Create the file
			fmt.Println("Creating file:", fullPath)
			file, err := os.Create(fullPath)
			if err != nil {
				fmt.Printf("Error creating file %s: %v\n", fullPath, err)
				continue
			}
			file.Close() // Close immediately to avoid "too many open files" in recursion
		}
	}
}

func main() {

	client := ollama.NewClient()

	var input string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Prompt: ")
	if scanner.Scan() {
		input = scanner.Text()
	} else {
		fmt.Println("Error reading input:", scanner.Err())
		return
	}

	projectDirectories := client.GenerateFromFile("prompt_list_project_directories.txt", input)

	jsonData := []byte(projectDirectories)
	prettifier.PrintProjectTree(jsonData)

	jsonStr := string(jsonData)
	if strings.HasPrefix(jsonStr, "```json") {
		jsonStr = strings.TrimPrefix(jsonStr, "```json")
		jsonStr = strings.TrimPrefix(jsonStr, "```")
		jsonStr = strings.TrimSuffix(jsonStr, "```")
		jsonStr = strings.TrimSpace(jsonStr)
		jsonData = []byte(jsonStr)
	}

	var p prettifier.Project

	if err := json.Unmarshal(jsonData, &p); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	os.Mkdir(ROOT_FOLDER, os.ModePerm)
	walkThroughDirs(p.Structure, ROOT_FOLDER)

	fmt.Println("---")

	contextInterface := client.GenerateFromFile("prompt_create_context_interface.txt", projectDirectories)

	fmt.Println(contextInterface)

}
