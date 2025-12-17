package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"orbis/abstractions/ollama"
	"orbis/abstractions/prettifier"
	"os"
)

func main() {

	client := ollama.NewClient()

	var input string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Prompt: ")
	if scanner.Scan() {
		input = scanner.Text()
	}

	fileCont, err := os.ReadFile("prompt_list_project_directories.txt")
	if err != nil {
		return
	}

	finalPrompt := fmt.Sprintf(string(fileCont), input)

	response, err := client.Generate(
		context.Background(),
		finalPrompt,
	)
	if err != nil {
		panic(err)
	}

	data := []byte(response)

	errr := prettifier.PrintProjectTree(data)
	if errr != nil {
		log.Fatal(err)
	}
}
