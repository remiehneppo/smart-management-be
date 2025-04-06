package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/be-task-management/internal/logger"
	"github.com/remiehneppo/be-task-management/internal/service"
	"github.com/remiehneppo/be-task-management/types"
)

type UserHandler interface {
	GetUserInfo(ctx *gin.Context)
	UpdatePassword(ctx *gin.Context)
	GetUsersSameWorkspace(ctx *gin.Context)
}

type userHandler struct {
	userService service.UserService
	logger      *logger.Logger
}

func NewUserHandler(
	userService service.UserService,
	logger *logger.Logger,
) UserHandler {
	return &userHandler{
		userService: userService,
		logger:      logger,
	}
}

// GetUserInfo godoc
// @Summary Get user information
// @Description Returns the authenticated user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} types.Response{data=types.User} "User information retrieved successfully"
// @Failure 401 {object} types.Response "Unauthorized or invalid token"
// @Security BearerAuth
// @Router /users/me [get]
func (h *userHandler) GetUserInfo(ctx *gin.Context) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		res := types.Response{
			Status:  false,
			Message: types.ErrInvalidCredentials.Error(),
		}
		ctx.JSON(401, res)
		return
	}
	user, err := h.userService.GetUserInfo(ctx, userID)
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
		Message: "User info retrieved successfully",
		Data:    user,
	}
	ctx.JSON(200, res)
}

// UpdatePassword godoc
// @Summary Update user password
// @Description Changes the authenticated user's password
// @Tags users
// @Accept json
// @Produce json
// @Param passwordData body types.UpdatePasswordRequest true "Old and new password"
// @Success 200 {object} types.Response "Password updated successfully"
// @Failure 400 {object} types.Response "Invalid request format"
// @Failure 401 {object} types.Response "Unauthorized, invalid token or incorrect old password"
// @Security BearerAuth
// @Router /users/password [post]
func (h *userHandler) UpdatePassword(ctx *gin.Context) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		res := types.Response{
			Status:  false,
			Message: types.ErrInvalidCredentials.Error(),
		}
		ctx.JSON(401, res)
		return
	}
	req := &types.UpdatePasswordRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(400, res)
		return
	}
	err := h.userService.UpdatePassword(ctx, userID, req.OldPassword, req.NewPassword)
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
		Message: "Password updated successfully",
	}
	ctx.JSON(200, res)
}

// GetUsersSameWorkspace godoc
// @Summary Get users in the same workspace
// @Description Returns a list of users in the same workspace as the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} types.Response "Users retrieved successfully"
// @Failure 400 {object} types.Response "Invalid request format"
// @Failure 401 {object} types.Response "Unauthorized or invalid token"
// @Security BearerAuth
// @Router /users/workspace [get]
func (h *userHandler) GetUsersSameWorkspace(ctx *gin.Context) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		res := types.Response{
			Status:  false,
			Message: types.ErrInvalidCredentials.Error(),
		}
		ctx.JSON(401, res)
		return
	}
	user, err := h.userService.GetUserInfo(ctx, userID)
	if err != nil {
		res := types.Response{
			Status:  false,
			Message: err.Error(),
		}
		ctx.JSON(401, res)
		return
	}

	users, err := h.userService.GetUsersInWorkspace(ctx, user.Workspace)
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
		Message: "Users retrieved successfully",
		Data:    users,
	}
	ctx.JSON(200, res)
}
