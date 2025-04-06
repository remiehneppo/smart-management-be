package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/be-task-management/internal/service"
	"github.com/remiehneppo/be-task-management/types"
)

const BearerPrefix = "bearer "

type AuthMiddleware struct {
	jwtService service.JWTService
}

func NewAuthMiddleware(jwtService service.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

func (a *AuthMiddleware) AuthBearerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		accessToken := ctx.GetHeader("Authorization")
		if accessToken == "" {
			res := types.Response{
				Status:  false,
				Message: "Authorization header is missing",
			}
			ctx.JSON(401, res)
			ctx.Abort()
			return
		}

		// Check if the token has the Bearer prefix
		if len(accessToken) < len(BearerPrefix) || strings.ToLower(accessToken[:len(BearerPrefix)]) != BearerPrefix {
			res := types.Response{
				Status:  false,
				Message: "Invalid token format. Expected Bearer token",
			}
			ctx.JSON(401, res)
			ctx.Abort()
			return
		}

		// Remove "Bearer " prefix
		accessToken = accessToken[len(BearerPrefix):]

		user, err := a.jwtService.ValidateAccessToken(accessToken)
		if err != nil {
			res := types.Response{
				Status:  false,
				Message: "Invalid token",
			}
			ctx.JSON(401, res)
			ctx.Abort()
			return
		}

		ctx.Set("user_id", user.ID)
		ctx.Next()
	}
}
