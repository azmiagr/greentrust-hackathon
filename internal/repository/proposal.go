package repository

import (
	"greentrust-hackathon/entity"
	"greentrust-hackathon/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IProposalRepository interface {
	CreateProposal(tx *gorm.DB, proposal *entity.Proposal) error
	GetProposalByID(tx *gorm.DB, proposalID uuid.UUID) (*entity.Proposal, error)
	GetProposalByIDAndUserID(tx *gorm.DB, proposalID uuid.UUID, userID uuid.UUID) (*entity.Proposal, error)
	GetProposalsByUserID(tx *gorm.DB, userID uuid.UUID, box string, status string) ([]*entity.Proposal, error)
	GetInvestorProposalDashboardStats(tx *gorm.DB, userID uuid.UUID) (*model.InvestorProposalDashboardStats, error)
	GetRecentInvestorDashboardProposals(tx *gorm.DB, userID uuid.UUID, limit int) ([]*entity.Proposal, error)
	GetInvestorPortfolio(tx *gorm.DB, userID uuid.UUID, query model.InvestorPortfolioQuery) ([]model.InvestorPortfolioRow, error)
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

func (r *ProposalRepository) GetInvestorProposalDashboardStats(tx *gorm.DB, userID uuid.UUID) (*model.InvestorProposalDashboardStats, error) {
	var stats model.InvestorProposalDashboardStats
	err := tx.Debug().
		Model(&entity.Proposal{}).
		Select(`
			COALESCE(SUM(CASE WHEN status IN ('sent', 'accepted') THEN 1 ELSE 0 END), 0) AS active_proposals_count,
			COALESCE(SUM(CASE WHEN status = 'accepted' THEN 1 ELSE 0 END), 0) AS accepted_proposals_count,
			COALESCE(SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END), 0) AS waiting_confirmation_count
		`).
		Where("(sender_user_id = ? OR receiver_user_id = ?) AND status <> ?", userID, userID, "draft").
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *ProposalRepository) GetRecentInvestorDashboardProposals(tx *gorm.DB, userID uuid.UUID, limit int) ([]*entity.Proposal, error) {
	var proposals []*entity.Proposal
	err := tx.Debug().
		Preload("Sender").
		Preload("Receiver").
		Where("(sender_user_id = ? OR receiver_user_id = ?) AND status IN ?", userID, userID, []string{"sent", "accepted", "rejected", "withdrawn"}).
		Order("updated_at DESC").
		Limit(limit).
		Find(&proposals).Error
	if err != nil {
		return nil, err
	}

	return proposals, nil
}

func (r *ProposalRepository) GetInvestorPortfolio(tx *gorm.DB, userID uuid.UUID, query model.InvestorPortfolioQuery) ([]model.InvestorPortfolioRow, error) {
	var rows []model.InvestorPortfolioRow

	portfolioQuery := tx.Debug().
		Table("proposals").
		Select(`
			umkm_profiles.profile_id,
			umkm_profiles.business_name,
			business_sectors.sector_name,
			umkm_profiles.business_city AS city,
			SUM(proposals.amount) AS total_value,
			COALESCE(green_passports.grs_score, 0) AS current_grs,
			MIN(COALESCE(proposals.accepted_at, proposals.updated_at)) AS funded_since,
			COUNT(*) AS accepted_proposal_count,
			SUBSTRING_INDEX(GROUP_CONCAT(proposals.proposal_id ORDER BY COALESCE(proposals.accepted_at, proposals.updated_at) DESC), ',', 1) AS latest_proposal_id,
			SUBSTRING_INDEX(GROUP_CONCAT(proposals.title ORDER BY COALESCE(proposals.accepted_at, proposals.updated_at) DESC SEPARATOR '\n'), '\n', 1) AS latest_proposal_title,
			MAX(COALESCE(proposals.accepted_at, proposals.updated_at)) AS latest_proposal_accepted
		`).
		Joins(`
			JOIN umkm_profiles ON umkm_profiles.profile_id = CASE
				WHEN proposals.sender_user_id = ? AND proposals.receiver_role = 'umkm' THEN proposals.receiver_profile_id
				WHEN proposals.receiver_user_id = ? AND proposals.sender_role = 'umkm' THEN proposals.sender_profile_id
			END
		`, userID, userID).
		Joins("JOIN business_sectors ON business_sectors.sector_id = umkm_profiles.sector_id").
		Joins("LEFT JOIN green_passports ON green_passports.profile_id = umkm_profiles.profile_id").
		Where("(proposals.sender_user_id = ? OR proposals.receiver_user_id = ?)", userID, userID).
		Where("proposals.status = ?", "accepted").
		Where("(proposals.sender_role = ? OR proposals.receiver_role = ?)", "umkm", "umkm").
		Group("umkm_profiles.profile_id, umkm_profiles.business_name, business_sectors.sector_name, umkm_profiles.business_city, green_passports.grs_score").
		Order("latest_proposal_accepted DESC")

	if query.SectorID != "" {
		portfolioQuery = portfolioQuery.Where("umkm_profiles.sector_id = ?", query.SectorID)
	}
	if query.MinGRS > 0 {
		portfolioQuery = portfolioQuery.Where("COALESCE(green_passports.grs_score, 0) >= ?", query.MinGRS)
	}

	err := portfolioQuery.Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
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
