package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/remiehneppo/be-task-management/types"
)

type JWTService interface {
	GenerateRefreshToken(user *types.User) (string, error)
	ValidateRefreshToken(token string) (*types.User, error)
	GetUserIdFromRefreshToken(token string) (string, error)
	GenerateAccessToken(user *types.User) (string, error)
	ValidateAccessToken(token string) (*types.User, error)
	GetUserIdFromAccessToken(token string) (string, error)
}

type jwtRefreshClaims struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type jwtAccessClaims struct {
	UserId          string `json:"user_id"`
	Username        string `json:"username"`
	ManagementLevel int    `json:"management_level"`
	WorkspaceRole   string `json:"workspace_role"`
	Workspace       string `json:"workspace"`
	jwt.RegisteredClaims
}

type jwtService struct {
	secretKey string
	issuer    string
	exp       int64
}

func NewJWTService(secretKey, issuer string, exp int64) JWTService {
	return &jwtService{
		secretKey: secretKey,
		issuer:    issuer,
		exp:       exp,
	}
}

func (j *jwtService) GenerateRefreshToken(user *types.User) (string, error) {
	claims := &jwtRefreshClaims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.exp))), // 1 day expiration
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (j *jwtService) ValidateRefreshToken(token string) (*types.User, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &jwtRefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := parsedToken.Claims.(*jwtRefreshClaims); ok && parsedToken.Valid {
		return &types.User{Username: claims.Username, ID: claims.Subject}, nil
	}
	return nil, jwt.ErrInvalidKey
}

func (j *jwtService) GetUserIdFromRefreshToken(token string) (string, error) {
	// Parse without validation by using ParseUnverified
	parsedToken, _, err := jwt.NewParser().ParseUnverified(token, &jwtAccessClaims{})
	if err != nil {
		return "", err
	}
	if claims, ok := parsedToken.Claims.(*jwtRefreshClaims); ok && parsedToken.Valid {
		return claims.Subject, nil
	}
	return "", jwt.ErrInvalidKey
}

func (j *jwtService) GenerateAccessToken(user *types.User) (string, error) {
	claims := &jwtAccessClaims{
		UserId:          user.ID,
		Username:        user.Username,
		ManagementLevel: user.ManagementLevel,
		WorkspaceRole:   user.WorkspaceRole,
		Workspace:       user.Workspace,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(6 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (j *jwtService) ValidateAccessToken(token string) (*types.User, error) {

	parsedToken, err := jwt.ParseWithClaims(token, &jwtAccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := parsedToken.Claims.(*jwtAccessClaims); ok && parsedToken.Valid {
		return &types.User{
			Username:        claims.Username,
			ID:              claims.Subject,
			ManagementLevel: claims.ManagementLevel,
			WorkspaceRole:   claims.WorkspaceRole,
			Workspace:       claims.Workspace,
		}, nil
	}
	return nil, jwt.ErrInvalidKey
}

func (j *jwtService) GetUserIdFromAccessToken(token string) (string, error) {
	// Parse without validation by using ParseUnverified
	parsedToken, _, err := jwt.NewParser().ParseUnverified(token, &jwtAccessClaims{})
	if err != nil {
		return "", err
	}
	if claims, ok := parsedToken.Claims.(*jwtAccessClaims); ok && parsedToken.Valid {
		return claims.Subject, nil
	}
	return "", jwt.ErrInvalidKey
}
