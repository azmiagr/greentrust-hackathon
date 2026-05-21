package repository

import (
	"greentrust-hackathon/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IEvidenceRepository interface {
	GetEvidenceCategories(tx *gorm.DB) ([]*entity.EvidenceCategory, error)
	GetEvidenceCategoriesWithDocuments(tx *gorm.DB, profileID uuid.UUID) ([]*entity.EvidenceCategory, error)
	GetEvidenceCategoryWithDocuments(tx *gorm.DB, categoryID string, profileID uuid.UUID) (*entity.EvidenceCategory, error)
	GetEvidenceRequirement(tx *gorm.DB, requirementID string) (*entity.EvidenceRequirement, error)
	GetEvidenceDocumentsByProfileID(tx *gorm.DB, profileID uuid.UUID) ([]*entity.EvidenceDocument, error)
	CreateEvidenceDocument(tx *gorm.DB, doc *entity.EvidenceDocument) error
	GetEvidenceDocumentByID(tx *gorm.DB, evidenceID uuid.UUID) (*entity.EvidenceDocument, error)
	UpdateEvidenceDocument(tx *gorm.DB, doc *entity.EvidenceDocument) error
	GetAIReviewsByProfileID(tx *gorm.DB, profileID uuid.UUID, status string) ([]*entity.EvidenceAIReview, error)
	GetAIReviewByID(tx *gorm.DB, reviewID uuid.UUID) (*entity.EvidenceAIReview, error)
	GetAIReviewByEvidenceID(tx *gorm.DB, evidenceID uuid.UUID) (*entity.EvidenceAIReview, error)
	CreateAIReview(tx *gorm.DB, review *entity.EvidenceAIReview) error
	UpdateAIReview(tx *gorm.DB, review *entity.EvidenceAIReview) error
	GetEvidenceDocumentsByIDs(tx *gorm.DB, evidenceIDs []uuid.UUID) ([]*entity.EvidenceDocument, error)
	GetReviewedEvidenceDocumentsByProfileID(tx *gorm.DB, profileID uuid.UUID) ([]*entity.EvidenceDocument, error)
	MarkDocumentsOnChain(tx *gorm.DB, profileID uuid.UUID, txHash string) error
	MarkDocumentsBlockchainFailed(tx *gorm.DB, profileID uuid.UUID) error
}

type EvidenceRepository struct {
	db *gorm.DB
}

func NewEvidenceRepository(db *gorm.DB) IEvidenceRepository {
	return &EvidenceRepository{db: db}
}

func (r *EvidenceRepository) GetEvidenceCategories(tx *gorm.DB) ([]*entity.EvidenceCategory, error) {
	var categories []*entity.EvidenceCategory
	err := tx.Debug().
		Preload("Requirements", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Order("sort_order ASC").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *EvidenceRepository) GetEvidenceCategoriesWithDocuments(tx *gorm.DB, profileID uuid.UUID) ([]*entity.EvidenceCategory, error) {
	var categories []*entity.EvidenceCategory
	err := tx.Debug().
		Preload("Requirements", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Requirements.EvidenceDocuments", func(db *gorm.DB) *gorm.DB {
			return db.Where("profile_id = ?", profileID).Order("created_at DESC")
		}).
		Order("sort_order ASC").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *EvidenceRepository) GetEvidenceCategoryWithDocuments(tx *gorm.DB, categoryID string, profileID uuid.UUID) (*entity.EvidenceCategory, error) {
	var category entity.EvidenceCategory
	err := tx.Debug().
		Preload("Requirements", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Requirements.EvidenceDocuments", func(db *gorm.DB) *gorm.DB {
			return db.Where("profile_id = ?", profileID).Order("created_at DESC")
		}).
		Where("category_id = ?", categoryID).
		First(&category).Error
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *EvidenceRepository) GetEvidenceRequirement(tx *gorm.DB, requirementID string) (*entity.EvidenceRequirement, error) {
	var requirement entity.EvidenceRequirement
	err := tx.Debug().
		Where("requirement_id = ?", requirementID).
		First(&requirement).Error
	if err != nil {
		return nil, err
	}

	return &requirement, nil
}

func (r *EvidenceRepository) GetEvidenceDocumentsByProfileID(tx *gorm.DB, profileID uuid.UUID) ([]*entity.EvidenceDocument, error) {
	var docs []*entity.EvidenceDocument
	err := tx.Debug().
		Where("profile_id = ?", profileID).
		Order("created_at DESC").
		Find(&docs).Error
	if err != nil {
		return nil, err
	}
	return docs, err
}

func (r *EvidenceRepository) CreateEvidenceDocument(tx *gorm.DB, doc *entity.EvidenceDocument) error {
	err := tx.Debug().Create(doc).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *EvidenceRepository) GetEvidenceDocumentByID(tx *gorm.DB, evidenceID uuid.UUID) (*entity.EvidenceDocument, error) {
	var doc entity.EvidenceDocument
	err := tx.Debug().
		Where("evidence_id = ?", evidenceID).
		First(&doc).Error
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *EvidenceRepository) GetEvidenceDocumentsByIDs(tx *gorm.DB, evidenceIDs []uuid.UUID) ([]*entity.EvidenceDocument, error) {
	var docs []*entity.EvidenceDocument
	if len(evidenceIDs) == 0 {
		return docs, nil
	}

	err := tx.Debug().
		Where("evidence_id IN ?", evidenceIDs).
		Find(&docs).Error
	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (r *EvidenceRepository) UpdateEvidenceDocument(tx *gorm.DB, doc *entity.EvidenceDocument) error {
	err := tx.Debug().Save(doc).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *EvidenceRepository) GetAIReviewsByProfileID(tx *gorm.DB, profileID uuid.UUID, status string) ([]*entity.EvidenceAIReview, error) {
	var reviews []*entity.EvidenceAIReview

	query := tx.Debug().
		Where("profile_id = ?", profileID).
		Order("created_at DESC")

	if status != "" {
		query = query.Where("review_status = ?", status)
	}

	err := query.Find(&reviews).Error
	if err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *EvidenceRepository) GetAIReviewByID(tx *gorm.DB, reviewID uuid.UUID) (*entity.EvidenceAIReview, error) {
	var review entity.EvidenceAIReview
	err := tx.Debug().
		Where("review_id = ?", reviewID).
		First(&review).Error
	if err != nil {
		return nil, err
	}

	return &review, nil
}

func (r *EvidenceRepository) GetAIReviewByEvidenceID(tx *gorm.DB, evidenceID uuid.UUID) (*entity.EvidenceAIReview, error) {
	var review entity.EvidenceAIReview
	err := tx.Debug().
		Where("evidence_id = ?", evidenceID).
		First(&review).Error
	if err != nil {
		return nil, err
	}

	return &review, nil
}

func (r *EvidenceRepository) CreateAIReview(tx *gorm.DB, review *entity.EvidenceAIReview) error {
	err := tx.Debug().Create(review).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *EvidenceRepository) UpdateAIReview(tx *gorm.DB, review *entity.EvidenceAIReview) error {
	err := tx.Debug().Save(review).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *EvidenceRepository) GetReviewedEvidenceDocumentsByProfileID(tx *gorm.DB, profileID uuid.UUID) ([]*entity.EvidenceDocument, error) {
	var docs []*entity.EvidenceDocument
	err := tx.Where("profile_id = ? AND status IN ?", profileID, []string{"reviewed", "on_chain"}).
		Order("created_at ASC").Find(&docs).Error

	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *EvidenceRepository) MarkDocumentsOnChain(tx *gorm.DB, profileID uuid.UUID, txHash string) error {
	err := tx.Model(&entity.EvidenceDocument{}).
		Where("profile_id = ? AND status = ?", profileID, "reviewed").
		Updates(map[string]any{"status": "on_chain", "blockchain_tx_hash": txHash}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *EvidenceRepository) MarkDocumentsBlockchainFailed(tx *gorm.DB, profileID uuid.UUID) error {
	err := tx.Model(&entity.EvidenceDocument{}).
		Where("profile_id = ? AND status = ?", profileID, "reviewed").
		Update("status", "blockchain_failed").Error
	if err != nil {
		return err
	}
	return nil
}
