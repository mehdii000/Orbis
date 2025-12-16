package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const OLLAMA_URL string = "http://localhost:11434/api/generate"
const MODEL string = "qwen2.5-coder"

func main() {

	// Prepare JSON request body
	requestBody := map[string]any{
		"model":  MODEL,
		"prompt": "tell me a joke.",
		"stream": false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		panic(err)
	}

	// Send POST request
	res, err := http.Post(OLLAMA_URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Read response
	responseBytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Response status:", res.Status)

	// Parse JSON response
	var responseMap map[string]interface{}
	if err := json.Unmarshal(responseBytes, &responseMap); err != nil {
		panic(err)
	}

	// Access the "response" field
	if val, ok := responseMap["response"]; ok {
		fmt.Println("Response:", val)
	} else {
		fmt.Println("Field 'response' not found in JSON")
	}

}
