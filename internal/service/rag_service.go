package service

import (
	"context"
	"fmt"

	"github.com/remiehneppo/be-task-management/types"
	"github.com/sashabaranov/go-openai"
)

var _ RAGService = (*ragService)(nil)

type RAGService interface {
	AskAI(ctx context.Context, question string, chunks []*types.ChunkDocumentResponse) (string, error)
}

type ragService struct {
	aiService    AIService
	systemPrompt string
}

func NewRAGService(aiService AIService, systemPrompt string) *ragService {
	return &ragService{
		aiService:    aiService,
		systemPrompt: systemPrompt,
	}
}

func (s *ragService) AskAI(ctx context.Context, question string, chunks []*types.ChunkDocumentResponse) (string, error) {
	prompt := s.ragPrompt(question, chunks)
	message, err := s.aiService.Chat(
		ctx,
		[]types.Message{
			{Role: openai.ChatMessageRoleSystem, Content: s.systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	)
	if err != nil {
		return "", err
	}
	return message.Content, nil
}

func (s *ragService) ragPrompt(question string, chunks []*types.ChunkDocumentResponse) string {
	prompt := "Bạn là một trợ lý AI thông minh có khả năng đọc và hiểu thông tin từ ngữ cảnh được cung cấp bên dưới.\n" +
		"Hãy sử dụng thông tin trong phần ngữ cảnh để trả lời câu hỏi phía dưới bằng tiếng Việt.\n\n" +
		"NGỮ CẢNH:\n{{"
	for _, chunk := range chunks {
		prompt += fmt.Sprintf("[Tên tài liệu: %s, Trang: %d\nNội dung: %s]\n\n", chunk.Title, chunk.PageNumber, chunk.Content)
	}
	prompt += "}}\n\n"
	prompt += "CÂU HỎI: {{" + question + "}}\n\n"
	return prompt

}
