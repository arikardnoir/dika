package main

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func mains() {
	client := openai.NewClient("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")
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "Can you analyze an image and tell me if the product is authentic or false?",
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}