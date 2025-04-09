package proxy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aniqp/formal_assessment/pkg/config"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
)

type LLMService struct {
	cfg    *config.Config
	client openai.Client
}

func NewLLMService(cfg *config.Config, client openai.Client) *LLMService {
	return &LLMService{cfg: cfg, client: client}
}

func (llm *LLMService) validateLLMKey() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// simple call to test if key is valid (no "key validate" method in OpenAI Go SDK yet)
	_, err := llm.client.Models.List(ctx)
	if err != nil {
		var apiErr *openai.Error
		if errors.As(err, &apiErr) && apiErr.Code == "invalid_api_key" {
			log.Printf("LLM key validation failed: %v", apiErr.Message)
			llm.cfg.LLM.Enabled = false
		}
		return
	}

	llm.cfg.LLM.Enabled = true
	log.Println("LLM key validated successfully")
}

func (llm *LLMService) summarizeSession(ctx context.Context, filePath string, username string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading log file: %v", err)
	}
	combinedPrompt := fmt.Sprintf("%s: %s", llm.cfg.LLM.Prompt, data)
	chatCompletion, err := llm.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(combinedPrompt),
		},
		Model:               openai.ChatModelGPT4oMini,
		Temperature:         param.NewOpt(0.3),
		MaxCompletionTokens: param.NewOpt(int64(300)),
	})
	if err != nil {
		return fmt.Errorf("error while generating LLM summary: %v", err)
	}

	response := chatCompletion.Choices[0].Message.Content
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error reopening file to append: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n\n" + response); err != nil {
		return fmt.Errorf("error writing summary: %v", err)
	}

	log.Printf("LLM session summary for user %s appended to log file.", username)
	return nil
}
