package service

import (
	"context"
	"errors"

	"github.com/remiehneppo/be-task-management/config"
	"github.com/sashabaranov/go-openai"
)

type VLMService interface {
	ChatMultiContent(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error)
}

type vlmService struct {
	client *openai.Client
	model  string
}

func NewVLMService(vlmCfg config.VLMConfig) VLMService {
	config := openai.DefaultConfig(vlmCfg.APIKey)
	config.BaseURL = vlmCfg.BaseUrl
	client := openai.NewClientWithConfig(config)
	return &vlmService{
		client: client,
		model:  vlmCfg.Model,
	}

}

func (s *vlmService) ChatMultiContent(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	req := openai.ChatCompletionRequest{
		Messages: messages,
		Model:    s.model,
	}

	resp, err := s.client.CreateChatCompletion(
		ctx,
		req,
	)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("no response generated")
	}
	return resp.Choices[0].Message.Content, nil
}
