package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/be-task-management/internal/logger"
	"github.com/remiehneppo/be-task-management/internal/service"
	"github.com/remiehneppo/be-task-management/types"
)

type LoginHandler interface {
	Login(ctx *gin.Context)
	Logout(ctx *gin.Context)
	Refresh(ctx *gin.Context)
}

type loginHandler struct {
	loginService service.LoginService
	logger       *logger.Logger
}

func NewLoginHandler(loginService service.LoginService, logger *logger.Logger) LoginHandler {
	return &loginHandler{
		loginService: loginService,
		logger:       logger,
	}
}

// Login godoc
// @Summary User login
// @Description Authenticates user and returns access and refresh tokens
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body types.LoginRequest true "User credentials"
// @Success 200 {object} types.Response{data=types.LoginResponse} "Login successful"
// @Failure 400 {object} types.Response "Invalid request"
// @Failure 401 {object} types.Response "Authentication failed"
// @Router /auth/login [post]
func (h *loginHandler) Login(ctx *gin.Context) {
	req := &types.LoginRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	accessToken, refreshToken, err := h.loginService.Login(ctx, *req)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(401, res)
		return
	}
	res := types.Response{
		Status:  true,
		Message: "Login successful",
		Data: types.LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}
	ctx.JSON(200, res)
}

// Logout godoc
// @Summary User logout
// @Description Logs out the current user
// @Tags authentication
// @Accept json
// @Produce json
// @Success 200 {object} types.Response "Logout successful"
// @Failure 401 {object} types.Response "Unauthorized"
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *loginHandler) Logout(ctx *gin.Context) {
	// Implementation needed
	res := types.Response{
		Status:  true,
		Message: "Logout successful",
	}
	ctx.JSON(200, res)
}

// Refresh godoc
// @Summary Refresh tokens
// @Description Refreshes access token using a valid refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param refresh body types.RefreshRequest true "Refresh token"
// @Success 200 {object} types.Response{data=types.LoginResponse} "Refresh successful"
// @Failure 400 {object} types.Response "Invalid request"
// @Failure 401 {object} types.Response "Invalid refresh token"
// @Router /auth/refresh [post]
func (h *loginHandler) Refresh(ctx *gin.Context) {
	req := &types.RefreshRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	accessToken, refreshToken, err := h.loginService.Refresh(ctx, req.RefreshToken)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(401, res)
		return
	}
	res := types.Response{
		Status:  true,
		Message: "Refresh successful",
		Data: types.LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}
	ctx.JSON(200, res)
}
