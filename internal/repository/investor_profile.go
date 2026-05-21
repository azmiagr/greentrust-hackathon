package repository

import (
	"greentrust-hackathon/entity"
	"greentrust-hackathon/model"
	"strings"

	"gorm.io/gorm"
)

type IInvestorProfileRepository interface {
	GetInvestorProfile(tx *gorm.DB, param model.GetInvestorProfileParam) (*entity.InvestorProfile, error)
	GetInvestorProfilesByIDs(tx *gorm.DB, profileIDs []string) ([]*entity.InvestorProfile, error)
	CreateInvestorProfile(tx *gorm.DB, profile *entity.InvestorProfile) error
	UpdateInvestorProfile(tx *gorm.DB, profile *entity.InvestorProfile) error
	ReplaceInvestorFocusSectors(tx *gorm.DB, profile *entity.InvestorProfile, sectors []entity.BusinessSector) error
	GetPublicInvestorDirectoryItems(tx *gorm.DB, param model.PublicInvestorDirectoryFiltersParam) ([]model.PublicInvestorDirectoryRow, error)
	CountPublicInvestorDirectoryItems(tx *gorm.DB, param model.PublicInvestorDirectoryFiltersParam) (int64, error)
	GetPublicInvestorTypeFilters(tx *gorm.DB) ([]model.PublicInvestorTypeFilterRow, error)
	GetPublicInvestorSectorFilters(tx *gorm.DB) ([]model.PublicInvestorSectorFilterRow, error)
	GetPublicInvestorTicketRangeFilters(tx *gorm.DB) ([]model.PublicInvestorTicketRangeFilterRow, error)
	GetPublicInvestorDetail(tx *gorm.DB, profileID string) (*model.PublicInvestorDetailRow, error)
}

type InvestorProfileRepository struct {
	db *gorm.DB
}

func NewInvestorProfileRepository(db *gorm.DB) IInvestorProfileRepository {
	return &InvestorProfileRepository{db: db}
}

func (r *InvestorProfileRepository) GetInvestorProfile(tx *gorm.DB, param model.GetInvestorProfileParam) (*entity.InvestorProfile, error) {
	var profile entity.InvestorProfile
	err := tx.Debug().
		Preload("FocusSectors").
		Where(&param).
		First(&profile).Error
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *InvestorProfileRepository) GetInvestorProfilesByIDs(tx *gorm.DB, profileIDs []string) ([]*entity.InvestorProfile, error) {
	var profiles []*entity.InvestorProfile
	if len(profileIDs) == 0 {
		return profiles, nil
	}

	err := tx.Debug().
		Preload("FocusSectors").
		Where("profile_id IN ?", profileIDs).
		Find(&profiles).Error
	if err != nil {
		return nil, err
	}

	return profiles, nil
}

func (r *InvestorProfileRepository) CreateInvestorProfile(tx *gorm.DB, profile *entity.InvestorProfile) error {
	err := tx.Debug().Create(profile).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *InvestorProfileRepository) UpdateInvestorProfile(tx *gorm.DB, profile *entity.InvestorProfile) error {
	err := tx.Debug().Save(profile).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *InvestorProfileRepository) ReplaceInvestorFocusSectors(tx *gorm.DB, profile *entity.InvestorProfile, sectors []entity.BusinessSector) error {
	err := tx.Debug().Model(profile).Association("FocusSectors").Replace(sectors)
	if err != nil {
		return err
	}

	return nil
}

func (r *InvestorProfileRepository) GetPublicInvestorDirectoryItems(tx *gorm.DB, param model.PublicInvestorDirectoryFiltersParam) ([]model.PublicInvestorDirectoryRow, error) {
	var rows []model.PublicInvestorDirectoryRow
	query := applyPublicInvestorDirectoryFilters(buildPublicInvestorDirectoryBaseQuery(tx), param)

	switch param.Sort {
	case "approval_rate_desc":
		query = query.Order("approval_rate_sort DESC")
	case "portfolio_desc":
		query = query.Order("portfolio_count DESC")
	case "name_asc":
		query = query.Order("full_name_sort ASC")
	default:
		query = query.Order("investor_profiles.updated_at DESC")
	}

	err := query.
		Limit(param.Limit).
		Offset(param.Offset).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *InvestorProfileRepository) CountPublicInvestorDirectoryItems(tx *gorm.DB, param model.PublicInvestorDirectoryFiltersParam) (int64, error) {
	var total int64
	query := applyPublicInvestorDirectoryFilters(buildPublicInvestorDirectoryCountQuery(tx), param)
	err := query.Count(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *InvestorProfileRepository) GetPublicInvestorTypeFilters(tx *gorm.DB) ([]model.PublicInvestorTypeFilterRow, error) {
	var rows []model.PublicInvestorTypeFilterRow
	err := tx.Debug().
		Table("investor_profiles").
		Select("investor_profiles.investor_type, COUNT(*) AS count").
		Where("investor_profiles.investor_type <> ''").
		Group("investor_profiles.investor_type").
		Order("investor_profiles.investor_type ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *InvestorProfileRepository) GetPublicInvestorSectorFilters(tx *gorm.DB) ([]model.PublicInvestorSectorFilterRow, error) {
	var rows []model.PublicInvestorSectorFilterRow
	err := tx.Debug().
		Table("investor_profile_focus_sectors").
		Select("business_sectors.sector_id, business_sectors.sector_name, COUNT(DISTINCT investor_profile_focus_sectors.profile_id) AS count").
		Joins("JOIN business_sectors ON business_sectors.sector_id = investor_profile_focus_sectors.sector_id").
		Group("business_sectors.sector_id, business_sectors.sector_name").
		Order("business_sectors.sector_name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *InvestorProfileRepository) GetPublicInvestorTicketRangeFilters(tx *gorm.DB) ([]model.PublicInvestorTicketRangeFilterRow, error) {
	var rows []model.PublicInvestorTicketRangeFilterRow
	err := tx.Debug().
		Table("investor_profiles").
		Select("investor_profiles.ticket_range, COUNT(*) AS count").
		Where("investor_profiles.ticket_range <> ''").
		Group("investor_profiles.ticket_range").
		Order("investor_profiles.ticket_range ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *InvestorProfileRepository) GetPublicInvestorDetail(tx *gorm.DB, profileID string) (*model.PublicInvestorDetailRow, error) {
	var row model.PublicInvestorDetailRow
	err := buildPublicInvestorDetailBaseQuery(tx).
		Where("investor_profiles.profile_id = ?", profileID).
		First(&row).Error
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func buildPublicInvestorDirectoryBaseQuery(tx *gorm.DB) *gorm.DB {
	primaryPositions := tx.
		Table("investor_positions").
		Select(`
			investor_positions.profile_id,
			investor_positions.title,
			investor_positions.institution_name,
			ROW_NUMBER() OVER (
				PARTITION BY investor_positions.profile_id
				ORDER BY investor_positions.is_current DESC, investor_positions.start_date DESC
			) AS row_num
		`)

	proposalStats := tx.
		Table("proposals").
		Select(`
			CASE
				WHEN proposals.sender_role = 'investor' THEN proposals.sender_profile_id
				WHEN proposals.receiver_role = 'investor' THEN proposals.receiver_profile_id
			END AS profile_id,
			COUNT(DISTINCT CASE
				WHEN proposals.status = 'accepted' AND proposals.sender_role = 'umkm' THEN proposals.sender_profile_id
				WHEN proposals.status = 'accepted' AND proposals.receiver_role = 'umkm' THEN proposals.receiver_profile_id
			END) AS portfolio_count,
			SUM(CASE WHEN proposals.status = 'accepted' THEN 1 ELSE 0 END) AS accepted_count,
			SUM(CASE WHEN proposals.status = 'rejected' THEN 1 ELSE 0 END) AS rejected_count
		`).
		Where("proposals.sender_role = ? OR proposals.receiver_role = ?", "investor", "investor").
		Group("profile_id")

	return tx.Debug().
		Table("investor_profiles").
		Select(`
			investor_profiles.profile_id,
			investor_profiles.user_id,
			investor_profiles.investor_type,
			investor_profiles.ticket_range,
			user_identities.first_name,
			user_identities.last_name,
			users.email,
			primary_positions.title,
			primary_positions.institution_name,
			COALESCE(proposal_stats.portfolio_count, 0) AS portfolio_count,
			COALESCE(proposal_stats.accepted_count, 0) AS accepted_count,
			COALESCE(proposal_stats.rejected_count, 0) AS rejected_count,
			CONCAT(COALESCE(user_identities.first_name, ''), ' ', COALESCE(user_identities.last_name, '')) AS full_name_sort,
			CASE
				WHEN (COALESCE(proposal_stats.accepted_count, 0) + COALESCE(proposal_stats.rejected_count, 0)) = 0 THEN 0
				ELSE (COALESCE(proposal_stats.accepted_count, 0) / (COALESCE(proposal_stats.accepted_count, 0) + COALESCE(proposal_stats.rejected_count, 0)))
			END AS approval_rate_sort
		`).
		Joins("JOIN users ON users.user_id = investor_profiles.user_id").
		Joins("LEFT JOIN user_identities ON user_identities.user_id = investor_profiles.user_id").
		Joins("LEFT JOIN (?) primary_positions ON primary_positions.profile_id = investor_profiles.profile_id AND primary_positions.row_num = 1", primaryPositions).
		Joins("LEFT JOIN (?) proposal_stats ON proposal_stats.profile_id = investor_profiles.profile_id", proposalStats)
}

func buildPublicInvestorDetailBaseQuery(tx *gorm.DB) *gorm.DB {
	primaryPositions := tx.
		Table("investor_positions").
		Select(`
			investor_positions.profile_id,
			investor_positions.title,
			investor_positions.institution_name,
			investor_positions.location,
			investor_positions.description,
			ROW_NUMBER() OVER (
				PARTITION BY investor_positions.profile_id
				ORDER BY investor_positions.is_current DESC, investor_positions.start_date DESC
			) AS row_num
		`)

	proposalStats := tx.
		Table("proposals").
		Select(`
			CASE
				WHEN proposals.sender_role = 'investor' THEN proposals.sender_profile_id
				WHEN proposals.receiver_role = 'investor' THEN proposals.receiver_profile_id
			END AS profile_id,
			COUNT(DISTINCT CASE
				WHEN proposals.status = 'accepted' AND proposals.sender_role = 'umkm' THEN proposals.sender_profile_id
				WHEN proposals.status = 'accepted' AND proposals.receiver_role = 'umkm' THEN proposals.receiver_profile_id
			END) AS portfolio_count,
			SUM(CASE WHEN proposals.status = 'accepted' THEN 1 ELSE 0 END) AS accepted_count,
			SUM(CASE WHEN proposals.status = 'rejected' THEN 1 ELSE 0 END) AS rejected_count
		`).
		Where("proposals.sender_role = ? OR proposals.receiver_role = ?", "investor", "investor").
		Group("profile_id")

	return tx.Debug().
		Table("investor_profiles").
		Select(`
			investor_profiles.profile_id,
			investor_profiles.user_id,
			investor_profiles.investor_type,
			investor_profiles.ticket_range,
			user_identities.first_name,
			user_identities.last_name,
			users.email,
			primary_positions.title,
			primary_positions.institution_name,
			primary_positions.location AS base_location,
			primary_positions.description AS public_bio,
			CASE WHEN user_identities.identity_id IS NULL THEN false ELSE true END AS is_verified,
			COALESCE(proposal_stats.portfolio_count, 0) AS portfolio_count,
			COALESCE(proposal_stats.accepted_count, 0) AS accepted_count,
			COALESCE(proposal_stats.rejected_count, 0) AS rejected_count
		`).
		Joins("JOIN users ON users.user_id = investor_profiles.user_id").
		Joins("LEFT JOIN user_identities ON user_identities.user_id = investor_profiles.user_id").
		Joins("LEFT JOIN (?) primary_positions ON primary_positions.profile_id = investor_profiles.profile_id AND primary_positions.row_num = 1", primaryPositions).
		Joins("LEFT JOIN (?) proposal_stats ON proposal_stats.profile_id = investor_profiles.profile_id", proposalStats)
}

func buildPublicInvestorDirectoryCountQuery(tx *gorm.DB) *gorm.DB {
	return tx.Debug().
		Table("investor_profiles").
		Joins("JOIN users ON users.user_id = investor_profiles.user_id").
		Joins("LEFT JOIN user_identities ON user_identities.user_id = investor_profiles.user_id")
}

func applyPublicInvestorDirectoryFilters(query *gorm.DB, param model.PublicInvestorDirectoryFiltersParam) *gorm.DB {
	if param.Search != "" {
		search := "%" + param.Search + "%"
		query = query.Where(`
			user_identities.first_name LIKE ?
			OR user_identities.last_name LIKE ?
			OR users.email LIKE ?
			OR investor_profiles.investor_type LIKE ?
			OR investor_profiles.ticket_range LIKE ?
		`, search, search, search, search, search)
	}
	if len(param.InvestorTypes) > 0 {
		query = query.Where("investor_profiles.investor_type IN ?", param.InvestorTypes)
	}
	if len(param.TicketRanges) > 0 {
		query = query.Where("investor_profiles.ticket_range IN ?", param.TicketRanges)
	}
	if len(param.SectorIDs) > 0 {
		query = query.Where(`
			EXISTS (
				SELECT 1
				FROM investor_profile_focus_sectors ipfs
				WHERE ipfs.profile_id = investor_profiles.profile_id
				AND ipfs.sector_id IN ?
			)
		`, param.SectorIDs)
	}

	return query
}

func normalizePublicInvestorFilterValues(input []string) []string {
	values := make([]string, 0, len(input))
	seen := map[string]bool{}
	for _, item := range input {
		value := strings.TrimSpace(item)
		key := strings.ToLower(value)
		if value == "" || seen[key] {
			continue
		}
		seen[key] = true
		values = append(values, value)
	}

	return values
}
