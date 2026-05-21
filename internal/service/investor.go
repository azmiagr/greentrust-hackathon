package service

import (
	"errors"
	"greentrust-hackathon/entity"
	"greentrust-hackathon/internal/repository"
	"greentrust-hackathon/model"
	constants "greentrust-hackathon/pkg/constant"
	"greentrust-hackathon/pkg/database/mariadb"
	apperrors "greentrust-hackathon/pkg/errors"
	"greentrust-hackathon/pkg/jwt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IInvestorService interface {
	GetPublicInvestorDirectory(query model.PublicInvestorDirectoryQuery) (*model.PublicInvestorDirectoryResponse, error)
	GetPublicInvestorDetail(profileID uuid.UUID) (*model.PublicInvestorDetailResponse, error)
	GetInvestorProfile(userID uuid.UUID) (*model.InvestorProfileResponse, error)
	GetInvestorDashboard(userID uuid.UUID) (*model.InvestorDashboardResponse, error)
	GetInvestorPortfolio(userID uuid.UUID, query model.InvestorPortfolioQuery) (*model.InvestorPortfolioResponse, error)
	SubmitInvestorProfileWithSession(sessionToken string, param model.SubmitInvestorProfileParam) (*model.SubmitInvestorProfileResponse, error)
	CreateInvestorPosition(userID uuid.UUID, param model.CreateInvestorPositionParam) (*model.InvestorPositionResponse, error)
	GetInvestorPositions(userID uuid.UUID) ([]model.InvestorPositionResponse, error)
	UpdateInvestorPosition(userID uuid.UUID, positionID uuid.UUID, param model.UpdateInvestorPositionParam) (*model.InvestorPositionResponse, error)
	DeleteInvestorPosition(userID uuid.UUID, positionID uuid.UUID) error
	CreateInvestorPositionWithSession(sessionToken string, param model.CreateInvestorPositionParam) (*model.InvestorPositionResponse, error)
	GetInvestorPositionsWithSession(sessionToken string) ([]model.InvestorPositionResponse, error)
	UpdateInvestorPositionWithSession(sessionToken string, positionID uuid.UUID, param model.UpdateInvestorPositionParam) (*model.InvestorPositionResponse, error)
	DeleteInvestorPositionWithSession(sessionToken string, positionID uuid.UUID) error
	SearchSkills(query string) ([]model.SkillResponse, error)
}

type InvestorService struct {
	db                   *gorm.DB
	userRepo             repository.IUserRepository
	userIdentityRepo     repository.IUserIdentityRepository
	jwtAuth              jwt.Interface
	investorProfileRepo  repository.IInvestorProfileRepository
	investorPositionRepo repository.IInvestorPositionRepository
	skillRepo            repository.ISkillRepository
	businessSectorRepo   repository.IBusinessSectorRepository
	umkmProfileRepo      repository.IUMKMProfileRepository
	proposalRepo         repository.IProposalRepository
}

func NewInvestorService(
	userRepo repository.IUserRepository,
	userIdentityRepo repository.IUserIdentityRepository,
	jwtAuth jwt.Interface,
	investorProfileRepo repository.IInvestorProfileRepository,
	investorPositionRepo repository.IInvestorPositionRepository,
	skillRepo repository.ISkillRepository,
	businessSectorRepo repository.IBusinessSectorRepository,
	umkmProfileRepo repository.IUMKMProfileRepository,
	proposalRepo repository.IProposalRepository,
) IInvestorService {
	return &InvestorService{
		db:                   mariadb.Connection,
		userRepo:             userRepo,
		userIdentityRepo:     userIdentityRepo,
		jwtAuth:              jwtAuth,
		investorProfileRepo:  investorProfileRepo,
		investorPositionRepo: investorPositionRepo,
		skillRepo:            skillRepo,
		businessSectorRepo:   businessSectorRepo,
		umkmProfileRepo:      umkmProfileRepo,
		proposalRepo:         proposalRepo,
	}
}

func (s *InvestorService) GetPublicInvestorDirectory(query model.PublicInvestorDirectoryQuery) (*model.PublicInvestorDirectoryResponse, error) {
	param, activeFilterCount, err := buildPublicInvestorDirectoryFiltersParam(query)
	if err != nil {
		return nil, err
	}

	rows, err := s.investorProfileRepo.GetPublicInvestorDirectoryItems(s.db, param)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get public investor directory")
	}

	total, err := s.investorProfileRepo.CountPublicInvestorDirectoryItems(s.db, param)
	if err != nil {
		return nil, apperrors.InternalServer("failed to count public investor directory")
	}

	investorTypes, err := s.investorProfileRepo.GetPublicInvestorTypeFilters(s.db)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get investor type filters")
	}

	sectors, err := s.investorProfileRepo.GetPublicInvestorSectorFilters(s.db)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get investor sector filters")
	}

	ticketRanges, err := s.investorProfileRepo.GetPublicInvestorTicketRangeFilters(s.db)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get ticket range filters")
	}

	items, err := s.mapPublicInvestorDirectoryItems(rows)
	if err != nil {
		return nil, err
	}

	return &model.PublicInvestorDirectoryResponse{
		Meta: model.PublicInvestorDirectoryMeta{
			Page:              param.Page,
			Limit:             param.Limit,
			Total:             int(total),
			Showing:           len(items),
			ActiveFilterCount: activeFilterCount,
		},
		Filters: model.PublicInvestorDirectoryFilters{
			InvestorTypes: mapPublicInvestorTypeFilters(investorTypes),
			FocusSectors:  mapPublicInvestorSectorFilters(sectors),
			TicketRanges:  mapPublicInvestorTicketRangeFilters(ticketRanges),
		},
		Items: items,
	}, nil
}

func (s *InvestorService) GetPublicInvestorDetail(profileID uuid.UUID) (*model.PublicInvestorDetailResponse, error) {
	row, err := s.investorProfileRepo.GetPublicInvestorDetail(s.db, profileID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("investor profile not found")
		}
		return nil, apperrors.InternalServer("failed to get public investor detail")
	}

	profile, err := s.investorProfileRepo.GetInvestorProfile(s.db, model.GetInvestorProfileParam{ProfileID: profileID})
	if err != nil {
		return nil, apperrors.InternalServer("failed to get investor profile")
	}

	positions, err := s.investorPositionRepo.GetInvestorPositionsByProfileID(s.db, profileID)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get investor positions")
	}

	proposals, err := s.proposalRepo.GetProposalsByUserID(s.db, profile.UserID, "", "")
	if err != nil {
		return nil, apperrors.InternalServer("failed to get investor proposal history")
	}

	fullName := strings.TrimSpace(strings.Join([]string{row.FirstName, row.LastName}, " "))
	userID, err := uuid.Parse(row.UserID)
	if err != nil {
		return nil, apperrors.InternalServer("failed to parse investor user id")
	}

	return &model.PublicInvestorDetailResponse{
		Profile: model.PublicInvestorDetailProfile{
			ProfileID:       profile.ProfileID,
			UserID:          userID,
			FullName:        fullName,
			Initials:        buildInitials(fullName, row.Email),
			Email:           row.Email,
			Title:           row.Title,
			InstitutionName: row.InstitutionName,
			BaseLocation:    row.BaseLocation,
			PublicBio:       row.PublicBio,
			IsVerified:      row.IsVerified,
		},
		InvestmentProfile: *buildInvestorInvestmentProfileResponse(*profile),
		Stats: model.PublicInvestorDetailStats{
			PortfolioCount:    int(row.PortfolioCount),
			AcceptedCount:     int(row.AcceptedCount),
			RejectedCount:     int(row.RejectedCount),
			ApprovalRate:      calculateApprovalRate(row.AcceptedCount, row.RejectedCount),
			ApprovalRateLabel: buildApprovalRateLabel(row.AcceptedCount, row.RejectedCount),
			PositionCount:     len(positions),
		},
		Positions:        mapPublicInvestorDetailPositions(positions),
		Portfolio:        s.mapPublicInvestorDetailPortfolio(proposals, profile.UserID),
		RecentActivities: s.mapPublicInvestorDetailActivities(proposals, profile.UserID),
		ProposalCTA: model.PublicInvestorDetailProposalCTA{
			RequiresAuth:      true,
			ReceiverRole:      "investor",
			ReceiverProfileID: profile.ProfileID,
			Message:           "Silakan masuk atau daftar untuk mengajukan proposal ke investor ini.",
		},
	}, nil
}

func (s *InvestorService) GetInvestorDashboard(userID uuid.UUID) (*model.InvestorDashboardResponse, error) {
	_, err := s.getInvestorUser(s.db, userID, false)
	if err != nil {
		return nil, err
	}

	profile, err := s.getInvestorProfileForUser(s.db, userID, false)
	if err != nil {
		return nil, err
	}

	weekStart := time.Now().AddDate(0, 0, -7)
	proposalStats, err := s.proposalRepo.GetInvestorProposalDashboardStats(s.db, userID)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get investor proposal summary")
	}

	watchedSummary, err := s.umkmProfileRepo.GetInvestorWatchedUMKMSummary(s.db, userID.String(), weekStart)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get watched umkm summary")
	}

	sectorIDs := make([]string, 0, len(profile.FocusSectors))
	for _, sector := range profile.FocusSectors {
		sectorIDs = append(sectorIDs, sector.SectorID.String())
	}

	recommendedRows, err := s.umkmProfileRepo.GetRecommendedUMKMsForInvestor(s.db, sectorIDs, 3)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get recommended umkm")
	}

	recommendedUMKMs := make([]model.InvestorDashboardUMKM, 0, len(recommendedRows))
	for _, row := range recommendedRows {
		profileID, err := uuid.Parse(row.ProfileID)
		if err != nil {
			return nil, apperrors.InternalServer("failed to parse recommended umkm profile id")
		}

		recommendedUMKMs = append(recommendedUMKMs, model.InvestorDashboardUMKM{
			ProfileID:            profileID,
			BusinessName:         row.BusinessName,
			SectorName:           row.SectorName,
			City:                 row.City,
			GRSScore:             row.GRSScore,
			Tier:                 buildGRSTier(row.GRSScore),
			OnChainDocumentCount: int(row.OnChainDocumentCount),
		})
	}

	recentProposals, err := s.proposalRepo.GetRecentInvestorDashboardProposals(s.db, userID, 5)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get recent investor activities")
	}

	return &model.InvestorDashboardResponse{
		Greeting: model.InvestorDashboardGreeting{
			Name:      s.getInvestorDashboardName(userID),
			Subtitle:  "Dashboard Investor",
			Highlight: buildInvestorDashboardHighlight(int(watchedSummary.NewUnggulUMKMThisWeek)),
		},
		Summary: model.InvestorDashboardSummary{
			WatchedUMKMCount:           int(watchedSummary.WatchedUMKMCount),
			WatchedUMKMGrowthThisWeek:  int(watchedSummary.WatchedUMKMGrowthThisWeek),
			ActiveProposalsCount:       int(proposalStats.ActiveProposalsCount),
			AcceptedProposalsCount:     int(proposalStats.AcceptedProposalsCount),
			WaitingConfirmationCount:   int(proposalStats.WaitingConfirmationCount),
			AveragePortfolioGRS:        roundDashboardScore(watchedSummary.AveragePortfolioGRS),
			PortfolioMajorityTier:      buildPortfolioMajorityTier(watchedSummary.ReadyOrUnggulCount, watchedSummary.WatchedUMKMCount),
			NewUnggulUMKMThisWeekCount: int(watchedSummary.NewUnggulUMKMThisWeek),
		},
		RecommendedUMKMs:  recommendedUMKMs,
		RecentActivities:  s.buildInvestorDashboardActivities(recentProposals, userID),
		InvestmentProfile: *buildInvestorInvestmentProfileResponse(*profile),
	}, nil
}

func (s *InvestorService) GetInvestorProfile(userID uuid.UUID) (*model.InvestorProfileResponse, error) {
	user, err := s.getInvestorUser(s.db, userID, false)
	if err != nil {
		return nil, err
	}

	profile, err := s.getInvestorProfileForUser(s.db, userID, false)
	if err != nil {
		return nil, err
	}

	identity, err := s.userIdentityRepo.GetUserIdentity(s.db, model.GetUserIdentityParam{UserID: userID})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.InternalServer("failed to get user identity")
	}

	positions, err := s.investorPositionRepo.GetInvestorPositionsByProfileID(s.db, profile.ProfileID)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get investor positions")
	}

	positionResponses := make([]model.InvestorPositionResponse, 0, len(positions))
	for _, position := range positions {
		positionResponses = append(positionResponses, *buildInvestorPositionResponse(*position))
	}

	primaryPosition := selectPrimaryInvestorPosition(positions)
	fullName := buildInvestorFullName(identity)
	phoneNumber := ""
	if identity != nil {
		phoneNumber = identity.PhoneNumber
	}

	response := &model.InvestorProfileResponse{
		ProfileID:         profile.ProfileID,
		FullName:          fullName,
		Initials:          buildInitials(fullName, user.Email),
		Email:             user.Email,
		PhoneNumber:       phoneNumber,
		InvestmentProfile: *buildInvestorInvestmentProfileResponse(*profile),
		Positions:         positionResponses,
	}

	if primaryPosition != nil {
		response.Company = primaryPosition.InstitutionName
		response.Title = primaryPosition.Title
		response.BaseLocation = primaryPosition.Location
		response.PublicBio = primaryPosition.Description
	}

	return response, nil
}

func (s *InvestorService) GetInvestorPortfolio(userID uuid.UUID, query model.InvestorPortfolioQuery) (*model.InvestorPortfolioResponse, error) {
	_, err := s.getInvestorUser(s.db, userID, false)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(query.SectorID) != "" {
		if _, err := uuid.Parse(strings.TrimSpace(query.SectorID)); err != nil {
			return nil, apperrors.BadRequest("sector_id must be a valid uuid")
		}
		query.SectorID = strings.TrimSpace(query.SectorID)
	}

	rows, err := s.proposalRepo.GetInvestorPortfolio(s.db, userID, query)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get investor portfolio")
	}

	items := make([]model.InvestorPortfolioItem, 0, len(rows))
	for _, row := range rows {
		profileID, err := uuid.Parse(row.ProfileID)
		if err != nil {
			return nil, apperrors.InternalServer("failed to parse portfolio profile id")
		}

		latestProposalID, err := uuid.Parse(row.LatestProposalID)
		if err != nil {
			return nil, apperrors.InternalServer("failed to parse portfolio proposal id")
		}

		items = append(items, model.InvestorPortfolioItem{
			ProfileID:              profileID,
			BusinessName:           row.BusinessName,
			SectorName:             row.SectorName,
			City:                   row.City,
			Status:                 "active",
			TotalValue:             row.TotalValue,
			FormattedTotalValue:    formatRupiahCompact(row.TotalValue),
			CurrentGRS:             roundDashboardScore(row.CurrentGRS),
			GRSTrend:               buildSimpleGRSTrend(row.CurrentGRS),
			FundedSince:            row.FundedSince,
			FundedSinceLabel:       buildFundedSinceLabel(row.FundedSince),
			AcceptedProposalCount:  int(row.AcceptedProposalCount),
			LatestProposalID:       latestProposalID,
			LatestProposalTitle:    row.LatestProposalTitle,
			LatestProposalAccepted: row.LatestProposalAccepted,
		})
	}

	return &model.InvestorPortfolioResponse{Items: items}, nil
}

func (s *InvestorService) SubmitInvestorProfileWithSession(sessionToken string, param model.SubmitInvestorProfileParam) (*model.SubmitInvestorProfileResponse, error) {
	userID, err := s.getOnboardingInvestorUserID(sessionToken)
	if err != nil {
		return nil, err
	}

	startDate, endDate, err := parseInvestorPositionDates(param.Position.StartDate, param.Position.EndDate, param.Position.IsCurrent)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	profile, err := s.getOrCreateInvestorProfile(tx, userID)
	if err != nil {
		return nil, err
	}

	focusSectors, err := s.getInvestorFocusSectors(tx, param.InvestmentProfile.SectorIDs)
	if err != nil {
		return nil, err
	}

	profile.InvestorType = strings.TrimSpace(param.InvestmentProfile.InvestorType)
	profile.TicketRange = strings.TrimSpace(param.InvestmentProfile.TicketRange)
	if profile.InvestorType == "" {
		return nil, apperrors.BadRequest("investor_type is required")
	}
	if profile.TicketRange == "" {
		return nil, apperrors.BadRequest("ticket_range is required")
	}

	err = s.investorProfileRepo.UpdateInvestorProfile(tx, profile)
	if err != nil {
		return nil, apperrors.InternalServer("failed to update investor profile")
	}

	err = s.investorProfileRepo.ReplaceInvestorFocusSectors(tx, profile, focusSectors)
	if err != nil {
		return nil, apperrors.InternalServer("failed to update investor focus sectors")
	}
	profile.FocusSectors = focusSectors

	skills, err := s.getOrCreateSkills(tx, param.Position.Skills)
	if err != nil {
		return nil, err
	}

	position := &entity.InvestorPosition{
		PositionID:      uuid.New(),
		ProfileID:       profile.ProfileID,
		Title:           strings.TrimSpace(param.Position.Title),
		InstitutionName: strings.TrimSpace(param.Position.InstitutionName),
		EmploymentType:  param.Position.EmploymentType,
		Location:        strings.TrimSpace(param.Position.Location),
		StartDate:       startDate,
		EndDate:         endDate,
		IsCurrent:       param.Position.IsCurrent,
		Description:     strings.TrimSpace(param.Position.Description),
	}

	err = s.investorPositionRepo.CreateInvestorPosition(tx, position)
	if err != nil {
		return nil, apperrors.InternalServer("failed to create investor position")
	}

	err = s.investorPositionRepo.ReplaceInvestorPositionSkills(tx, position, skills)
	if err != nil {
		return nil, apperrors.InternalServer("failed to attach skills to investor position")
	}
	position.Skills = skills

	err = tx.Commit().Error
	if err != nil {
		return nil, apperrors.InternalServer("failed to commit transaction")
	}

	return &model.SubmitInvestorProfileResponse{
		ProfileID:         profile.ProfileID,
		Position:          *buildInvestorPositionResponse(*position),
		InvestmentProfile: *buildInvestorInvestmentProfileResponse(*profile),
		Message:           "investor profile submitted successfully",
	}, nil
}

func (s *InvestorService) CreateInvestorPosition(userID uuid.UUID, param model.CreateInvestorPositionParam) (*model.InvestorPositionResponse, error) {
	startDate, endDate, err := parseInvestorPositionDates(param.StartDate, param.EndDate, param.IsCurrent)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	profile, err := s.getOrCreateInvestorProfile(tx, userID)
	if err != nil {
		return nil, err
	}

	skills, err := s.getOrCreateSkills(tx, param.Skills)
	if err != nil {
		return nil, err
	}

	position := &entity.InvestorPosition{
		PositionID:      uuid.New(),
		ProfileID:       profile.ProfileID,
		Title:           strings.TrimSpace(param.Title),
		InstitutionName: strings.TrimSpace(param.InstitutionName),
		EmploymentType:  param.EmploymentType,
		Location:        strings.TrimSpace(param.Location),
		StartDate:       startDate,
		EndDate:         endDate,
		IsCurrent:       param.IsCurrent,
		Description:     strings.TrimSpace(param.Description),
	}

	if err := s.investorPositionRepo.CreateInvestorPosition(tx, position); err != nil {
		return nil, apperrors.InternalServer("failed to create investor position")
	}

	if err := s.investorPositionRepo.ReplaceInvestorPositionSkills(tx, position, skills); err != nil {
		return nil, apperrors.InternalServer("failed to attach skills to investor position")
	}
	position.Skills = skills

	if err := tx.Commit().Error; err != nil {
		return nil, apperrors.InternalServer("failed to commit transaction")
	}

	return buildInvestorPositionResponse(*position), nil
}

func (s *InvestorService) GetInvestorPositions(userID uuid.UUID) ([]model.InvestorPositionResponse, error) {
	_, err := s.getInvestorUser(s.db, userID, false)
	if err != nil {
		return nil, err
	}

	profile, err := s.investorProfileRepo.GetInvestorProfile(s.db, model.GetInvestorProfileParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []model.InvestorPositionResponse{}, nil
		}
		return nil, err
	}

	positions, err := s.investorPositionRepo.GetInvestorPositionsByProfileID(s.db, profile.ProfileID)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get investor positions")
	}

	responses := make([]model.InvestorPositionResponse, 0, len(positions))
	for _, position := range positions {
		responses = append(responses, *buildInvestorPositionResponse(*position))
	}

	return responses, nil
}

func (s *InvestorService) UpdateInvestorPosition(userID uuid.UUID, positionID uuid.UUID, param model.UpdateInvestorPositionParam) (*model.InvestorPositionResponse, error) {
	startDate, endDate, err := parseInvestorPositionDates(param.StartDate, param.EndDate, param.IsCurrent)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	profile, err := s.getInvestorProfileForUser(tx, userID, true)
	if err != nil {
		return nil, err
	}

	position, err := s.investorPositionRepo.GetInvestorPositionByIDAndProfileID(tx, positionID, profile.ProfileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("investor position not found")
		}
		return nil, apperrors.InternalServer("failed to get investor position")
	}

	skills, err := s.getOrCreateSkills(tx, param.Skills)
	if err != nil {
		return nil, err
	}

	position.Title = strings.TrimSpace(param.Title)
	position.InstitutionName = strings.TrimSpace(param.InstitutionName)
	position.EmploymentType = param.EmploymentType
	position.Location = strings.TrimSpace(param.Location)
	position.StartDate = startDate
	position.EndDate = endDate
	position.IsCurrent = param.IsCurrent
	position.Description = strings.TrimSpace(param.Description)

	if err := s.investorPositionRepo.UpdateInvestorPosition(tx, position); err != nil {
		return nil, apperrors.InternalServer("failed to update investor position")
	}

	if err := s.investorPositionRepo.ReplaceInvestorPositionSkills(tx, position, skills); err != nil {
		return nil, apperrors.InternalServer("failed to update investor position skills")
	}
	position.Skills = skills

	if err := tx.Commit().Error; err != nil {
		return nil, apperrors.InternalServer("failed to commit transaction")
	}

	return buildInvestorPositionResponse(*position), nil
}

func (s *InvestorService) DeleteInvestorPosition(userID uuid.UUID, positionID uuid.UUID) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	profile, err := s.getInvestorProfileForUser(tx, userID, true)
	if err != nil {
		return err
	}

	position, err := s.investorPositionRepo.GetInvestorPositionByIDAndProfileID(tx, positionID, profile.ProfileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NotFound("investor position not found")
		}
		return apperrors.InternalServer("failed to get investor position")
	}

	if err := s.investorPositionRepo.ClearInvestorPositionSkills(tx, position); err != nil {
		return apperrors.InternalServer("failed to clear investor position skills")
	}

	if err := s.investorPositionRepo.DeleteInvestorPosition(tx, position); err != nil {
		return apperrors.InternalServer("failed to delete investor position")
	}

	if err := tx.Commit().Error; err != nil {
		return apperrors.InternalServer("failed to commit transaction")
	}

	return nil
}

func (s *InvestorService) CreateInvestorPositionWithSession(sessionToken string, param model.CreateInvestorPositionParam) (*model.InvestorPositionResponse, error) {
	userID, err := s.getOnboardingInvestorUserID(sessionToken)
	if err != nil {
		return nil, err
	}

	return s.CreateInvestorPosition(userID, param)
}

func (s *InvestorService) GetInvestorPositionsWithSession(sessionToken string) ([]model.InvestorPositionResponse, error) {
	userID, err := s.getOnboardingInvestorUserID(sessionToken)
	if err != nil {
		return nil, err
	}

	return s.GetInvestorPositions(userID)
}

func (s *InvestorService) UpdateInvestorPositionWithSession(sessionToken string, positionID uuid.UUID, param model.UpdateInvestorPositionParam) (*model.InvestorPositionResponse, error) {
	userID, err := s.getOnboardingInvestorUserID(sessionToken)
	if err != nil {
		return nil, err
	}

	return s.UpdateInvestorPosition(userID, positionID, param)
}

func (s *InvestorService) DeleteInvestorPositionWithSession(sessionToken string, positionID uuid.UUID) error {
	userID, err := s.getOnboardingInvestorUserID(sessionToken)
	if err != nil {
		return err
	}

	return s.DeleteInvestorPosition(userID, positionID)
}

func (s *InvestorService) SearchSkills(query string) ([]model.SkillResponse, error) {
	skills, err := s.skillRepo.SearchSkills(s.db, strings.TrimSpace(query), 20)
	if err != nil {
		return nil, apperrors.InternalServer("failed to search skills")
	}

	responses := make([]model.SkillResponse, 0, len(skills))
	for _, skill := range skills {
		responses = append(responses, model.SkillResponse{
			SkillID: skill.SkillID,
			Name:    skill.Name,
		})
	}

	return responses, nil
}

func (s *InvestorService) getOrCreateInvestorProfile(tx *gorm.DB, userID uuid.UUID) (*entity.InvestorProfile, error) {
	_, err := s.getInvestorUser(tx, userID, true)
	if err != nil {
		return nil, err
	}

	profile, err := s.investorProfileRepo.GetInvestorProfile(tx, model.GetInvestorProfileParam{UserID: userID})
	if err == nil {
		return profile, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.InternalServer("failed to get investor profile")
	}

	profile = &entity.InvestorProfile{
		ProfileID: uuid.New(),
		UserID:    userID,
	}
	if err := s.investorProfileRepo.CreateInvestorProfile(tx, profile); err != nil {
		return nil, apperrors.InternalServer("failed to create investor profile")
	}

	return profile, nil
}

func (s *InvestorService) getOnboardingInvestorUserID(sessionToken string) (uuid.UUID, error) {
	if sessionToken == "" {
		return uuid.Nil, apperrors.Unauthorized("missing session token")
	}

	claims, err := s.jwtAuth.ValidateRegistrationSessionToken(sessionToken)
	if err != nil {
		return uuid.Nil, apperrors.Unauthorized("invalid session token")
	}
	if !claims.Verified {
		return uuid.Nil, apperrors.Forbidden("email verification required")
	}
	if !claims.IdentityCompleted {
		return uuid.Nil, apperrors.Forbidden("identity step must be completed first")
	}
	if claims.UserID == nil {
		return uuid.Nil, apperrors.Unauthorized("invalid session token")
	}

	return *claims.UserID, nil
}

func (s *InvestorService) getInvestorProfileForUser(tx *gorm.DB, userID uuid.UUID, requireIdentity bool) (*entity.InvestorProfile, error) {
	_, err := s.getInvestorUser(tx, userID, requireIdentity)
	if err != nil {
		return nil, err
	}

	profile, err := s.investorProfileRepo.GetInvestorProfile(tx, model.GetInvestorProfileParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("investor profile not found")
		}
		return nil, apperrors.InternalServer("failed to get investor profile")
	}

	return profile, nil
}

func (s *InvestorService) getInvestorUser(tx *gorm.DB, userID uuid.UUID, requireIdentity bool) (*entity.User, error) {
	user, err := s.userRepo.GetUser(tx, model.GetUserParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.InternalServer("failed to get user")
	}
	if user.RoleID != constants.RoleInvestor {
		return nil, apperrors.Forbidden("investor role required")
	}

	if requireIdentity {
		_, err := s.userIdentityRepo.GetUserIdentity(tx, model.GetUserIdentityParam{UserID: user.UserID})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.Forbidden("identity step must be completed first")
			}
			return nil, apperrors.InternalServer("failed to get user identity")
		}
	}

	return user, nil
}

func (s *InvestorService) getOrCreateSkills(tx *gorm.DB, skillNames []string) ([]entity.Skill, error) {
	seen := map[string]bool{}
	skills := make([]entity.Skill, 0, len(skillNames))

	for _, skillName := range skillNames {
		normalized := strings.TrimSpace(skillName)
		if normalized == "" {
			continue
		}

		key := strings.ToLower(normalized)
		if seen[key] {
			continue
		}
		seen[key] = true

		skill, err := s.skillRepo.GetSkillByName(tx, normalized)
		if err == nil {
			skills = append(skills, *skill)
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.InternalServer("failed to get skill")
		}

		newSkill := &entity.Skill{
			SkillID: uuid.New(),
			Name:    normalized,
		}
		if err := s.skillRepo.CreateSkill(tx, newSkill); err != nil {
			if isDuplicateEntryError(err) {
				skill, err := s.skillRepo.GetSkillByName(tx, normalized)
				if err != nil {
					return nil, apperrors.InternalServer("failed to get existing skill")
				}
				skills = append(skills, *skill)
				continue
			}
			return nil, apperrors.InternalServer("failed to create skill")
		}

		skills = append(skills, *newSkill)
	}

	return skills, nil
}

func (s *InvestorService) getInvestorFocusSectors(tx *gorm.DB, sectorIDInputs []string) ([]entity.BusinessSector, error) {
	if len(sectorIDInputs) == 0 {
		return nil, apperrors.BadRequest("sector_ids is required")
	}
	if len(sectorIDInputs) > 3 {
		return nil, apperrors.BadRequest("sector_ids must not exceed 3 items")
	}

	seen := map[uuid.UUID]bool{}
	sectors := make([]entity.BusinessSector, 0, len(sectorIDInputs))
	for _, sectorIDInput := range sectorIDInputs {
		sectorID, err := uuid.Parse(strings.TrimSpace(sectorIDInput))
		if err != nil {
			return nil, apperrors.BadRequest("sector_ids must contain valid uuid values")
		}
		if seen[sectorID] {
			continue
		}
		seen[sectorID] = true

		sector, err := s.businessSectorRepo.GetBusinessSector(tx, model.GetBusinessSectorParam{SectorID: sectorID})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.BadRequest("business sector not found")
			}
			return nil, apperrors.InternalServer("failed to get business sector")
		}

		sectors = append(sectors, *sector)
	}

	if len(sectors) == 0 {
		return nil, apperrors.BadRequest("sector_ids is required")
	}

	return sectors, nil
}

func parseInvestorPositionDates(startDateInput string, endDateInput string, isCurrent bool) (time.Time, *time.Time, error) {
	startDate, err := time.Parse("2006-01-02", startDateInput)
	if err != nil {
		return time.Time{}, nil, apperrors.BadRequest("start_date must use YYYY-MM-DD format")
	}

	if isCurrent {
		if strings.TrimSpace(endDateInput) != "" {
			return time.Time{}, nil, apperrors.BadRequest("end_date must be empty when position is current")
		}
		return startDate, nil, nil
	}

	if strings.TrimSpace(endDateInput) == "" {
		return time.Time{}, nil, apperrors.BadRequest("end_date is required when position is not current")
	}

	endDate, err := time.Parse("2006-01-02", endDateInput)
	if err != nil {
		return time.Time{}, nil, apperrors.BadRequest("end_date must use YYYY-MM-DD format")
	}
	if endDate.Before(startDate) {
		return time.Time{}, nil, apperrors.BadRequest("end_date must be after start_date")
	}

	return startDate, &endDate, nil
}

func buildInvestorPositionResponse(position entity.InvestorPosition) *model.InvestorPositionResponse {
	startDate := position.StartDate.Format("2006-01-02")
	var endDate *string
	if position.EndDate != nil {
		formattedEndDate := position.EndDate.Format("2006-01-02")
		endDate = &formattedEndDate
	}

	skills := make([]model.SkillResponse, 0, len(position.Skills))
	for _, skill := range position.Skills {
		skills = append(skills, model.SkillResponse{
			SkillID: skill.SkillID,
			Name:    skill.Name,
		})
	}

	return &model.InvestorPositionResponse{
		PositionID:      position.PositionID,
		Title:           position.Title,
		InstitutionName: position.InstitutionName,
		EmploymentType:  position.EmploymentType,
		Location:        position.Location,
		StartDate:       startDate,
		EndDate:         endDate,
		IsCurrent:       position.IsCurrent,
		Description:     position.Description,
		Skills:          skills,
	}
}

func buildInvestorInvestmentProfileResponse(profile entity.InvestorProfile) *model.InvestorInvestmentProfileResponse {
	focusSectors := make([]model.BusinessSectorResponse, 0, len(profile.FocusSectors))
	for _, sector := range profile.FocusSectors {
		focusSectors = append(focusSectors, model.BusinessSectorResponse{
			SectorID:   sector.SectorID,
			SectorName: sector.SectorName,
		})
	}

	return &model.InvestorInvestmentProfileResponse{
		InvestorType: profile.InvestorType,
		TicketRange:  profile.TicketRange,
		FocusSectors: focusSectors,
	}
}

func (s *InvestorService) getInvestorDashboardName(userID uuid.UUID) string {
	identity, err := s.userIdentityRepo.GetUserIdentity(s.db, model.GetUserIdentityParam{UserID: userID})
	if err == nil {
		name := strings.TrimSpace(strings.Join([]string{identity.FirstName, identity.LastName}, " "))
		if name != "" {
			return name
		}
	}

	user, err := s.userRepo.GetUser(s.db, model.GetUserParam{UserID: userID})
	if err != nil {
		return "Investor"
	}

	return user.Email
}

func (s *InvestorService) buildInvestorDashboardActivities(proposals []*entity.Proposal, viewerUserID uuid.UUID) []model.InvestorDashboardActivity {
	activities := make([]model.InvestorDashboardActivity, 0, len(proposals))
	for _, proposal := range proposals {
		umkmName := s.getDashboardProposalUMKMName(proposal)
		if umkmName == "" {
			umkmName = "UMKM"
		}

		activity := model.InvestorDashboardActivity{
			Type:      "proposal_" + proposal.Status,
			CreatedAt: proposal.UpdatedAt,
		}

		switch proposal.Status {
		case "accepted":
			activity.Title = "Proposal Diterima"
			if proposal.SenderUserID == viewerUserID {
				activity.Description = umkmName + " menerima proposal pendanaan Anda."
			} else {
				activity.Description = "Anda menerima proposal dari " + umkmName + "."
			}
		case "sent":
			activity.Title = "Proposal Terkirim"
			if proposal.SenderUserID == viewerUserID {
				activity.Description = "Proposal Anda terkirim ke " + umkmName + "."
			} else {
				activity.Description = umkmName + " mengirim proposal kepada Anda."
			}
		case "rejected":
			activity.Title = "Proposal Ditolak"
			activity.Description = "Proposal terkait " + umkmName + " ditolak."
		case "withdrawn":
			activity.Title = "Proposal Ditarik"
			activity.Description = "Proposal terkait " + umkmName + " ditarik."
		default:
			activity.Title = "Aktivitas Proposal"
			activity.Description = "Ada pembaruan proposal terkait " + umkmName + "."
		}

		activities = append(activities, activity)
	}

	return activities
}

func (s *InvestorService) getDashboardProposalUMKMName(proposal *entity.Proposal) string {
	if proposal.SenderRole == "umkm" {
		profile, err := s.umkmProfileRepo.GetUMKMProfile(s.db, model.GetUMKMProfileParam{ProfileID: proposal.SenderProfileID})
		if err == nil {
			return profile.BusinessName
		}
	}

	if proposal.ReceiverRole == "umkm" {
		profile, err := s.umkmProfileRepo.GetUMKMProfile(s.db, model.GetUMKMProfileParam{ProfileID: proposal.ReceiverProfileID})
		if err == nil {
			return profile.BusinessName
		}
	}

	return ""
}

func selectPrimaryInvestorPosition(positions []*entity.InvestorPosition) *entity.InvestorPosition {
	if len(positions) == 0 {
		return nil
	}

	for _, position := range positions {
		if position.IsCurrent {
			return position
		}
	}

	return positions[0]
}

func buildInvestorFullName(identity *entity.UserIdentity) string {
	if identity == nil {
		return ""
	}

	return strings.TrimSpace(strings.Join([]string{identity.FirstName, identity.LastName}, " "))
}

func buildInitials(name string, fallback string) string {
	source := strings.TrimSpace(name)
	if source == "" {
		source = strings.TrimSpace(fallback)
	}
	if source == "" {
		return "I"
	}

	parts := strings.Fields(source)
	if len(parts) == 1 {
		return strings.ToUpper(string([]rune(parts[0])[0]))
	}

	first := []rune(parts[0])
	last := []rune(parts[len(parts)-1])
	return strings.ToUpper(string(first[0]) + string(last[0]))
}

func buildInvestorDashboardHighlight(newUnggulCount int) string {
	if newUnggulCount == 0 {
		return "Belum ada UMKM baru mencapai tier Unggul minggu ini."
	}

	return "Ada " + strconv.Itoa(newUnggulCount) + " UMKM baru mencapai tier Unggul minggu ini."
}

func buildPortfolioMajorityTier(readyOrUnggulCount int64, totalCount int64) string {
	if totalCount == 0 {
		return "Belum ada portfolio"
	}
	if readyOrUnggulCount*2 >= totalCount {
		return "Siap & Unggul"
	}

	return "Berkembang"
}

func buildGRSTier(score float64) string {
	switch {
	case score >= 85:
		return "Unggul"
	case score >= 70:
		return "Siap"
	default:
		return "Berkembang"
	}
}

func roundDashboardScore(score float64) float64 {
	return math.Round(score*100) / 100
}

func buildPublicInvestorDirectoryFiltersParam(query model.PublicInvestorDirectoryQuery) (model.PublicInvestorDirectoryFiltersParam, int, error) {
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
		sort = "approval_rate_desc"
	}
	switch sort {
	case "approval_rate_desc", "portfolio_desc", "newest", "name_asc":
	default:
		return model.PublicInvestorDirectoryFiltersParam{}, 0, apperrors.BadRequest("sort must be approval_rate_desc, portfolio_desc, newest, or name_asc")
	}

	sectorIDs := splitInvestorCSV(query.SectorIDs, true)
	for _, sectorID := range sectorIDs {
		if _, err := uuid.Parse(sectorID); err != nil {
			return model.PublicInvestorDirectoryFiltersParam{}, 0, apperrors.BadRequest("sector_ids must contain valid uuid values")
		}
	}

	param := model.PublicInvestorDirectoryFiltersParam{
		Search:        strings.TrimSpace(query.Search),
		InvestorTypes: splitInvestorCSV(query.InvestorTypes, false),
		SectorIDs:     sectorIDs,
		TicketRanges:  splitInvestorCSV(query.TicketRanges, false),
		Sort:          sort,
		Page:          page,
		Limit:         limit,
		Offset:        (page - 1) * limit,
	}

	activeFilterCount := 0
	if param.Search != "" {
		activeFilterCount++
	}
	activeFilterCount += len(param.InvestorTypes)
	activeFilterCount += len(param.SectorIDs)
	activeFilterCount += len(param.TicketRanges)

	return param, activeFilterCount, nil
}

func splitInvestorCSV(input string, lower bool) []string {
	parts := strings.Split(input, ",")
	values := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if lower {
			value = strings.ToLower(value)
		}
		key := strings.ToLower(value)
		if value == "" || seen[key] {
			continue
		}
		seen[key] = true
		values = append(values, value)
	}

	return values
}

func (s *InvestorService) mapPublicInvestorDirectoryItems(rows []model.PublicInvestorDirectoryRow) ([]model.PublicInvestorDirectoryItem, error) {
	profileIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		profileIDs = append(profileIDs, row.ProfileID)
	}

	profiles, err := s.investorProfileRepo.GetInvestorProfilesByIDs(s.db, profileIDs)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get investor focus sectors")
	}

	focusSectorsByProfileID := map[string][]model.BusinessSectorResponse{}
	for _, profile := range profiles {
		sectors := make([]model.BusinessSectorResponse, 0, len(profile.FocusSectors))
		for _, sector := range profile.FocusSectors {
			sectors = append(sectors, model.BusinessSectorResponse{
				SectorID:   sector.SectorID,
				SectorName: sector.SectorName,
			})
		}
		focusSectorsByProfileID[profile.ProfileID.String()] = sectors
	}

	items := make([]model.PublicInvestorDirectoryItem, 0, len(rows))
	for _, row := range rows {
		profileID, err := uuid.Parse(row.ProfileID)
		if err != nil {
			continue
		}
		userID, err := uuid.Parse(row.UserID)
		if err != nil {
			continue
		}

		fullName := strings.TrimSpace(strings.Join([]string{row.FirstName, row.LastName}, " "))
		items = append(items, model.PublicInvestorDirectoryItem{
			ProfileID:       profileID,
			UserID:          userID,
			FullName:        fullName,
			Initials:        buildInitials(fullName, row.Email),
			Title:           row.Title,
			InstitutionName: row.InstitutionName,
			InvestorType:    row.InvestorType,
			TicketRange:     row.TicketRange,
			FocusSectors:    focusSectorsByProfileID[row.ProfileID],
			PortfolioCount:  int(row.PortfolioCount),
			ApprovalRate:    calculateApprovalRate(row.AcceptedCount, row.RejectedCount),
		})
	}

	return items, nil
}

func mapPublicInvestorTypeFilters(rows []model.PublicInvestorTypeFilterRow) []model.PublicInvestorTypeFilter {
	filters := make([]model.PublicInvestorTypeFilter, 0, len(rows))
	for _, row := range rows {
		filters = append(filters, model.PublicInvestorTypeFilter{
			InvestorType: row.InvestorType,
			Count:        int(row.Count),
		})
	}

	return filters
}

func mapPublicInvestorSectorFilters(rows []model.PublicInvestorSectorFilterRow) []model.PublicInvestorSectorFilter {
	filters := make([]model.PublicInvestorSectorFilter, 0, len(rows))
	for _, row := range rows {
		sectorID, err := uuid.Parse(row.SectorID)
		if err != nil {
			continue
		}
		filters = append(filters, model.PublicInvestorSectorFilter{
			SectorID:   sectorID,
			SectorName: row.SectorName,
			Count:      int(row.Count),
		})
	}

	return filters
}

func mapPublicInvestorTicketRangeFilters(rows []model.PublicInvestorTicketRangeFilterRow) []model.PublicInvestorTicketRangeFilter {
	filters := make([]model.PublicInvestorTicketRangeFilter, 0, len(rows))
	for _, row := range rows {
		filters = append(filters, model.PublicInvestorTicketRangeFilter{
			TicketRange: row.TicketRange,
			Count:       int(row.Count),
		})
	}

	return filters
}

func calculateApprovalRate(acceptedCount int64, rejectedCount int64) int {
	total := acceptedCount + rejectedCount
	if total == 0 {
		return 0
	}

	return int(((acceptedCount * 100) + (total / 2)) / total)
}

func buildApprovalRateLabel(acceptedCount int64, rejectedCount int64) string {
	return strconv.Itoa(calculateApprovalRate(acceptedCount, rejectedCount)) + "% (" + strconv.FormatInt(acceptedCount, 10) + "/" + strconv.FormatInt(acceptedCount+rejectedCount, 10) + ")"
}

func mapPublicInvestorDetailPositions(positions []*entity.InvestorPosition) []model.PublicInvestorDetailPosition {
	responses := make([]model.PublicInvestorDetailPosition, 0, len(positions))
	for _, position := range positions {
		base := buildInvestorPositionResponse(*position)
		responses = append(responses, model.PublicInvestorDetailPosition{
			PositionID:      base.PositionID,
			Initials:        buildInitials(position.InstitutionName, position.Title),
			Title:           base.Title,
			InstitutionName: base.InstitutionName,
			EmploymentType:  base.EmploymentType,
			Location:        base.Location,
			StartDate:       base.StartDate,
			EndDate:         base.EndDate,
			IsCurrent:       base.IsCurrent,
			DurationLabel:   buildDateRangeDurationLabel(position.StartDate, position.EndDate),
			Description:     base.Description,
			Skills:          base.Skills,
		})
	}

	return responses
}

func (s *InvestorService) mapPublicInvestorDetailPortfolio(proposals []*entity.Proposal, investorUserID uuid.UUID) []model.PublicInvestorDetailPortfolioItem {
	items := make([]model.PublicInvestorDetailPortfolioItem, 0)
	seen := map[uuid.UUID]bool{}
	for _, proposal := range proposals {
		if proposal.Status != "accepted" {
			continue
		}

		umkmProfileID := uuid.Nil
		if proposal.SenderRole == "umkm" {
			umkmProfileID = proposal.SenderProfileID
		}
		if proposal.ReceiverRole == "umkm" {
			umkmProfileID = proposal.ReceiverProfileID
		}
		if umkmProfileID == uuid.Nil || seen[umkmProfileID] {
			continue
		}
		seen[umkmProfileID] = true

		profile, err := s.umkmProfileRepo.GetUMKMProfile(s.db, model.GetUMKMProfileParam{ProfileID: umkmProfileID})
		if err != nil {
			continue
		}

		fundedAt := proposal.UpdatedAt
		if proposal.AcceptedAt != nil {
			fundedAt = *proposal.AcceptedAt
		}

		items = append(items, model.PublicInvestorDetailPortfolioItem{
			ProfileID:         profile.ProfileID,
			BusinessName:      profile.BusinessName,
			SectorName:        profile.SectorID.String(),
			City:              profile.BusinessCity,
			FundedYear:        fundedAt.Year(),
			DurationLabel:     buildDateRangeDurationLabel(fundedAt, nil),
			ProposalType:      proposal.ProposalType,
			ProposalTypeLabel: buildInvestorProposalTypeLabel(proposal.ProposalType),
		})
	}

	return items
}

func (s *InvestorService) mapPublicInvestorDetailActivities(proposals []*entity.Proposal, investorUserID uuid.UUID) []model.PublicInvestorDetailActivity {
	activities := make([]model.PublicInvestorDetailActivity, 0)
	for _, proposal := range proposals {
		if proposal.Status != "accepted" {
			continue
		}

		umkmName := s.getDashboardProposalUMKMName(proposal)
		if umkmName == "" {
			umkmName = "UMKM"
		}

		createdAt := proposal.UpdatedAt
		if proposal.AcceptedAt != nil {
			createdAt = *proposal.AcceptedAt
		}

		activities = append(activities, model.PublicInvestorDetailActivity{
			Type:        "proposal_accepted",
			Title:       "Proposal disetujui",
			Description: "Proposal dengan " + umkmName + " disetujui.",
			CreatedAt:   createdAt,
		})

		if len(activities) >= 5 {
			break
		}
	}

	return activities
}

func buildDateRangeDurationLabel(start time.Time, end *time.Time) string {
	until := time.Now()
	if end != nil {
		until = *end
	}
	if until.Before(start) {
		return "0 bln"
	}

	months := (until.Year()-start.Year())*12 + int(until.Month()-start.Month())
	if until.Day() < start.Day() {
		months--
	}
	if months < 0 {
		months = 0
	}

	years := months / 12
	remainingMonths := months % 12
	if years == 0 {
		return strconv.Itoa(remainingMonths) + " bln"
	}
	if remainingMonths == 0 {
		return strconv.Itoa(years) + " thn"
	}

	return strconv.Itoa(years) + " thn " + strconv.Itoa(remainingMonths) + " bln"
}

func buildInvestorProposalTypeLabel(proposalType string) string {
	switch proposalType {
	case "funding":
		return "Pinjaman"
	case "supply":
		return "Pengadaan"
	default:
		return proposalType
	}
}

func formatRupiahCompact(amount int64) string {
	switch {
	case amount >= 1_000_000_000 && amount%1_000_000_000 == 0:
		return "Rp " + strconv.FormatInt(amount/1_000_000_000, 10) + " Miliar"
	case amount >= 1_000_000_000:
		value := float64(amount) / 1_000_000_000
		return "Rp " + formatCompactDecimal(value) + " Miliar"
	case amount >= 1_000_000 && amount%1_000_000 == 0:
		return "Rp " + strconv.FormatInt(amount/1_000_000, 10) + " Juta"
	case amount >= 1_000_000:
		value := float64(amount) / 1_000_000
		return "Rp " + formatCompactDecimal(value) + " Juta"
	default:
		return "Rp " + strconv.FormatInt(amount, 10)
	}
}

func formatCompactDecimal(value float64) string {
	rounded := math.Round(value*10) / 10
	if rounded == math.Trunc(rounded) {
		return strconv.FormatInt(int64(rounded), 10)
	}

	return strconv.FormatFloat(rounded, 'f', 1, 64)
}

func buildSimpleGRSTrend(score float64) string {
	if score > 0 {
		return "up"
	}

	return "stable"
}

func buildFundedSinceLabel(fundedSince *time.Time) string {
	if fundedSince == nil {
		return ""
	}

	return "Sejak " + indonesianShortMonth(fundedSince.Month()) + " " + strconv.Itoa(fundedSince.Year())
}

func indonesianShortMonth(month time.Month) string {
	months := map[time.Month]string{
		time.January:   "Jan",
		time.February:  "Feb",
		time.March:     "Mar",
		time.April:     "Apr",
		time.May:       "Mei",
		time.June:      "Jun",
		time.July:      "Jul",
		time.August:    "Agu",
		time.September: "Sep",
		time.October:   "Okt",
		time.November:  "Nov",
		time.December:  "Des",
	}

	return months[month]
}
