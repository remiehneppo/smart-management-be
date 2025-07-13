package service

import (
	"context"

	"github.com/remiehneppo/be-task-management/types"
	"github.com/sashabaranov/go-openai"
)

var _ AIAssistantService = (*aiAssistantService)(nil)

type AIAssistantService interface {
	ChatWithAssistant(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error)
	ChatWithAssistantStateless(ctx context.Context, req types.ChatStatelessRequest) (*types.ChatResponse, error)
	PaginateMessages(ctx context.Context, req types.PaginateMessagesRequest) ([]*types.ChatMessage, int64, error)
}

type aiAssistantService struct {
	aiService    AIService
	systemPrompt string
}

func NewAIAssistantService(aiService AIService, systemPrompt string) *aiAssistantService {
	return &aiAssistantService{
		aiService:    aiService,
		systemPrompt: systemPrompt,
	}
}

func (s *aiAssistantService) ChatWithAssistant(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	// TODO: Implement the chat with assistant logic
	// Load the chat history from the database
	// Call the AI service to get a response
	// Save the response to the database
	// Return the response
	// For now, just return a dummy response
	return &types.ChatResponse{
		Content: "Dummy response from AI assistant",
	}, nil
}

func (s *aiAssistantService) ChatWithAssistantStateless(ctx context.Context, req types.ChatStatelessRequest) (*types.ChatResponse, error) {
	messages := make([]types.Message, 0)
	// Add system prompt if it exists
	if s.systemPrompt != "" {
		messages = append(messages, types.Message{
			Role:    openai.ChatMessageRoleSystem,
			Content: s.systemPrompt,
		})
	}
	messages = append(messages, req.Messages...)
	res, err := s.aiService.Chat(
		ctx,
		messages,
	)
	if err != nil {
		return nil, err
	}
	return &types.ChatResponse{
		Content: res.Content,
	}, nil
}

func (s *aiAssistantService) PaginateMessages(ctx context.Context, req types.PaginateMessagesRequest) ([]*types.ChatMessage, int64, error) {
	// TODO: Implement the pagination logic
	// Load the messages from the database
	// Return the messages and the total count
	// For now, just return a dummy response
	return nil, 0, nil
}
