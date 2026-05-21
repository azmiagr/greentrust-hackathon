package repository

import (
	"greentrust-hackathon/entity"
	"greentrust-hackathon/model"
	"strings"
	"time"

	"gorm.io/gorm"
)

type IUMKMProfileRepository interface {
	GetUMKMProfile(tx *gorm.DB, param model.GetUMKMProfileParam) (*entity.UMKMProfile, error)
	CreateUMKMProfile(tx *gorm.DB, profile *entity.UMKMProfile) error
	UpdateUMKMProfile(tx *gorm.DB, profile *entity.UMKMProfile) error
	GetInvestorWatchedUMKMSummary(tx *gorm.DB, investorUserID string, since time.Time) (*model.InvestorWatchedUMKMSummary, error)
	GetRecommendedUMKMsForInvestor(tx *gorm.DB, sectorIDs []string, limit int) ([]model.InvestorDashboardUMKMRow, error)
	GetPublicUMKMDirectoryItems(tx *gorm.DB, param model.PublicUMKMDirectoryFiltersParam) ([]model.PublicUMKMDirectoryRow, error)
	GetPublicUMKMDirectoryItemByProfileID(tx *gorm.DB, profileID string) (*model.PublicUMKMDirectoryRow, error)
	CountPublicUMKMDirectoryItems(tx *gorm.DB, param model.PublicUMKMDirectoryFiltersParam) (int64, error)
	GetPublicUMKMDirectorySectorFilters(tx *gorm.DB) ([]model.PublicUMKMSectorFilterRow, error)
	GetPublicUMKMDirectoryProvinceFilters(tx *gorm.DB) ([]model.PublicUMKMProvinceFilterRow, error)
	GetPublicUMKMDirectoryTierFilters(tx *gorm.DB) ([]model.PublicUMKMTierFilterRow, error)
}

type UMKMProfileRepository struct {
	db *gorm.DB
}

func NewUMKMProfileRepository(db *gorm.DB) IUMKMProfileRepository {
	return &UMKMProfileRepository{db: db}
}

func (r *UMKMProfileRepository) GetUMKMProfile(tx *gorm.DB, param model.GetUMKMProfileParam) (*entity.UMKMProfile, error) {
	var profile entity.UMKMProfile
	err := tx.Debug().Where(&param).First(&profile).Error
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *UMKMProfileRepository) CreateUMKMProfile(tx *gorm.DB, profile *entity.UMKMProfile) error {
	err := tx.Debug().Create(profile).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UMKMProfileRepository) UpdateUMKMProfile(tx *gorm.DB, profile *entity.UMKMProfile) error {
	err := tx.Debug().Save(profile).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UMKMProfileRepository) GetInvestorWatchedUMKMSummary(tx *gorm.DB, investorUserID string, since time.Time) (*model.InvestorWatchedUMKMSummary, error) {
	var summary model.InvestorWatchedUMKMSummary
	watchedProfiles := tx.
		Table("proposals").
		Select(`
			DISTINCT CASE
				WHEN sender_user_id = ? AND receiver_role = 'umkm' THEN receiver_profile_id
				WHEN receiver_user_id = ? AND sender_role = 'umkm' THEN sender_profile_id
			END AS profile_id
		`, investorUserID, investorUserID).
		Where("(sender_user_id = ? OR receiver_user_id = ?) AND status <> ?", investorUserID, investorUserID, "draft")

	err := tx.Debug().
		Table("(?) watched", watchedProfiles).
		Select(`
			COUNT(watched.profile_id) AS watched_umkm_count,
			COALESCE(SUM(CASE WHEN umkm_profiles.created_at >= ? THEN 1 ELSE 0 END), 0) AS watched_umkm_growth_this_week,
			COALESCE(AVG(green_passports.grs_score), 0) AS average_portfolio_grs,
			COALESCE(SUM(CASE WHEN green_passports.grs_score >= 70 THEN 1 ELSE 0 END), 0) AS ready_or_unggul_count,
			COALESCE(SUM(CASE WHEN green_passports.grs_score >= 85 AND green_passports.last_updated_at >= ? THEN 1 ELSE 0 END), 0) AS new_unggul_umkm_this_week
		`, since, since).
		Joins("JOIN umkm_profiles ON umkm_profiles.profile_id = watched.profile_id").
		Joins("LEFT JOIN green_passports ON green_passports.profile_id = watched.profile_id").
		Where("watched.profile_id IS NOT NULL").
		Scan(&summary).Error
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

func (r *UMKMProfileRepository) GetRecommendedUMKMsForInvestor(tx *gorm.DB, sectorIDs []string, limit int) ([]model.InvestorDashboardUMKMRow, error) {
	var rows []model.InvestorDashboardUMKMRow
	if len(sectorIDs) == 0 {
		return rows, nil
	}

	onChainDocCounts := tx.
		Table("evidence_documents").
		Select("profile_id, COUNT(*) AS on_chain_document_count").
		Where("status = ?", "on_chain").
		Group("profile_id")

	err := tx.Debug().
		Table("umkm_profiles").
		Select(`
			umkm_profiles.profile_id,
			umkm_profiles.business_name,
			business_sectors.sector_name,
			umkm_profiles.business_city AS city,
			COALESCE(green_passports.grs_score, 0) AS grs_score,
			COALESCE(on_chain_docs.on_chain_document_count, 0) AS on_chain_document_count
		`).
		Joins("JOIN business_sectors ON business_sectors.sector_id = umkm_profiles.sector_id").
		Joins("LEFT JOIN green_passports ON green_passports.profile_id = umkm_profiles.profile_id").
		Joins("LEFT JOIN (?) on_chain_docs ON on_chain_docs.profile_id = umkm_profiles.profile_id", onChainDocCounts).
		Where("umkm_profiles.sector_id IN ?", sectorIDs).
		Order("COALESCE(green_passports.grs_score, 0) DESC, umkm_profiles.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *UMKMProfileRepository) GetPublicUMKMDirectoryItems(tx *gorm.DB, param model.PublicUMKMDirectoryFiltersParam) ([]model.PublicUMKMDirectoryRow, error) {
	var rows []model.PublicUMKMDirectoryRow
	query := applyPublicUMKMDirectoryFilters(buildPublicUMKMDirectoryBaseQuery(tx), param)

	switch param.Sort {
	case "grs_asc":
		query = query.Order("green_passports.grs_score ASC")
	case "newest":
		query = query.Order("green_passports.issued_at DESC")
	case "name_asc":
		query = query.Order("umkm_profiles.business_name ASC")
	default:
		query = query.Order("green_passports.grs_score DESC")
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

func (r *UMKMProfileRepository) GetPublicUMKMDirectoryItemByProfileID(tx *gorm.DB, profileID string) (*model.PublicUMKMDirectoryRow, error) {
	var row model.PublicUMKMDirectoryRow
	err := buildPublicUMKMDirectoryBaseQuery(tx).
		Where("umkm_profiles.profile_id = ?", profileID).
		First(&row).Error
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (r *UMKMProfileRepository) CountPublicUMKMDirectoryItems(tx *gorm.DB, param model.PublicUMKMDirectoryFiltersParam) (int64, error) {
	var total int64
	query := applyPublicUMKMDirectoryFilters(buildPublicUMKMDirectoryCountQuery(tx), param)
	err := query.Count(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *UMKMProfileRepository) GetPublicUMKMDirectorySectorFilters(tx *gorm.DB) ([]model.PublicUMKMSectorFilterRow, error) {
	var rows []model.PublicUMKMSectorFilterRow
	err := tx.Debug().
		Table("umkm_profiles").
		Select("business_sectors.sector_id, business_sectors.sector_name, COUNT(*) AS count").
		Joins("JOIN business_sectors ON business_sectors.sector_id = umkm_profiles.sector_id").
		Joins("JOIN green_passports ON green_passports.profile_id = umkm_profiles.profile_id AND green_passports.status = ?", "active").
		Group("business_sectors.sector_id, business_sectors.sector_name").
		Order("business_sectors.sector_name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *UMKMProfileRepository) GetPublicUMKMDirectoryProvinceFilters(tx *gorm.DB) ([]model.PublicUMKMProvinceFilterRow, error) {
	var rows []model.PublicUMKMProvinceFilterRow
	err := tx.Debug().
		Table("umkm_profiles").
		Select("umkm_profiles.business_province AS name, COUNT(*) AS count").
		Joins("JOIN green_passports ON green_passports.profile_id = umkm_profiles.profile_id AND green_passports.status = ?", "active").
		Group("umkm_profiles.business_province").
		Order("umkm_profiles.business_province ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *UMKMProfileRepository) GetPublicUMKMDirectoryTierFilters(tx *gorm.DB) ([]model.PublicUMKMTierFilterRow, error) {
	var rows []model.PublicUMKMTierFilterRow
	err := tx.Debug().
		Table("green_passports").
		Select(`
			CASE
				WHEN green_passports.grs_score >= 85 THEN 'unggul'
				WHEN green_passports.grs_score >= 70 THEN 'siap'
				ELSE 'hampir'
			END AS tier,
			COUNT(*) AS count
		`).
		Where("green_passports.status = ?", "active").
		Group("tier").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func buildPublicUMKMDirectoryBaseQuery(tx *gorm.DB) *gorm.DB {
	onChainDocCounts := tx.
		Table("evidence_documents").
		Select("profile_id, COUNT(*) AS on_chain_document_count").
		Where("status = ?", "on_chain").
		Group("profile_id")

	firstPhotos := tx.
		Table("location_photos").
		Select("profile_id, MIN(url) AS photo_url").
		Group("profile_id")

	return tx.Debug().
		Table("umkm_profiles").
		Select(`
			umkm_profiles.profile_id,
			umkm_profiles.business_name,
			business_sectors.sector_name,
			umkm_profiles.business_province AS province,
			umkm_profiles.business_city AS city,
			umkm_profiles.business_description AS description,
			COALESCE(first_photos.photo_url, '') AS photo_url,
			green_passports.passport_id,
			green_passports.grs_score,
			green_passports.status AS passport_status,
			green_passports.public_slug,
			green_passports.blockchain_tx_hash,
			green_passports.issued_at,
			COALESCE(on_chain_docs.on_chain_document_count, 0) AS on_chain_document_count
		`).
		Joins("JOIN business_sectors ON business_sectors.sector_id = umkm_profiles.sector_id").
		Joins("JOIN green_passports ON green_passports.profile_id = umkm_profiles.profile_id AND green_passports.status = ?", "active").
		Joins("LEFT JOIN (?) first_photos ON first_photos.profile_id = umkm_profiles.profile_id", firstPhotos).
		Joins("LEFT JOIN (?) on_chain_docs ON on_chain_docs.profile_id = umkm_profiles.profile_id", onChainDocCounts)
}

func buildPublicUMKMDirectoryCountQuery(tx *gorm.DB) *gorm.DB {
	return tx.Debug().
		Table("umkm_profiles").
		Joins("JOIN green_passports ON green_passports.profile_id = umkm_profiles.profile_id AND green_passports.status = ?", "active")
}

func applyPublicUMKMDirectoryFilters(query *gorm.DB, param model.PublicUMKMDirectoryFiltersParam) *gorm.DB {
	if param.Search != "" {
		search := "%" + param.Search + "%"
		query = query.Where(
			"umkm_profiles.business_name LIKE ? OR umkm_profiles.business_city LIKE ? OR umkm_profiles.business_province LIKE ?",
			search,
			search,
			search,
		)
	}
	if len(param.SectorIDs) > 0 {
		query = query.Where("umkm_profiles.sector_id IN ?", param.SectorIDs)
	}
	if len(param.Provinces) > 0 {
		query = query.Where("umkm_profiles.business_province IN ?", param.Provinces)
	}
	if len(param.Tiers) > 0 {
		query = applyPublicUMKMTierFilters(query, param.Tiers)
	}

	return query
}

func applyPublicUMKMTierFilters(query *gorm.DB, tiers []string) *gorm.DB {
	clauses := make([]string, 0, len(tiers))
	for _, tier := range tiers {
		switch strings.ToLower(strings.TrimSpace(tier)) {
		case "unggul":
			clauses = append(clauses, "green_passports.grs_score >= 85")
		case "siap":
			clauses = append(clauses, "(green_passports.grs_score >= 70 AND green_passports.grs_score < 85)")
		case "hampir":
			clauses = append(clauses, "green_passports.grs_score < 70")
		}
	}

	if len(clauses) == 0 {
		return query
	}

	return query.Where("(" + strings.Join(clauses, " OR ") + ")")
}
