package service

import (
	"context"

	"github.com/remiehneppo/be-task-management/internal/repository"
	"github.com/remiehneppo/be-task-management/types"
)

type LoginService interface {
	Login(ctx context.Context, req types.LoginRequest) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context) error
	Refresh(ctx context.Context, oldRefreshToken string) (accessToken, refreshToken string, err error)
}

type loginService struct {
	jwtService JWTService
	userRepo   repository.UserRepository
}

func NewLoginService(jwtService JWTService, userRepo repository.UserRepository) LoginService {
	return &loginService{
		jwtService: jwtService,
		userRepo:   userRepo,
	}
}

func (s *loginService) Login(ctx context.Context, req types.LoginRequest) (accessToken, refreshToken string, err error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return "", "", err
	}

	if user.Password != req.Password {
		return "", "", types.ErrInvalidCredentials
	}

	// Generate tokens
	refreshToken, err = s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	accessToken, err = s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *loginService) Logout(ctx context.Context) error {
	// Invalidate the refresh token in the database
	// This is a placeholder implementation
	return nil
}

func (s *loginService) Refresh(ctx context.Context, oldRefreshToken string) (accessToken, refreshToken string, err error) {
	user, err := s.jwtService.ValidateRefreshToken(oldRefreshToken)
	if err != nil {
		return "", "", err
	}

	// Generate new tokens
	refreshToken, err = s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	accessToken, err = s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
