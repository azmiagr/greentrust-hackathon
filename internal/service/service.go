package service

import (
	"greentrust-hackathon/internal/repository"
	"greentrust-hackathon/pkg/bcrypt"
	"greentrust-hackathon/pkg/jwt"
	"greentrust-hackathon/pkg/supabase"
)

type Service struct {
	UserService     IUserService
	EvidenceService IEvidenceService
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface, supabase supabase.Interface) *Service {
	userService := NewUserService(repository.UserRepository, repository.OTPRepository, jwtAuth, bcrypt, repository.UserIdentityRepository, repository.BusinessSectorRepository, repository.UMKMProfileRepository, repository.LocationPhotoRepository, supabase)
	evidenceService := NewEvidenceService(repository.UMKMProfileRepository, repository.EvidenceRepository, supabase)
	return &Service{
		UserService:     userService,
		EvidenceService: evidenceService,
	}
}
