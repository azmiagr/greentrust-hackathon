package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"greentrust-hackathon/entity"
	"greentrust-hackathon/internal/blockchain"
	"greentrust-hackathon/internal/repository"
	"greentrust-hackathon/model"
	"greentrust-hackathon/pkg/database/mariadb"
	apperrors "greentrust-hackathon/pkg/errors"
	"greentrust-hackathon/pkg/qrcode"
	"greentrust-hackathon/pkg/supabase"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IGreenPassportService interface {
	IssueGreenPassport(userID uuid.UUID) (*model.IssueGreenPassportResponse, error)
}

type GreenPassportService struct {
	db               *gorm.DB
	umkmProfileRepo  repository.IUMKMProfileRepository
	evidenceRepo     repository.IEvidenceRepository
	passportRepo     repository.IGreenPassportRepository
	blockchainClient blockchain.Client
	supabase         supabase.Interface
}

func NewGreenPassportService(
	umkmRepo repository.IUMKMProfileRepository,
	evidenceRepo repository.IEvidenceRepository,
	passportRepo repository.IGreenPassportRepository,
	blockchainClient blockchain.Client,
	supabaseClient supabase.Interface,
) IGreenPassportService {
	return &GreenPassportService{
		db:               mariadb.Connection,
		umkmProfileRepo:  umkmRepo,
		evidenceRepo:     evidenceRepo,
		passportRepo:     passportRepo,
		blockchainClient: blockchainClient,
		supabase:         supabaseClient,
	}
}

func (s *GreenPassportService) IssueGreenPassport(userID uuid.UUID) (*model.IssueGreenPassportResponse, error) {
	if s.blockchainClient == nil {
		return nil, apperrors.InternalServer("blockchain client is not configured")
	}

	profile, err := s.umkmProfileRepo.GetUMKMProfile(s.db, model.GetUMKMProfileParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("business profile not found")
		}
		return nil, apperrors.InternalServer("failed to get business profile")
	}

	if existing, err := s.passportRepo.GetByProfileID(s.db, profile.ProfileID); err == nil && existing.Status == "active" {
		return nil, apperrors.Conflict("green passport already active")
	}

	categories, err := s.evidenceRepo.GetEvidenceCategoriesWithDocuments(s.db, profile.ProfileID)
	if err != nil {
		return nil, apperrors.InternalServer("failed to calculate grs")
	}

	grs, breakdown := calculatePassportGRS(categories)
	if grs < evidencePassportThreshold {
		return nil, apperrors.BadRequest("grs score is below passport threshold")
	}

	docs, err := s.evidenceRepo.GetReviewedEvidenceDocumentsByProfileID(s.db, profile.ProfileID)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get reviewed documents")
	}

	if len(docs) == 0 {
		return nil, apperrors.BadRequest("reviewed evidence document is required")
	}

	passportID := uuid.New()
	now := time.Now()
	hashes := collectDocumentHashes(docs)

	chainResult, err := s.blockchainClient.IssuePassport(context.Background(), blockchain.IssuePassportInput{
		ProfileID: profile.ProfileID.String(), PassportID: passportID.String(), DocumentHashes: hashes, GRSScore: grs, IssuedAt: now,
	})
	if err != nil {
		_ = s.evidenceRepo.MarkDocumentsBlockchainFailed(s.db, profile.ProfileID)
		return nil, apperrors.InternalServer("failed to issue green passport on blockchain")
	}

	passport := &entity.GreenPassport{
		PassportID:       passportID,
		ProfileID:        profile.ProfileID,
		GRSScore:         grs,
		GRSBreakdown:     breakdown,
		Status:           "active",
		PublicSlug:       buildPassportSlug(profile.BusinessName, passportID),
		IssuedAt:         now,
		LastUpdatedAt:    now,
		BlockchainTxHash: chainResult.TxHash,
		ContractAddress:  os.Getenv("GREEN_PASSPORT_CONTRACT_ADDRESS"),
		NetworkName:      getEnvWithDefault("BLOCKCHAIN_NETWORK_NAME", "Local Anvil"),
		ChainID:          getInt64EnvWithDefault("BLOCKCHAIN_CHAIN_ID", 31337),
		BlockNumber:      chainResult.BlockNumber,
	}

	passportURL := buildPassportURL(passport.PublicSlug)

	qrPNG, err := qrcode.GeneratePNG(passportURL, 512)
	if err != nil {
		return nil, apperrors.InternalServer("failed to generate green passport qr code")
	}

	qrURL, err := s.supabase.UploadBytes(
		qrPNG,
		fmt.Sprintf("green-passports/qr/%s.png", passport.PassportID.String()),
		"image/png",
		"31536000",
	)
	if err != nil {
		return nil, apperrors.InternalServer("failed to upload green passport qr code")
	}

	passport.QRCodeURL = qrURL

	tx := s.db.Begin()
	defer tx.Rollback()

	err = s.passportRepo.UpsertPassport(tx, passport)
	if err != nil {
		return nil, apperrors.InternalServer("failed to save green passport")
	}

	err = s.evidenceRepo.MarkDocumentsOnChain(tx, profile.ProfileID, chainResult.TxHash)
	if err != nil {
		return nil, apperrors.InternalServer("failed to update evidence blockchain status")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, apperrors.InternalServer("failed to commit green passport")
	}

	return mapIssueGreenPassportResponse(passport, len(docs)), nil
}

func buildPassportURL(publicSlug string) string {
	baseURL := strings.TrimRight(os.Getenv("PUBLIC_PASSPORT_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:5173/passport"
	}

	return fmt.Sprintf("%s/%s", baseURL, publicSlug)
}

func collectDocumentHashes(docs []*entity.EvidenceDocument) []string {
	hashes := make([]string, 0, len(docs))
	for _, doc := range docs {
		hashes = append(hashes, doc.FileHash)
	}
	return hashes
}

func calculatePassportGRS(categories []*entity.EvidenceCategory) (float64, entity.GRSBreakdown) {
	total := 0.0
	breakdown := make(entity.GRSBreakdown, 0, len(categories))
	for _, category := range categories {
		progress := buildCategoryProgress(category)
		total += progress.Score
		breakdown = append(breakdown, entity.GRSCategoryBreakdown{
			Category: category.CategoryID, Weight: category.Weight,
			CI: progress.Score, DocCount: progress.FulfilledCount, DocsNeeded: progress.RequiredCount,
		})
	}
	return roundScore(total), breakdown
}

func buildPassportSlug(name string, id uuid.UUID) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "green-passport"
	}
	return fmt.Sprintf("%s-%s", slug, id.String()[:8])
}

func mapIssueGreenPassportResponse(passport *entity.GreenPassport, documentCount int) *model.IssueGreenPassportResponse {
	return &model.IssueGreenPassportResponse{
		PassportID:       passport.PassportID,
		ProfileID:        passport.ProfileID,
		GRSScore:         passport.GRSScore,
		Status:           passport.Status,
		PublicSlug:       passport.PublicSlug,
		PassportURL:      buildPassportURL(passport.PublicSlug),
		QRCodeURL:        passport.QRCodeURL,
		DocumentCount:    documentCount,
		CategoryScores:   mapGreenPassportCategoryScores(passport.GRSBreakdown),
		Network:          passport.NetworkName,
		ChainID:          passport.ChainID,
		ContractAddress:  passport.ContractAddress,
		BlockchainTxHash: passport.BlockchainTxHash,
		BlockNumber:      passport.BlockNumber,
		IssuedAt:         passport.IssuedAt,
	}
}

func mapGreenPassportCategoryScores(breakdown entity.GRSBreakdown) []model.GreenPassportCategoryScore {
	scores := make([]model.GreenPassportCategoryScore, 0, len(breakdown))

	for _, item := range breakdown {
		percent := 0.0
		if item.DocsNeeded > 0 {
			percent = roundScore((float64(item.DocCount) / float64(item.DocsNeeded)) * 100)
		}

		scores = append(scores, model.GreenPassportCategoryScore{
			CategoryID: item.Category,
			Score:      item.CI,
			Percent:    percent,
			DocCount:   item.DocCount,
			DocsNeeded: item.DocsNeeded,
		})
	}

	return scores
}

func getEnvWithDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getInt64EnvWithDefault(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}
