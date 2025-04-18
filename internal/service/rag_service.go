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
	prompt := "Based on the extracted document chunks below, answer the question concisely and accurately in Vietnamese.\n" +
		"Only use information from relevant chunks and ignore unrelated or uncertain content.\n\n" +
		"Question: " + question + "\n\n" +
		"Document Chunks:\n"
	for _, chunk := range chunks {
		prompt += fmt.Sprintf("[Title: %s, Page: %d]\n%s\n\n", chunk.Title, chunk.PageNumber, chunk.Content)
	}
	return prompt

}
