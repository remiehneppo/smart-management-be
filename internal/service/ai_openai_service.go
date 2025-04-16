package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/remiehneppo/be-task-management/config"
	"github.com/remiehneppo/be-task-management/types"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// var (
// 	SystemMessageInitiateMechanicalEngineer = openai.ChatCompletionMessage{
// 		Role: openai.ChatMessageRoleSystem,
// 		Content: `You are an AI technical assistant for the X52 factory (Nhà máy X52). Your task is to support and answer technical questions related to the operation, maintenance, repair, and optimization of equipment and production processes in the factory.

// You always respond in Vietnamese with accurate, clear, and concise answers. If in-depth information is available, you can provide detailed explanations to help users understand the issue thoroughly.

// If a question falls outside your area of expertise or there is not enough data to answer, politely inform the user instead of making assumptions.

// Always maintain a professional, polite, and helpful approach when assisting users.
// `,
// 	}
// )

var _ AIService = (*OpenAIService)(nil)

type AIService interface {
	Chat(ctx context.Context, messages []types.Message) (*types.Message, error)
}

type OpenAIService struct {
	systemPromt   string
	client        *openai.Client
	allowTool     bool
	functionsCall map[string]types.FunctionHandler
	tools         []openai.Tool
	model         string
}

func NewOpenAIService(aiCfg config.OpenaiConfig) *OpenAIService {
	config := openai.DefaultConfig(aiCfg.APIKey)
	config.BaseURL = aiCfg.BaseUrl
	client := openai.NewClientWithConfig(config)
	return &OpenAIService{
		systemPromt:   aiCfg.SystemPrompt,
		client:        client,
		functionsCall: make(map[string]types.FunctionHandler),
		tools:         make([]openai.Tool, 0),
		model:         aiCfg.Model,
		allowTool:     aiCfg.AllowTool,
	}
}

func (s *OpenAIService) Chat(ctx context.Context, messages []types.Message) (*types.Message, error) {
	// Convert our Message type to OpenAI chat messages
	openaiMessages := make([]openai.ChatCompletionMessage, 0)
	openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: s.systemPromt,
	})
	for _, msg := range messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: msg.Content,
		})
	}
	req := openai.ChatCompletionRequest{
		Messages: openaiMessages,
		Model:    s.model,
	}
	if s.allowTool {
		req.Tools = s.tools
	}
	// Create chat completion request
	resp, err := s.client.CreateChatCompletion(
		ctx,
		req,
	)

	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response generated")
	}

	if resp.Choices[0].FinishReason == openai.FinishReasonToolCalls {

		resp, err = s.handleFunctionCall(ctx, openaiMessages, resp)
		if err != nil {
			return nil, err
		}

	}

	// Convert response back to our Message type
	return &types.Message{
		Role:    openai.ChatMessageRoleAssistant,
		Content: resp.Choices[0].Message.Content,
	}, nil
}

// func (s *OpenAIService) ChatStream(ctx context.Context, messages []types.Message, streamHandler types.StreamHandler) error {
// 	// Convert our Message type to OpenAI chat messages
// 	openaiMessages := make([]openai.ChatCompletionMessage, 0)
// 	openaiMessages = append(openaiMessages, SystemMessageInitiateMechanicalEngineer)
// 	for _, msg := range messages {
// 		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
// 			Role:    openai.ChatMessageRoleUser,
// 			Content: msg.Content,
// 		})
// 	}

// 	// Create chat completion request
// 	stream, err := s.client.CreateChatCompletionStream(
// 		ctx,
// 		openai.ChatCompletionRequest{
// 			Messages: openaiMessages,
// 			// Tools:    s.tools,
// 			Model: s.model,
// 		},
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	defer stream.Close()
// 	for {
// 		resp, err := stream.Recv()

// 		if err != nil {
// 			if err == io.EOF {
// 				return nil
// 			}
// 			log.Println("Error receiving response from stream:", err)
// 		}
// 		streamHandler(resp.Choices[0].Delta.Content)
// 	}

// }

func (s *OpenAIService) RegisterFunctionCall(name, description string, params jsonschema.Definition, handler types.FunctionHandler) error {
	f := openai.FunctionDefinition{
		Name:        name,
		Description: description,
		Parameters:  params,
	}
	t := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &f,
	}
	s.functionsCall[name] = handler
	s.tools = append(s.tools, t)
	return nil
}

func (s *OpenAIService) handleFunctionCall(ctx context.Context, openaiMessages []openai.ChatCompletionMessage, resp openai.ChatCompletionResponse) (openai.ChatCompletionResponse, error) {
	openaiMessages = append(openaiMessages, resp.Choices[0].Message)
	for _, toolCall := range resp.Choices[0].Message.ToolCalls {
		if toolCall.Type == openai.ToolTypeFunction {
			handler := s.functionsCall[toolCall.Function.Name]
			if handler == nil {
				return openai.ChatCompletionResponse{}, errors.New("no handler found for function call")
			}
			result, err := handler(ctx, []byte(toolCall.Function.Arguments))
			if err != nil {
				return openai.ChatCompletionResponse{}, err
			}
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result.(string),
				Name:       toolCall.Function.Name,
				ToolCallID: toolCall.ID,
			})
		}
	}
	req := openai.ChatCompletionRequest{
		Messages: openaiMessages,
		Model:    s.model,
	}
	if s.allowTool {
		req.Tools = s.tools
	}

	resp, err := s.client.CreateChatCompletion(
		ctx,
		req,
	)
	if err != nil {
		return openai.ChatCompletionResponse{}, err
	}
	if len(resp.Choices) == 0 {
		return openai.ChatCompletionResponse{}, errors.New("no response generated")
	}
	if resp.Choices[0].FinishReason == openai.FinishReasonToolCalls {
		return s.handleFunctionCall(ctx, openaiMessages, resp)
	}
	return resp, nil
}

func createRetrieveDocumentPrompt(ctx, question string) string {
	return fmt.Sprintf(`
Use the following CONTEXT to answer the QUESTION at the end.
If you don't know the answer, just say that you don't know, don't try to make up an answer.
Use an unbiased and journalistic tone.

CONTEXT: %s

QUESTION: %s`, ctx, question)
}
