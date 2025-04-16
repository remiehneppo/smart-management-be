package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/be-task-management/internal/service"
	"github.com/remiehneppo/be-task-management/types"
)

var _ AIAssistantHandler = (*aiAssistantHandler)(nil)

type AIAssistantHandler interface {
	ChatWithAssistant(ctx *gin.Context)
	ChatWithAssistantStateless(ctx *gin.Context)
}

type aiAssistantHandler struct {
	aiAssistantService service.AIAssistantService
}

func NewAIAssistantHandler(aiAssistantService service.AIAssistantService) *aiAssistantHandler {
	return &aiAssistantHandler{
		aiAssistantService: aiAssistantService,
	}
}

func (h *aiAssistantHandler) ChatWithAssistant(ctx *gin.Context) {
	var req types.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid request",
		})
		return
	}
	res, err := h.aiAssistantService.ChatWithAssistant(
		ctx,
		req,
	)
	if err != nil {
		ctx.JSON(500, types.Response{
			Status:  false,
			Message: "Internal server error",
		})
		return
	}
	data := types.Response{
		Status:  true,
		Message: "Success",
		Data: types.ChatResponse{
			Content: res.Content,
		},
	}
	ctx.JSON(200, data)
}

// ChatWithAssistantStateless godoc
// @Summary Chat with Assistant Stateless
// @Description Chat with Assistant Stateless
// @Tags assistant
// @Accept json
// @Produce json
// @Param request body types.ChatStatelessRequest true "Chat request"
// @Success 200 {object} types.Response{data=types.ChatResponse} "Success"
// @Failure 401 {object} types.Response "Unauthorized"
// @Failure 400 {object} types.Response "Invalid request"
// @Failure 500 {object} types.Response "Internal server error"
// @Security BearerAuth
// @Router /assistant/chat-stateless [post]
func (h *aiAssistantHandler) ChatWithAssistantStateless(ctx *gin.Context) {
	var req types.ChatStatelessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid request",
		})
		return
	}
	res, err := h.aiAssistantService.ChatWithAssistantStateless(
		ctx,
		req,
	)
	if err != nil {
		ctx.JSON(500, types.Response{
			Status:  false,
			Message: "Internal server error",
		})
		return
	}
	data := types.Response{
		Status:  true,
		Message: "Success",
		Data: types.ChatResponse{
			Content: res.Content,
		},
	}
	ctx.JSON(200, data)
}
