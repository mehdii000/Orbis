package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	defaultURL   = "http://localhost:11434/api/generate"
	defaultModel = "qwen2.5-coder"
)

type Client struct {
	http  *http.Client
	url   string
	model string
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 5 * time.Minute, // FIXED: Increased from 30s to 5min for LLM generation
		},
		url:   defaultURL,
		model: defaultModel,
	}
}

type request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type response struct {
	Response string `json:"response"`
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := request{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.url,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.New("ollama returned " + res.Status)
	}

	var out response
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}

	return out.Response, nil
}

func (c *Client) GenerateFromFile(filename string, inputs ...string) string {
	fileCont, err := os.ReadFile(filename)
	if err != nil {
		panic("Error reading file: " + err.Error())
	}

	finalPrompt := fmt.Sprintf(string(fileCont), inputs)

	response, err := c.Generate(context.Background(), finalPrompt)
	if err != nil {
		panic("Error generating response: " + err.Error())
	}

	return response
}
