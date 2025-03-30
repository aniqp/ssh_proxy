package proxy

import (
	"context"
	"errors"
	"fmt"
	"github.com/aniqp/formal_assessment/pkg/config"
	"github.com/gliderlabs/ssh"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"log"
	"os"
	"time"
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

func (llm *LLMService) summarizeSession(ctx context.Context, f *os.File, s ssh.Session) error {
	data, err := os.ReadFile(f.Name())
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
		return fmt.Errorf("error while generating llm summary %v", err)
	}
	response := chatCompletion.Choices[0].Message.Content

	_, err = f.WriteString("\n\n" + response)
	if err != nil {
		return fmt.Errorf("error writing to file %v", err)
	}
	log.Printf("Session summary available in logs.")
	return nil
}
