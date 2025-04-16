package service

import (
	"context"

	"github.com/remiehneppo/be-task-management/types"
)

var _ AIAssistantService = (*aiAssistantService)(nil)

type AIAssistantService interface {
	ChatWithAssistant(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error)
	ChatWithAssistantStateless(ctx context.Context, req types.ChatStatelessRequest) (*types.ChatResponse, error)
	PaginateMessages(ctx context.Context, req types.PaginateMessagesRequest) ([]*types.ChatMessage, int64, error)
}

type aiAssistantService struct {
	aiService AIService
}

func NewAIAssistantService(aiService AIService) *aiAssistantService {
	return &aiAssistantService{
		aiService: aiService,
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
	res, err := s.aiService.Chat(
		ctx,
		req.Messages,
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
