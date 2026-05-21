package repository

import (
	"greentrust-hackathon/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IProposalRepository interface {
	CreateProposal(tx *gorm.DB, proposal *entity.Proposal) error
	GetProposalByID(tx *gorm.DB, proposalID uuid.UUID) (*entity.Proposal, error)
	GetProposalByIDAndUserID(tx *gorm.DB, proposalID uuid.UUID, userID uuid.UUID) (*entity.Proposal, error)
	GetProposalsByUserID(tx *gorm.DB, userID uuid.UUID, box string, status string) ([]*entity.Proposal, error)
	UpdateProposal(tx *gorm.DB, proposal *entity.Proposal) error
	DeleteProposal(tx *gorm.DB, proposal *entity.Proposal) error
	CreateProposalAttachments(tx *gorm.DB, attachments []entity.ProposalAttachment) error
	DeleteAttachmentsByProposalID(tx *gorm.DB, proposalID uuid.UUID) error
}

type ProposalRepository struct {
	db *gorm.DB
}

func NewProposalRepository(db *gorm.DB) IProposalRepository {
	return &ProposalRepository{db: db}
}

func (r *ProposalRepository) CreateProposal(tx *gorm.DB, proposal *entity.Proposal) error {
	err := tx.Debug().Create(proposal).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *ProposalRepository) GetProposalByID(tx *gorm.DB, proposalID uuid.UUID) (*entity.Proposal, error) {
	var proposal entity.Proposal
	err := tx.Debug().
		Preload("Sender").
		Preload("Receiver").
		Preload("Attachments").
		Where("proposal_id = ?", proposalID).
		First(&proposal).Error
	if err != nil {
		return nil, err
	}

	return &proposal, nil
}

func (r *ProposalRepository) GetProposalByIDAndUserID(tx *gorm.DB, proposalID uuid.UUID, userID uuid.UUID) (*entity.Proposal, error) {
	var proposal entity.Proposal
	err := tx.Debug().
		Preload("Sender").
		Preload("Receiver").
		Preload("Attachments").
		Where("proposal_id = ? AND (sender_user_id = ? OR receiver_user_id = ?)", proposalID, userID, userID).
		First(&proposal).Error
	if err != nil {
		return nil, err
	}

	return &proposal, nil
}

func (r *ProposalRepository) GetProposalsByUserID(tx *gorm.DB, userID uuid.UUID, box string, status string) ([]*entity.Proposal, error) {
	var proposals []*entity.Proposal

	query := tx.Debug().
		Preload("Sender").
		Preload("Receiver").
		Preload("Attachments").
		Order("created_at DESC")

	switch box {
	case "inbox":
		query = query.Where("receiver_user_id = ?", userID).
			Where("status <> ?", "draft")
	case "sent":
		query = query.Where("sender_user_id = ?", userID)
	default:
		query = query.Where("sender_user_id = ? OR receiver_user_id = ?", userID, userID).
			Where("(status <> ? OR sender_user_id = ?)", "draft", userID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&proposals).Error
	if err != nil {
		return nil, err
	}

	return proposals, nil
}

func (r *ProposalRepository) UpdateProposal(tx *gorm.DB, proposal *entity.Proposal) error {
	err := tx.Debug().Save(proposal).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *ProposalRepository) DeleteProposal(tx *gorm.DB, proposal *entity.Proposal) error {
	err := tx.Debug().Delete(proposal).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *ProposalRepository) CreateProposalAttachments(tx *gorm.DB, attachments []entity.ProposalAttachment) error {
	if len(attachments) == 0 {
		return nil
	}

	err := tx.Debug().Create(&attachments).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *ProposalRepository) DeleteAttachmentsByProposalID(tx *gorm.DB, proposalID uuid.UUID) error {
	err := tx.Debug().
		Where("proposal_id = ?", proposalID).
		Delete(&entity.ProposalAttachment{}).Error
	if err != nil {
		return err
	}

	return nil
}
