package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type request struct {
	Model   string    `json:"model"`
	Message []Message `json:"messages"`
}
type response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

func getMessage(message1 string) string {
	client := &http.Client{}
	reqe := request{
		Model: "gpt-3.5-turbo",
		Message: []Message{
			{
				Role:    "user",
				Content: message1,
			},
		},
	}
	by, _ := json.Marshal(&reqe)
	fmt.Println(string(by))
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(by))
	req.Header.Set("Authorization", "Bearer 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")
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	message, _ := io.ReadAll(resp.Body)
	var response response
	err = json.Unmarshal(message, &response)
	if err != nil {
		panic(err)
	}
	return response.Choices[0].Message.Content
}
