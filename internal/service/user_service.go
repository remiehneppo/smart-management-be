package service

import (
	"context"

	"github.com/remiehneppo/be-task-management/internal/repository"
	"github.com/remiehneppo/be-task-management/types"
)

type UserService interface {
	GetUserInfo(ctx context.Context, id string) (*types.User, error)
	UpdateUserInfo(ctx context.Context, id string, user *types.User) error
	UpdatePassword(ctx context.Context, id string, oldPassword, newPassword string) error
	GetUsersInWorkspace(ctx context.Context, workspace string) ([]*types.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetUserInfo(ctx context.Context, id string) (*types.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	user.Password = "" // Clear password before returning
	// This is important for security reasons
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) UpdateUserInfo(ctx context.Context, id string, user *types.User) error {
	err := s.userRepo.Update(ctx, id, user)
	if err != nil {
		return err
	}
	return nil
}

func (s *userService) UpdatePassword(ctx context.Context, id string, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if user.Password != oldPassword {
		return types.ErrInvalidCredentials
	}

	user.Password = newPassword

	err = s.userRepo.Update(ctx, id, user)
	if err != nil {
		return err
	}
	return nil
}

func (s *userService) GetUsersInWorkspace(ctx context.Context, workspace string) ([]*types.User, error) {

	users, err := s.userRepo.FindByWorkspace(ctx, workspace)
	if err != nil {
		return nil, err
	}
	return users, nil
}
