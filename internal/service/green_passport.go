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
	GetGreenPassportStatus(userID uuid.UUID) (*model.GreenPassportStatusResponse, error)
	GetPublicUMKMDirectory(query model.PublicUMKMDirectoryQuery) (*model.PublicUMKMDirectoryResponse, error)
	GetPublicUMKMDetail(profileID uuid.UUID) (*model.PublicUMKMDetailResponse, error)
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

func (s *GreenPassportService) GetPublicUMKMDirectory(query model.PublicUMKMDirectoryQuery) (*model.PublicUMKMDirectoryResponse, error) {
	param, activeFilterCount, err := buildPublicUMKMDirectoryFiltersParam(query)
	if err != nil {
		return nil, err
	}

	items, err := s.umkmProfileRepo.GetPublicUMKMDirectoryItems(s.db, param)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get public umkm directory")
	}

	total, err := s.umkmProfileRepo.CountPublicUMKMDirectoryItems(s.db, param)
	if err != nil {
		return nil, apperrors.InternalServer("failed to count public umkm directory")
	}

	result := &model.PublicUMKMDirectoryResponse{
		Meta: model.PublicUMKMDirectoryMeta{
			Page:              param.Page,
			Limit:             param.Limit,
			Total:             int(total),
			Showing:           len(items),
			ActiveFilterCount: activeFilterCount,
		},
		Items: mapPublicUMKMDirectoryItems(items),
	}

	if activeFilterCount > 0 {
		filters, err := s.getPublicUMKMDirectoryFilters()
		if err != nil {
			return nil, err
		}
		result.Filters = filters
	}

	return result, nil
}

func (s *GreenPassportService) GetPublicUMKMDetail(profileID uuid.UUID) (*model.PublicUMKMDetailResponse, error) {
	profile, err := s.umkmProfileRepo.GetUMKMProfile(s.db, model.GetUMKMProfileParam{ProfileID: profileID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("business profile not found")
		}
		return nil, apperrors.InternalServer("failed to get business profile")
	}

	passport, err := s.passportRepo.GetActiveByProfileID(s.db, profileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("active green passport not found")
		}
		return nil, apperrors.InternalServer("failed to get green passport")
	}

	directoryRow, err := s.umkmProfileRepo.GetPublicUMKMDirectoryItemByProfileID(s.db, profileID.String())
	if err != nil {
		return nil, apperrors.InternalServer("failed to get public profile metadata")
	}

	categories, err := s.evidenceRepo.GetEvidenceCategoriesWithDocuments(s.db, profileID)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get evidence categories")
	}

	tier, tierLabel := buildPublicUMKMTier(passport.GRSScore)
	categoryResponses, verifiedDocumentCount := mapPublicUMKMDetailCategories(categories)

	photoURL := directoryRow.PhotoURL
	sectorName := directoryRow.SectorName
	if sectorName == "" {
		sectorName = profile.SectorID.String()
	}

	return &model.PublicUMKMDetailResponse{
		Profile: model.PublicUMKMDetailProfile{
			ProfileID:      profile.ProfileID,
			BusinessName:   profile.BusinessName,
			SectorName:     sectorName,
			Province:       profile.BusinessProvince,
			City:           profile.BusinessCity,
			Description:    profile.BusinessDescription,
			PhotoURL:       photoURL,
			WhatsappNumber: profile.WhatsappNumber,
		},
		GreenPassport: model.PublicGreenPassportDetail{
			PassportID:       passport.PassportID,
			PublicSlug:       passport.PublicSlug,
			PassportURL:      buildPassportURL(passport.PublicSlug),
			QRCodeURL:        passport.QRCodeURL,
			GRSScore:         passport.GRSScore,
			Tier:             tier,
			TierLabel:        tierLabel,
			Status:           passport.Status,
			IssuedAt:         passport.IssuedAt,
			LastUpdatedAt:    passport.LastUpdatedAt,
			BlockchainTxHash: passport.BlockchainTxHash,
			ContractAddress:  passport.ContractAddress,
			Network:          passport.NetworkName,
			ChainID:          passport.ChainID,
			BlockNumber:      passport.BlockNumber,
		},
		Summary: model.PublicUMKMDetailSummary{
			VerifiedDocumentCount: verifiedDocumentCount,
			PrivateDocumentCount:  0,
			CategoryCount:         len(categoryResponses),
		},
		Categories: categoryResponses,
	}, nil
}

func (s *GreenPassportService) getPublicUMKMDirectoryFilters() (*model.PublicUMKMDirectoryFilters, error) {
	sectors, err := s.umkmProfileRepo.GetPublicUMKMDirectorySectorFilters(s.db)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get sector filters")
	}

	provinces, err := s.umkmProfileRepo.GetPublicUMKMDirectoryProvinceFilters(s.db)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get province filters")
	}

	tiers, err := s.umkmProfileRepo.GetPublicUMKMDirectoryTierFilters(s.db)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get tier filters")
	}

	return &model.PublicUMKMDirectoryFilters{
		Sectors:   mapPublicUMKMSectorFilters(sectors),
		Provinces: mapPublicUMKMProvinceFilters(provinces),
		Tiers:     mapPublicUMKMTierFilters(tiers),
	}, nil
}

func (s *GreenPassportService) GetGreenPassportStatus(userID uuid.UUID) (*model.GreenPassportStatusResponse, error) {
	profile, err := s.umkmProfileRepo.GetUMKMProfile(s.db, model.GetUMKMProfileParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("business profile not found")
		}
		return nil, apperrors.InternalServer("failed to get business profile")
	}

	passport, err := s.passportRepo.GetByProfileID(s.db, profile.ProfileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.GreenPassportStatusResponse{
				Issued:  false,
				Message: "you haven't been issued a passport yet",
				Profile: &model.GreenPassportStatusProfile{
					ProfileID:    profile.ProfileID,
					BusinessName: profile.BusinessName,
					SectorName:   profile.SectorID.String(),
					Province:     profile.BusinessProvince,
					City:         profile.BusinessCity,
				},
				NextAction: &model.GreenPassportStatusNextAction{
					Title:       "Issue your GreenTrust Passport",
					Description: "Complete the required evidence and issue your passport once your GRS reaches the threshold.",
					TargetScore: evidencePassportThreshold,
				},
			}, nil
		}
		return nil, apperrors.InternalServer("failed to get green passport")
	}

	tier, tierLabel := buildPublicUMKMTier(passport.GRSScore)
	passportURL := buildPassportURL(passport.PublicSlug)

	return &model.GreenPassportStatusResponse{
		Issued:  true,
		Message: "green passport has been issued",
		Profile: &model.GreenPassportStatusProfile{
			ProfileID:    profile.ProfileID,
			BusinessName: profile.BusinessName,
			SectorName:   profile.SectorID.String(),
			Province:     profile.BusinessProvince,
			City:         profile.BusinessCity,
		},
		GreenPassport: &model.GreenPassportStatusDetail{
			PassportID:    passport.PassportID,
			PublicSlug:    passport.PublicSlug,
			PassportURL:   passportURL,
			QRCodeURL:     passport.QRCodeURL,
			GRSScore:      passport.GRSScore,
			Tier:          tier,
			TierLabel:     tierLabel,
			Status:        passport.Status,
			IssuedAt:      passport.IssuedAt,
			LastUpdatedAt: passport.LastUpdatedAt,
		},
		Share: &model.GreenPassportStatusShare{
			URL: passportURL,
		},
		OnChainProof: &model.GreenPassportStatusOnChainProof{
			Network:          passport.NetworkName,
			ChainID:          passport.ChainID,
			ContractAddress:  passport.ContractAddress,
			BlockchainTxHash: passport.BlockchainTxHash,
			ShortTxHash:      shortenTxHash(passport.BlockchainTxHash),
			BlockNumber:      passport.BlockNumber,
			ExplorerURL:      buildTxExplorerURL(passport.BlockchainTxHash),
			Confirmation:     "confirmed",
		},
		NextAction:     buildGreenPassportStatusNextAction(passport.GRSScore),
		CategoryScores: mapGreenPassportCategoryScores(passport.GRSBreakdown),
	}, nil
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

	docs, err := collectIssuableRequiredEvidenceDocuments(categories)
	if err != nil {
		return nil, err
	}

	grs, breakdown := calculatePassportGRS(categories)
	if grs < evidencePassportThreshold {
		return nil, apperrors.BadRequest("grs score is below passport threshold")
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

func collectIssuableRequiredEvidenceDocuments(categories []*entity.EvidenceCategory) ([]*entity.EvidenceDocument, error) {
	docs := make([]*entity.EvidenceDocument, 0)
	missingRequirements := make([]string, 0)

	for _, category := range categories {
		if category == nil {
			continue
		}

		for _, requirement := range category.Requirements {
			doc := selectIssuableEvidenceDocument(requirement.EvidenceDocuments)
			if !requirement.IsRequired {
				if doc != nil {
					docs = append(docs, doc)
				}
				continue
			}

			if doc == nil {
				missingRequirements = append(missingRequirements, fmt.Sprintf("%s - %s", category.CategoryID, requirement.Name))
				continue
			}

			docs = append(docs, doc)
		}
	}

	if len(missingRequirements) > 0 {
		return nil, apperrors.BadRequest("required evidence documents must be reviewed before issuing passport")
	}

	return docs, nil
}

func selectIssuableEvidenceDocument(docs []entity.EvidenceDocument) *entity.EvidenceDocument {
	for i := range docs {
		if isIssuableEvidenceStatus(docs[i].Status) {
			return &docs[i]
		}
	}

	return nil
}

func isIssuableEvidenceStatus(status string) bool {
	switch status {
	case "reviewed", "on_chain":
		return true
	default:
		return false
	}
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

func shortenTxHash(txHash string) string {
	txHash = strings.TrimSpace(txHash)
	if len(txHash) <= 13 {
		return txHash
	}

	return fmt.Sprintf("%s...%s", txHash[:6], txHash[len(txHash)-4:])
}

func buildTxExplorerURL(txHash string) string {
	txHash = strings.TrimSpace(txHash)
	if txHash == "" {
		return ""
	}

	template := strings.TrimSpace(os.Getenv("BLOCKCHAIN_EXPLORER_TX_URL"))
	if template != "" {
		return strings.ReplaceAll(template, "{tx_hash}", txHash)
	}

	baseURL := strings.TrimRight(os.Getenv("BLOCKCHAIN_EXPLORER_BASE_URL"), "/")
	if baseURL == "" {
		return ""
	}

	return fmt.Sprintf("%s/tx/%s", baseURL, txHash)
}

func buildGreenPassportStatusNextAction(grsScore float64) *model.GreenPassportStatusNextAction {
	if grsScore >= 92 {
		return nil
	}

	return &model.GreenPassportStatusNextAction{
		Title:       "Improve your GreenTrust score",
		Description: "Add more verified evidence to unlock a stronger badge on your public profile.",
		TargetScore: 92,
	}
}

func buildPublicUMKMDirectoryFiltersParam(query model.PublicUMKMDirectoryQuery) (model.PublicUMKMDirectoryFiltersParam, int, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 12
	}
	if limit > 50 {
		limit = 50
	}

	sort := strings.TrimSpace(query.Sort)
	if sort == "" {
		sort = "grs_desc"
	}
	switch sort {
	case "grs_desc", "grs_asc", "newest", "name_asc":
	default:
		return model.PublicUMKMDirectoryFiltersParam{}, 0, apperrors.BadRequest("sort must be grs_desc, grs_asc, newest, or name_asc")
	}

	sectorIDs := splitCSV(query.SectorIDs)
	for _, sectorID := range sectorIDs {
		if _, err := uuid.Parse(sectorID); err != nil {
			return model.PublicUMKMDirectoryFiltersParam{}, 0, apperrors.BadRequest("sector_ids must contain valid uuid values")
		}
	}

	tiers := splitCSV(query.Tiers)
	for _, tier := range tiers {
		switch tier {
		case "unggul", "siap", "hampir":
		default:
			return model.PublicUMKMDirectoryFiltersParam{}, 0, apperrors.BadRequest("tiers must contain unggul, siap, or hampir")
		}
	}

	param := model.PublicUMKMDirectoryFiltersParam{
		Search:    strings.TrimSpace(query.Search),
		SectorIDs: sectorIDs,
		Tiers:     tiers,
		Provinces: splitCSVPreserveCase(query.Provinces),
		Sort:      sort,
		Page:      page,
		Limit:     limit,
		Offset:    (page - 1) * limit,
	}

	activeFilterCount := 0
	if param.Search != "" {
		activeFilterCount++
	}
	activeFilterCount += len(param.SectorIDs)
	activeFilterCount += len(param.Tiers)
	activeFilterCount += len(param.Provinces)

	return param, activeFilterCount, nil
}

func splitCSV(input string) []string {
	parts := strings.Split(input, ",")
	values := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		value := strings.ToLower(strings.TrimSpace(part))
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		values = append(values, value)
	}

	return values
}

func splitCSVPreserveCase(input string) []string {
	parts := strings.Split(input, ",")
	values := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		value := strings.TrimSpace(part)
		key := strings.ToLower(value)
		if value == "" || seen[key] {
			continue
		}
		seen[key] = true
		values = append(values, value)
	}

	return values
}

func mapPublicUMKMDirectoryItems(rows []model.PublicUMKMDirectoryRow) []model.PublicUMKMDirectoryItem {
	items := make([]model.PublicUMKMDirectoryItem, 0, len(rows))
	for _, row := range rows {
		profileID, err := uuid.Parse(row.ProfileID)
		if err != nil {
			continue
		}
		passportID, err := uuid.Parse(row.PassportID)
		if err != nil {
			continue
		}

		tier, tierLabel := buildPublicUMKMTier(row.GRSScore)
		items = append(items, model.PublicUMKMDirectoryItem{
			ProfileID:            profileID,
			BusinessName:         row.BusinessName,
			SectorName:           row.SectorName,
			Province:             row.Province,
			City:                 row.City,
			Description:          row.Description,
			PhotoURL:             row.PhotoURL,
			GRSScore:             row.GRSScore,
			Tier:                 tier,
			TierLabel:            tierLabel,
			OnChainDocumentCount: int(row.OnChainDocumentCount),
			GreenPassport: model.PublicGreenPassportBrief{
				PassportID:       passportID,
				PublicSlug:       row.PublicSlug,
				PassportURL:      buildPassportURL(row.PublicSlug),
				Status:           row.PassportStatus,
				IssuedAt:         row.IssuedAt,
				BlockchainTxHash: row.BlockchainTxHash,
			},
		})
	}

	return items
}

func mapPublicUMKMDetailCategories(categories []*entity.EvidenceCategory) ([]model.PublicUMKMDetailEvidenceCategory, int) {
	responses := make([]model.PublicUMKMDetailEvidenceCategory, 0, len(categories))
	verifiedDocumentCount := 0

	for _, category := range categories {
		progress := buildCategoryProgress(category)
		progressPercent := 0.0
		if progress.RequiredCount > 0 {
			progressPercent = roundScore((float64(progress.FulfilledCount) / float64(progress.RequiredCount)) * 100)
		}

		documents := make([]model.PublicUMKMDetailEvidenceDocument, 0)
		for _, requirement := range category.Requirements {
			for _, document := range requirement.EvidenceDocuments {
				if document.Status == "on_chain" {
					verifiedDocumentCount++
				}

				documents = append(documents, model.PublicUMKMDetailEvidenceDocument{
					EvidenceID:       document.EvidenceID,
					RequirementID:    document.RequirementID,
					FileName:         document.OriginalName,
					FileHash:         document.FileHash,
					MimeType:         document.MimeType,
					FileSize:         document.FileSize,
					Status:           document.Status,
					BlockchainTxHash: document.BlockchainTxHash,
					CreatedAt:        document.CreatedAt,
				})
			}
		}

		responses = append(responses, model.PublicUMKMDetailEvidenceCategory{
			CategoryID:      category.CategoryID,
			Code:            category.CategoryID,
			Name:            category.Name,
			Weight:          category.Weight,
			RequiredCount:   progress.RequiredCount,
			FulfilledCount:  progress.FulfilledCount,
			ProgressPercent: progressPercent,
			Score:           progress.Score,
			Status:          progress.Status,
			Documents:       documents,
		})
	}

	return responses, verifiedDocumentCount
}

func mapPublicUMKMSectorFilters(rows []model.PublicUMKMSectorFilterRow) []model.PublicUMKMSectorFilter {
	filters := make([]model.PublicUMKMSectorFilter, 0, len(rows))
	for _, row := range rows {
		sectorID, err := uuid.Parse(row.SectorID)
		if err != nil {
			continue
		}
		filters = append(filters, model.PublicUMKMSectorFilter{
			SectorID:   sectorID,
			SectorName: row.SectorName,
			Count:      int(row.Count),
		})
	}

	return filters
}

func mapPublicUMKMProvinceFilters(rows []model.PublicUMKMProvinceFilterRow) []model.PublicUMKMProvinceFilter {
	filters := make([]model.PublicUMKMProvinceFilter, 0, len(rows))
	for _, row := range rows {
		filters = append(filters, model.PublicUMKMProvinceFilter{
			Name:  row.Name,
			Count: int(row.Count),
		})
	}

	return filters
}

func mapPublicUMKMTierFilters(rows []model.PublicUMKMTierFilterRow) []model.PublicUMKMTierFilter {
	counts := map[string]int{}
	for _, row := range rows {
		counts[row.Tier] = int(row.Count)
	}

	return []model.PublicUMKMTierFilter{
		{Tier: "unggul", Label: "Unggul", MinScore: 85, MaxScore: 100, Count: counts["unggul"]},
		{Tier: "siap", Label: "Siap", MinScore: 70, MaxScore: 84, Count: counts["siap"]},
		{Tier: "hampir", Label: "Hampir", MinScore: 0, MaxScore: 69, Count: counts["hampir"]},
	}
}

func buildPublicUMKMTier(score float64) (string, string) {
	switch {
	case score >= 85:
		return "unggul", "Unggul"
	case score >= 70:
		return "siap", "Siap"
	default:
		return "hampir", "Hampir"
	}
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
