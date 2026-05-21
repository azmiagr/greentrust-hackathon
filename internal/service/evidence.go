package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"greentrust-hackathon/entity"
	"greentrust-hackathon/internal/repository"
	"greentrust-hackathon/model"
	"greentrust-hackathon/pkg/database/mariadb"
	apperrors "greentrust-hackathon/pkg/errors"
	"greentrust-hackathon/pkg/supabase"
	"io"
	"math"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	evidencePassportThreshold = 70
	maxEvidenceFileSize       = 10 * 1024 * 1024
)

type IEvidenceService interface {
	GetEvidenceSummary(userID uuid.UUID) (*model.EvidenceSummaryResponse, error)
	GetEvidenceCategoryDetail(userID uuid.UUID, categoryID string) (*model.EvidenceCategoryDetailResponse, error)
	UploadEvidenceDocument(userID uuid.UUID, param model.UploadEvidenceDocumentParam) (*model.EvidenceDocumentResponse, error)
	GetEvidenceCategories() ([]model.EvidenceCategoryResponse, error)
	SubmitEvidenceAIReview(userID uuid.UUID, param model.SubmitEvidenceAIReviewParam) (*model.EvidenceAIReviewResponse, error)
	GetEvidenceAIReviews(userID uuid.UUID, status string) ([]model.EvidenceAIReviewResponse, error)
	ReviewEvidenceAI(userID uuid.UUID, param model.ReviewEvidenceAIParam) (*model.EvidenceAIReviewResponse, error)
}

type EvidenceService struct {
	db              *gorm.DB
	umkmProfileRepo repository.IUMKMProfileRepository
	evidenceRepo    repository.IEvidenceRepository
	supabase        supabase.Interface
}

func NewEvidenceService(
	umkmProfileRepo repository.IUMKMProfileRepository,
	evidenceRepo repository.IEvidenceRepository,
	supabase supabase.Interface,
) IEvidenceService {
	return &EvidenceService{
		db:              mariadb.Connection,
		umkmProfileRepo: umkmProfileRepo,
		evidenceRepo:    evidenceRepo,
		supabase:        supabase,
	}
}

func (s *EvidenceService) GetEvidenceSummary(userID uuid.UUID) (*model.EvidenceSummaryResponse, error) {
	profile, err := s.getProfileByUserID(s.db, userID)
	if err != nil {
		return nil, err
	}

	categories, err := s.evidenceRepo.GetEvidenceCategoriesWithDocuments(s.db, profile.ProfileID)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get evidence categories")
	}

	categoryProgress := make([]model.EvidenceCategoryProgress, 0, len(categories))
	totalScore := 0.0

	for _, category := range categories {
		progress := buildCategoryProgress(category)
		categoryProgress = append(categoryProgress, progress)
		totalScore += progress.Score
	}

	passportStatus := "draft"
	if totalScore >= evidencePassportThreshold {
		passportStatus = "eligible"
	}

	return &model.EvidenceSummaryResponse{
		GRSScore:          roundScore(totalScore),
		PassportThreshold: evidencePassportThreshold,
		PassportStatus:    passportStatus,
		Categories:        categoryProgress,
	}, nil
}

func (s *EvidenceService) GetEvidenceCategoryDetail(userID uuid.UUID, categoryID string) (*model.EvidenceCategoryDetailResponse, error) {
	categoryID = strings.ToUpper(strings.TrimSpace(categoryID))
	if categoryID == "" {
		return nil, apperrors.BadRequest("category_id is required")
	}

	profile, err := s.getProfileByUserID(s.db, userID)
	if err != nil {
		return nil, err
	}

	category, err := s.evidenceRepo.GetEvidenceCategoryWithDocuments(s.db, categoryID, profile.ProfileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("evidence category not found")
		}
		return nil, apperrors.InternalServer("failed to get evidence category")
	}

	categoryProgress := buildCategoryProgress(category)

	requirements := make([]model.EvidenceRequirementItem, 0, len(category.Requirements))
	documents := make([]model.EvidenceDocumentItem, 0)

	for _, requirement := range category.Requirements {
		selectedDoc := selectRequirementDocument(requirement.EvidenceDocuments)

		var docItem *model.EvidenceDocumentItem
		if selectedDoc != nil {
			mapped := mapEvidenceDocumentItem(*selectedDoc)
			docItem = &mapped
		}

		requirements = append(requirements, model.EvidenceRequirementItem{
			RequirementID: requirement.RequirementID,
			Name:          requirement.Name,
			Description:   requirement.Description,
			IsRequired:    requirement.IsRequired,
			Document:      docItem,
		})

		for _, doc := range requirement.EvidenceDocuments {
			documents = append(documents, mapEvidenceDocumentItem(doc))
		}
	}

	return &model.EvidenceCategoryDetailResponse{
		Category:     categoryProgress,
		Requirements: requirements,
		Documents:    documents,
		NextPriority: []model.EvidenceCategoryProgress{},
	}, nil
}

func (s *EvidenceService) UploadEvidenceDocument(userID uuid.UUID, param model.UploadEvidenceDocumentParam) (*model.EvidenceDocumentResponse, error) {
	categoryID := strings.ToUpper(strings.TrimSpace(param.CategoryID))
	if categoryID == "" {
		return nil, apperrors.BadRequest("category_id is required")
	}

	if param.File == nil {
		return nil, apperrors.BadRequest("document is required")
	}

	if param.File.Size > maxEvidenceFileSize {
		return nil, apperrors.BadRequest("document must not exceed 10MB")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	profile, err := s.getProfileByUserID(tx, userID)
	if err != nil {
		return nil, err
	}

	requirement, err := s.evidenceRepo.GetEvidenceRequirement(tx, param.RequirementID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.BadRequest("evidence requirement not found")
		}
		return nil, apperrors.InternalServer("failed to get evidence requirement")
	}

	if requirement.CategoryID != categoryID {
		return nil, apperrors.BadRequest("requirement does not belong to selected category")
	}

	fileHash, err := hashMultipartFile(param.File)
	if err != nil {
		return nil, apperrors.InternalServer("failed to hash evidence document")
	}

	fileURL, err := s.supabase.UploadEvidenceFile(param.File)
	if err != nil {
		return nil, apperrors.BadRequest("failed to upload evidence document")
	}

	doc := &entity.EvidenceDocument{
		EvidenceID:    uuid.New(),
		ProfileID:     profile.ProfileID,
		RequirementID: &param.RequirementID,
		FilePath:      fileURL,
		FileHash:      fileHash,
		AiConfidence:  param.AIConfidence,
		OriginalName:  param.File.Filename,
		MimeType:      param.File.Header.Get("Content-Type"),
		FileSize:      param.File.Size,
		Status:        "uploaded",
	}

	err = s.evidenceRepo.CreateEvidenceDocument(tx, doc)
	if err != nil {
		_ = s.supabase.DeleteFile(fileURL)
		return nil, apperrors.InternalServer("failed to save evidence document")
	}

	err = tx.Commit().Error
	if err != nil {
		_ = s.supabase.DeleteFile(fileURL)
		return nil, apperrors.InternalServer("failed to commit evidence document")
	}

	return mapEvidenceDocumentResponse(*doc), nil
}

func (s *EvidenceService) GetEvidenceCategories() ([]model.EvidenceCategoryResponse, error) {
	categories, err := s.evidenceRepo.GetEvidenceCategories(s.db)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get evidence categories")
	}

	responses := make([]model.EvidenceCategoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, mapEvidenceCategoryResponse(category))
	}

	return responses, nil
}

func (s *EvidenceService) SubmitEvidenceAIReview(userID uuid.UUID, param model.SubmitEvidenceAIReviewParam) (*model.EvidenceAIReviewResponse, error) {
	if strings.TrimSpace(param.AISummary) == "" {
		return nil, apperrors.BadRequest("ai_summary is required")
	}

	if param.AIConfidence < 0 || param.AIConfidence > 1 {
		return nil, apperrors.BadRequest("ai_confidence must be between 0 and 1")
	}

	param.SuggestedCategoryID = strings.ToUpper(strings.TrimSpace(param.SuggestedCategoryID))
	if param.SuggestedCategoryID == "" {
		return nil, apperrors.BadRequest("suggested_category_id is required")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	profile, err := s.getProfileByUserID(tx, userID)
	if err != nil {
		return nil, err
	}

	doc, err := s.evidenceRepo.GetEvidenceDocumentByID(tx, param.EvidenceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("evidence document not found")
		}
		return nil, apperrors.InternalServer("failed to get evidence document")
	}

	if doc.ProfileID != profile.ProfileID {
		return nil, apperrors.Forbidden("evidence document does not belong to user")
	}

	if param.SuggestedRequirementID != nil && strings.TrimSpace(*param.SuggestedRequirementID) != "" {
		suggestedRequirementID := strings.TrimSpace(*param.SuggestedRequirementID)
		req, err := s.evidenceRepo.GetEvidenceRequirement(tx, suggestedRequirementID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.BadRequest("suggested requirement not found")
			}
			return nil, apperrors.InternalServer("failed to get suggested requirement")
		}

		if req.CategoryID != param.SuggestedCategoryID {
			return nil, apperrors.BadRequest("suggested requirement does not belong to suggested category")
		}

		param.SuggestedRequirementID = &suggestedRequirementID
	}

	existingReview, err := s.evidenceRepo.GetAIReviewByEvidenceID(tx, param.EvidenceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.InternalServer("failed to get ai review")
	}

	var review *entity.EvidenceAIReview
	if errors.Is(err, gorm.ErrRecordNotFound) {
		review = &entity.EvidenceAIReview{
			ReviewID:               uuid.New(),
			EvidenceID:             doc.EvidenceID,
			ProfileID:              profile.ProfileID,
			AISummary:              param.AISummary,
			AIConfidence:           param.AIConfidence,
			SuggestedCategoryID:    param.SuggestedCategoryID,
			SuggestedRequirementID: param.SuggestedRequirementID,
			ReviewStatus:           "pending",
		}

		if err := s.evidenceRepo.CreateAIReview(tx, review); err != nil {
			return nil, apperrors.InternalServer("failed to create ai review")
		}
	} else {
		review = existingReview
		review.AISummary = param.AISummary
		review.AIConfidence = param.AIConfidence
		review.SuggestedCategoryID = param.SuggestedCategoryID
		review.SuggestedRequirementID = param.SuggestedRequirementID
		review.FinalRequirementID = nil
		review.ReviewStatus = "pending"
		review.ReviewedByUserID = nil
		review.ReviewedAt = nil

		if err := s.evidenceRepo.UpdateAIReview(tx, review); err != nil {
			return nil, apperrors.InternalServer("failed to update ai review")
		}
	}

	doc.Status = "classified"
	doc.AiConfidence = param.AIConfidence

	if err := s.evidenceRepo.UpdateEvidenceDocument(tx, doc); err != nil {
		return nil, apperrors.InternalServer("failed to update evidence document")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, apperrors.InternalServer("failed to commit ai review")
	}

	return mapEvidenceAIReviewResponse(*review, doc), nil
}

func (s *EvidenceService) GetEvidenceAIReviews(userID uuid.UUID, status string) ([]model.EvidenceAIReviewResponse, error) {
	profile, err := s.getProfileByUserID(s.db, userID)
	if err != nil {
		return nil, err
	}

	status = strings.TrimSpace(status)
	reviews, err := s.evidenceRepo.GetAIReviewsByProfileID(s.db, profile.ProfileID, status)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get ai reviews")
	}

	evidenceIDs := make([]uuid.UUID, 0, len(reviews))
	for _, review := range reviews {
		evidenceIDs = append(evidenceIDs, review.EvidenceID)
	}

	docs, err := s.evidenceRepo.GetEvidenceDocumentsByIDs(s.db, evidenceIDs)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get evidence documents")
	}

	docByID := make(map[uuid.UUID]*entity.EvidenceDocument, len(docs))
	for _, doc := range docs {
		docByID[doc.EvidenceID] = doc
	}

	responses := make([]model.EvidenceAIReviewResponse, 0, len(reviews))
	for _, review := range reviews {
		if doc, ok := docByID[review.EvidenceID]; ok {
			responses = append(responses, *mapEvidenceAIReviewResponse(*review, doc))
			continue
		}

		responses = append(responses, *mapEvidenceAIReviewResponse(*review, nil))
	}

	return responses, nil
}

func (s *EvidenceService) ReviewEvidenceAI(userID uuid.UUID, param model.ReviewEvidenceAIParam) (*model.EvidenceAIReviewResponse, error) {
	action := strings.ToLower(strings.TrimSpace(param.Action))
	if action != "approve" && action != "correct" && action != "reject" {
		return nil, apperrors.BadRequest("action must be approve, correct, or reject")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	profile, err := s.getProfileByUserID(tx, userID)
	if err != nil {
		return nil, err
	}

	review, err := s.evidenceRepo.GetAIReviewByID(tx, param.ReviewID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("ai review not found")
		}
		return nil, apperrors.InternalServer("failed to get ai review")
	}

	if review.ProfileID != profile.ProfileID {
		return nil, apperrors.Forbidden("ai review does not belong to user")
	}

	doc, err := s.evidenceRepo.GetEvidenceDocumentByID(tx, review.EvidenceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("evidence document not found")
		}
		return nil, apperrors.InternalServer("failed to get evidence document")
	}

	now := time.Now()
	reviewedBy := userID
	review.ReviewedAt = &now
	review.ReviewedByUserID = &reviewedBy

	switch action {
	case "approve":
		if review.SuggestedRequirementID == nil || strings.TrimSpace(*review.SuggestedRequirementID) == "" {
			return nil, apperrors.BadRequest("suggested_requirement_id is required to approve")
		}

		finalRequirementID := strings.TrimSpace(*review.SuggestedRequirementID)
		review.FinalRequirementID = &finalRequirementID
		review.ReviewStatus = "approved"
		doc.RequirementID = &finalRequirementID
		doc.Status = "reviewed"

	case "correct":
		if param.FinalRequirementID == nil || strings.TrimSpace(*param.FinalRequirementID) == "" {
			return nil, apperrors.BadRequest("final_requirement_id is required to correct")
		}

		finalRequirementID := strings.TrimSpace(*param.FinalRequirementID)
		req, err := s.evidenceRepo.GetEvidenceRequirement(tx, finalRequirementID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.BadRequest("final requirement not found")
			}
			return nil, apperrors.InternalServer("failed to get final requirement")
		}

		review.FinalRequirementID = &finalRequirementID
		review.SuggestedCategoryID = req.CategoryID
		review.ReviewStatus = "corrected"
		doc.RequirementID = &finalRequirementID
		doc.Status = "reviewed"

	case "reject":
		review.FinalRequirementID = nil
		review.ReviewStatus = "rejected"
		doc.Status = "rejected"
	}

	if err := s.evidenceRepo.UpdateAIReview(tx, review); err != nil {
		return nil, apperrors.InternalServer("failed to update ai review")
	}

	if err := s.evidenceRepo.UpdateEvidenceDocument(tx, doc); err != nil {
		return nil, apperrors.InternalServer("failed to update evidence document")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, apperrors.InternalServer("failed to commit ai review decision")
	}

	return mapEvidenceAIReviewResponse(*review, doc), nil
}

func mapEvidenceAIReviewResponse(review entity.EvidenceAIReview, doc *entity.EvidenceDocument) *model.EvidenceAIReviewResponse {
	var document *model.EvidenceDocumentItem
	if doc != nil {
		item := mapEvidenceDocumentItem(*doc)
		document = &item
	}

	return &model.EvidenceAIReviewResponse{
		ReviewID:               review.ReviewID,
		EvidenceID:             review.EvidenceID,
		ProfileID:              review.ProfileID,
		AISummary:              review.AISummary,
		AIConfidence:           review.AIConfidence,
		SuggestedCategoryID:    review.SuggestedCategoryID,
		SuggestedRequirementID: review.SuggestedRequirementID,
		FinalRequirementID:     review.FinalRequirementID,
		ReviewStatus:           review.ReviewStatus,
		ReviewedAt:             review.ReviewedAt,
		Document:               document,
		CreatedAt:              review.CreatedAt,
		UpdatedAt:              review.UpdatedAt,
	}
}

func mapEvidenceCategoryResponse(category *entity.EvidenceCategory) model.EvidenceCategoryResponse {
	requirements := make([]model.EvidenceRequirementResponse, 0, len(category.Requirements))
	for _, requirement := range category.Requirements {
		requirements = append(requirements, model.EvidenceRequirementResponse{
			RequirementID: requirement.RequirementID,
			CategoryID:    requirement.CategoryID,
			Name:          requirement.Name,
			Description:   requirement.Description,
			IsRequired:    requirement.IsRequired,
			SortOrder:     requirement.SortOrder,
		})
	}

	return model.EvidenceCategoryResponse{
		CategoryID:   category.CategoryID,
		Code:         category.CategoryID,
		Name:         category.Name,
		Weight:       category.Weight,
		SortOrder:    category.SortOrder,
		Requirements: requirements,
	}
}

func (s *EvidenceService) getProfileByUserID(tx *gorm.DB, userID uuid.UUID) (*entity.UMKMProfile, error) {
	profile, err := s.umkmProfileRepo.GetUMKMProfile(tx, model.GetUMKMProfileParam{
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("business profile not found")
		}
		return nil, apperrors.InternalServer("failed to get business profile")
	}

	return profile, nil
}

func buildCategoryProgress(category *entity.EvidenceCategory) model.EvidenceCategoryProgress {
	requiredCount := 0
	fulfilledCount := 0

	for _, requirement := range category.Requirements {
		if !requirement.IsRequired {
			continue
		}

		requiredCount++

		if hasFulfilledDocument(requirement.EvidenceDocuments) {
			fulfilledCount++
		}
	}

	score := 0.0
	if requiredCount > 0 {
		score = (float64(fulfilledCount) / float64(requiredCount)) * category.Weight
	}

	return model.EvidenceCategoryProgress{
		CategoryID:     category.CategoryID,
		Code:           category.CategoryID,
		Name:           category.Name,
		Weight:         category.Weight,
		RequiredCount:  requiredCount,
		FulfilledCount: fulfilledCount,
		Score:          roundScore(score),
		Status:         buildCategoryStatus(requiredCount, fulfilledCount),
	}
}

func buildCategoryStatus(requiredCount int, fulfilledCount int) string {
	if fulfilledCount == 0 {
		return "empty"
	}

	if fulfilledCount < requiredCount {
		return "partial"
	}

	return "complete"
}

func hasFulfilledDocument(docs []entity.EvidenceDocument) bool {
	for _, doc := range docs {
		if isFulfilledEvidenceStatus(doc.Status) {
			return true
		}
	}

	return false
}

func selectRequirementDocument(docs []entity.EvidenceDocument) *entity.EvidenceDocument {
	if len(docs) == 0 {
		return nil
	}

	for i := range docs {
		if isFulfilledEvidenceStatus(docs[i].Status) {
			return &docs[i]
		}
	}

	return &docs[0]
}

func isFulfilledEvidenceStatus(status string) bool {
	switch status {
	case "reviewed", "on_chain":
		return true
	default:
		return false
	}
}

func hashMultipartFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, src); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func mapEvidenceDocumentItem(doc entity.EvidenceDocument) model.EvidenceDocumentItem {
	return model.EvidenceDocumentItem{
		EvidenceID:    doc.EvidenceID,
		RequirementID: doc.RequirementID,
		FileName:      doc.OriginalName,
		FilePath:      doc.FilePath,
		MimeType:      doc.MimeType,
		FileSize:      doc.FileSize,
		Status:        doc.Status,
		AiConfidence:  doc.AiConfidence,
		CreatedAt:     doc.CreatedAt,
	}
}

func mapEvidenceDocumentResponse(doc entity.EvidenceDocument) *model.EvidenceDocumentResponse {
	return &model.EvidenceDocumentResponse{
		EvidenceID:       doc.EvidenceID,
		ProfileID:        doc.ProfileID,
		RequirementID:    doc.RequirementID,
		FileName:         doc.OriginalName,
		FilePath:         doc.FilePath,
		FileHash:         doc.FileHash,
		MimeType:         doc.MimeType,
		FileSize:         doc.FileSize,
		AiConfidence:     doc.AiConfidence,
		Status:           doc.Status,
		BlockchainTxHash: doc.BlockchainTxHash,
		CreatedAt:        doc.CreatedAt,
		UpdatedAt:        doc.UpdatedAt,
	}
}

func roundScore(score float64) float64 {
	return math.Round(score*100) / 100
}
