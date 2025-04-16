package types

import "context"

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// FunctionHandler is a type for handling function calls
type FunctionHandler func(ctx context.Context, args []byte) (any, error)
