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
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IInvestorService interface {
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
}

func NewInvestorService(
	userRepo repository.IUserRepository,
	userIdentityRepo repository.IUserIdentityRepository,
	jwtAuth jwt.Interface,
	investorProfileRepo repository.IInvestorProfileRepository,
	investorPositionRepo repository.IInvestorPositionRepository,
	skillRepo repository.ISkillRepository,
) IInvestorService {
	return &InvestorService{
		db:                   mariadb.Connection,
		userRepo:             userRepo,
		userIdentityRepo:     userIdentityRepo,
		jwtAuth:              jwtAuth,
		investorProfileRepo:  investorProfileRepo,
		investorPositionRepo: investorPositionRepo,
		skillRepo:            skillRepo,
	}
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

	if err := tx.Model(position).Association("Skills").Replace(skills); err != nil {
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

	if err := tx.Model(position).Association("Skills").Replace(skills); err != nil {
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

	if err := tx.Model(position).Association("Skills").Clear(); err != nil {
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
