package service

import (
	"greentrust-hackathon/internal/repository"
	"greentrust-hackathon/pkg/bcrypt"
	"greentrust-hackathon/pkg/jwt"
)

type Service struct {
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface) *Service {
	return &Service{}
}
