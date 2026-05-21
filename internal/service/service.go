package service

import (
	"greentrust-hackathon/internal/repository"
	"greentrust-hackathon/pkg/bcrypt"
	"greentrust-hackathon/pkg/jwt"
)

type Service struct {
	UserService IUserService
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface) *Service {
	userService := NewUserService(repository.UserRepository, repository.OTPRepository, jwtAuth, bcrypt, repository.UserIdentityRepository)
	return &Service{
		UserService: userService,
	}
}
