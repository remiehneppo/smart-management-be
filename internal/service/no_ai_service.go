package service

import (
	"context"

	"github.com/remiehneppo/be-task-management/types"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

var _ AIService = (AIService)(nil)

type NoAIService struct {
}

func NewNoAIService() *NoAIService {
	logrus.Info("No AI service is enabled")
	return &NoAIService{}
}

func (s *NoAIService) Chat(ctx context.Context, messages []types.Message) (*types.Message, error) {
	return &types.Message{
		Content: "AI service is not available",
		Role:    openai.ChatMessageRoleAssistant,
	}, nil
}
