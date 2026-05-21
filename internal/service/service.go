package service

import (
	"greentrust-hackathon/internal/blockchain"
	"greentrust-hackathon/internal/repository"
	"greentrust-hackathon/pkg/bcrypt"
	"greentrust-hackathon/pkg/jwt"
	"greentrust-hackathon/pkg/supabase"
	"log"
)

type Service struct {
	UserService          IUserService
	EvidenceService      IEvidenceService
	GreenPassportService IGreenPassportService
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface, supabase supabase.Interface) *Service {
	chainClient, err := blockchain.NewEthereumClientFromEnv()
	if err != nil {
		log.Printf("warning: blockchain client disabled: %v", err)
	}

	userService := NewUserService(repository.UserRepository, repository.OTPRepository, jwtAuth, bcrypt, repository.UserIdentityRepository, repository.BusinessSectorRepository, repository.UMKMProfileRepository, repository.LocationPhotoRepository, supabase)
	evidenceService := NewEvidenceService(repository.UMKMProfileRepository, repository.EvidenceRepository, supabase)
	greenPassportService := NewGreenPassportService(repository.UMKMProfileRepository, repository.EvidenceRepository, repository.GreenPassportRepository, chainClient, supabase)

	return &Service{
		UserService:          userService,
		EvidenceService:      evidenceService,
		GreenPassportService: greenPassportService,
	}
}
