package service

import (
	"errors"
	"fmt"
	"greentrust-hackathon/entity"
	"greentrust-hackathon/internal/repository"
	"greentrust-hackathon/model"
	"greentrust-hackathon/pkg/bcrypt"
	constants "greentrust-hackathon/pkg/constant"
	"greentrust-hackathon/pkg/database/mariadb"
	apperrors "greentrust-hackathon/pkg/errors"
	"greentrust-hackathon/pkg/jwt"
	"greentrust-hackathon/pkg/mail"
	"greentrust-hackathon/pkg/supabase"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IUserService interface {
	RegisterUser(param model.RegisterUserParam) (*model.RegisterResponse, error)
	VerifyOTP(sessionToken string, param model.VerifyOTPParam) (*model.VerifyOTPResponse, error)
	SubmitUserIdentity(sessionToken string, param model.SubmitUserIdentityParam) (*model.SubmitUserIdentityResponse, error)
	SubmitBusinessProfile(sessionToken string, param model.SubmitBusinessProfileParam) (*model.SubmitBusinessProfileResponse, error)
	LoginUser(param model.LoginUserParam) (*model.LoginResponse, error)
	GetUser(param model.GetUserParam) (*entity.User, error)
}

type UserService struct {
	db                 *gorm.DB
	userRepo           repository.IUserRepository
	otpRepo            repository.IOTPRepository
	jwtAuth            jwt.Interface
	bcrypt             bcrypt.Interface
	userIdentityRepo   repository.IUserIdentityRepository
	businessSectorRepo repository.IBusinessSectorRepository
	umkmProfileRepo    repository.IUMKMProfileRepository
	locationPhotoRepo  repository.ILocationPhotoRepository
	supabase           supabase.Interface
}

func NewUserService(
	userRepo repository.IUserRepository,
	otpRepo repository.IOTPRepository,
	jwtAuth jwt.Interface,
	bcrypt bcrypt.Interface,
	userIdentityRepo repository.IUserIdentityRepository,
	businessSectorRepo repository.IBusinessSectorRepository,
	umkmProfileRepo repository.IUMKMProfileRepository,
	locationPhotoRepo repository.ILocationPhotoRepository,
	supabase supabase.Interface,
) IUserService {
	return &UserService{
		db:                 mariadb.Connection,
		userRepo:           userRepo,
		otpRepo:            otpRepo,
		jwtAuth:            jwtAuth,
		bcrypt:             bcrypt,
		userIdentityRepo:   userIdentityRepo,
		businessSectorRepo: businessSectorRepo,
		umkmProfileRepo:    umkmProfileRepo,
		locationPhotoRepo:  locationPhotoRepo,
		supabase:           supabase,
	}
}

func (s *UserService) RegisterUser(param model.RegisterUserParam) (*model.RegisterResponse, error) {
	if param.Password != param.ConfirmPassword {
		return nil, apperrors.BadRequest("password and confirm password do not match")
	}

	hashedPassword, err := s.bcrypt.GenerateFromPassword(param.Password)
	if err != nil {
		return nil, apperrors.InternalServer("failed to generate password")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	user, err := s.userRepo.GetUser(tx, model.GetUserParam{Email: param.Email})
	if err == nil && user.Status == "active" {
		return nil, apperrors.Conflict("email already registered")
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.InternalServer("failed to check existing user")
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = &entity.User{
			UserID:   uuid.New(),
			RoleID:   constants.RoleUMKM,
			Email:    param.Email,
			Password: hashedPassword,
			Status:   "inactive",
		}

		if err := s.userRepo.CreateUser(tx, user); err != nil {
			return nil, apperrors.InternalServer("failed to create user")
		}
	} else {
		user.Password = hashedPassword
		user.Status = "inactive"

		err := s.userRepo.UpdateUser(tx, user)
		if err != nil {
			return nil, apperrors.InternalServer("failed to update inactive user")
		}

		err = s.otpRepo.DeleteOTPByUserID(tx, model.GetOtp{UserID: user.UserID})
		if err != nil {
			return nil, apperrors.InternalServer("failed to reset otp")
		}
	}

	otpCode := mail.GenerateCode()
	otp := &entity.OTP{
		OtpID:  uuid.New(),
		UserID: user.UserID,
		Code:   otpCode,
	}

	err = s.otpRepo.CreateOTP(tx, otp)
	if err != nil {
		return nil, apperrors.InternalServer("failed to create otp")
	}

	sessionToken, err := s.jwtAuth.CreateRegistrationSessionToken(user.Email, &user.UserID, false, false)
	if err != nil {
		return nil, apperrors.InternalServer("failed to generate registration session token")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, apperrors.InternalServer("failed to commit transaction")
	}

	verificationLink := fmt.Sprintf(
		"%s/verify-email?email=%s&code=%s",
		os.Getenv("FRONTEND_BASE_URL"),
		url.QueryEscape(user.Email),
		otpCode,
	)

	err = mail.SendVerificationEmail(user.Email, user.Email, otpCode, verificationLink)
	if err != nil {
		return nil, apperrors.InternalServer("failed to send verification email")
	}

	return &model.RegisterResponse{
		SessionToken: sessionToken,
		Message:      "registration successful, verification email sent",
	}, nil
}

func (s *UserService) VerifyOTP(sessionToken string, param model.VerifyOTPParam) (*model.VerifyOTPResponse, error) {
	if sessionToken == "" {
		return nil, apperrors.Unauthorized("missing session token")
	}

	claims, err := s.jwtAuth.ValidateRegistrationSessionToken(sessionToken)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid session token")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	user, err := s.userRepo.GetUser(tx, model.GetUserParam{
		UserID: *claims.UserID,
		Email:  claims.Email,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}

		return nil, apperrors.InternalServer("failed to get user")
	}

	if user.Status == "active" {
		newSessionToken, err := s.jwtAuth.CreateRegistrationSessionToken(user.Email, &user.UserID, true, claims.IdentityCompleted)
		if err != nil {
			return nil, apperrors.InternalServer("failed to generate registration session token")
		}

		return &model.VerifyOTPResponse{
			SessionToken: newSessionToken,
			Message:      "account already verified",
		}, nil
	}

	otp, err := s.otpRepo.GetOTP(tx, model.GetOtp{
		UserID: user.UserID,
		Code:   param.Code,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.BadRequest("invalid otp code")
		}

		return nil, apperrors.InternalServer("failed to get otp")
	}

	expiredDuration := getExpiredOTPDuration()
	if time.Since(otp.CreatedAt) > expiredDuration {
		if err := s.otpRepo.DeleteOTP(tx, otp); err != nil {
			return nil, apperrors.InternalServer("failed to delete expired otp")
		}

		err := tx.Commit().Error
		if err != nil {
			return nil, apperrors.InternalServer("failed to commit transaction")
		}

		return nil, apperrors.BadRequest("otp code has expired")
	}

	user.Status = "active"
	err = s.userRepo.UpdateUser(tx, user)
	if err != nil {
		return nil, apperrors.InternalServer("failed to activate user")
	}

	err = s.otpRepo.DeleteOTPByUserID(tx, model.GetOtp{UserID: user.UserID})
	if err != nil {
		return nil, apperrors.InternalServer("failed to delete otp")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, apperrors.InternalServer("failed to commit transaction")
	}

	newSessionToken, err := s.jwtAuth.CreateRegistrationSessionToken(user.Email, &user.UserID, true, claims.IdentityCompleted)
	if err != nil {
		return nil, apperrors.InternalServer("failed to generate registration session token")
	}

	return &model.VerifyOTPResponse{
		SessionToken: newSessionToken,
		Message:      "account verified successfully",
	}, nil
}

func (s *UserService) SubmitUserIdentity(sessionToken string, param model.SubmitUserIdentityParam) (*model.SubmitUserIdentityResponse, error) {
	if sessionToken == "" {
		return nil, apperrors.Unauthorized("missing session token")
	}

	claims, err := s.jwtAuth.ValidateRegistrationSessionToken(sessionToken)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid session token")
	}

	if !claims.Verified {
		return nil, apperrors.Forbidden("email verification required")
	}

	birthDate, err := time.Parse("2006-01-02", param.BirthDate)
	if err != nil {
		return nil, apperrors.BadRequest("birth_date must use YYYY-MM-DD format")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	var userID uuid.UUID
	if claims.UserID != nil {
		userID = *claims.UserID
	}

	user, err := s.userRepo.GetUser(tx, model.GetUserParam{
		UserID: userID,
		Email:  claims.Email,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.InternalServer("failed to get user")
	}

	if user.Status != "active" {
		return nil, apperrors.Forbidden("user account is not active")
	}

	existingIdentity, err := s.userIdentityRepo.GetUserIdentity(tx, model.GetUserIdentityParam{
		UserID: user.UserID,
	})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.InternalServer("failed to get user identity")
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		identity := &entity.UserIdentity{
			IdentityID:   uuid.New(),
			UserID:       user.UserID,
			KTPFilePath:  param.KTPFilePath,
			FirstName:    param.FirstName,
			LastName:     param.LastName,
			NIK:          param.NIK,
			BirthPlace:   param.BirthPlace,
			BirthDate:    birthDate,
			Address:      param.Address,
			Province:     param.Province,
			City:         param.City,
			PhoneNumber:  param.PhoneNumber,
			EmailContact: param.EmailContact,
			IsConfirmed:  param.IsConfirmed,
		}

		err = s.userIdentityRepo.CreateUserIdentity(tx, identity)
		if err != nil {
			if isDuplicateEntryError(err) {
				return nil, apperrors.Conflict("nik already registered")
			}
			return nil, apperrors.InternalServer("failed to create user identity")
		}

		err = tx.Commit().Error
		if err != nil {
			return nil, apperrors.InternalServer("failed to commit transaction")
		}

		newSessionToken, err := s.jwtAuth.CreateRegistrationSessionToken(user.Email, &user.UserID, true, true)
		if err != nil {
			return nil, apperrors.InternalServer("failed to generate registration session token")
		}

		return &model.SubmitUserIdentityResponse{
			IdentityID:   identity.IdentityID,
			SessionToken: newSessionToken,
			Message:      "identity submitted successfully",
		}, nil
	}

	existingIdentity.KTPFilePath = param.KTPFilePath
	existingIdentity.FirstName = param.FirstName
	existingIdentity.LastName = param.LastName
	existingIdentity.NIK = param.NIK
	existingIdentity.BirthPlace = param.BirthPlace
	existingIdentity.BirthDate = birthDate
	existingIdentity.Address = param.Address
	existingIdentity.Province = param.Province
	existingIdentity.City = param.City
	existingIdentity.PhoneNumber = param.PhoneNumber
	existingIdentity.EmailContact = param.EmailContact
	existingIdentity.IsConfirmed = param.IsConfirmed

	err = s.userIdentityRepo.UpdateUserIdentity(tx, existingIdentity)
	if err != nil {
		if isDuplicateEntryError(err) {
			return nil, apperrors.Conflict("nik already registered")
		}
		return nil, apperrors.InternalServer("failed to update user identity")
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, apperrors.InternalServer("failed to commit transaction")
	}

	newSessionToken, err := s.jwtAuth.CreateRegistrationSessionToken(user.Email, &user.UserID, true, true)
	if err != nil {
		return nil, apperrors.InternalServer("failed to generate registration session token")
	}

	return &model.SubmitUserIdentityResponse{
		IdentityID:   existingIdentity.IdentityID,
		SessionToken: newSessionToken,
		Message:      "identity submitted successfully",
	}, nil
}

func (s *UserService) SubmitBusinessProfile(sessionToken string, param model.SubmitBusinessProfileParam) (*model.SubmitBusinessProfileResponse, error) {
	if sessionToken == "" {
		return nil, apperrors.Unauthorized("missing session token")
	}

	claims, err := s.jwtAuth.ValidateRegistrationSessionToken(sessionToken)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid session token")
	}

	if !claims.Verified {
		return nil, apperrors.Forbidden("email verification required")
	}

	if len(param.BusinessDescription) > 280 {
		return nil, apperrors.BadRequest("business_description must not exceed 280 characters")
	}
	if len(param.Photos) > 5 {
		return nil, apperrors.BadRequest("business photos must not exceed 5 files")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	user, err := s.userRepo.GetUser(tx, model.GetUserParam{
		UserID: *claims.UserID,
		Email:  claims.Email,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.InternalServer("failed to get user")
	}

	if user.Status != "active" {
		return nil, apperrors.Forbidden("user account is not active")
	}

	_, err = s.userIdentityRepo.GetUserIdentity(tx, model.GetUserIdentityParam{UserID: user.UserID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Forbidden("identity step must be completed first")
		}
		return nil, apperrors.InternalServer("failed to get user identity")
	}

	sectorID := uuid.Nil
	if strings.TrimSpace(param.SectorID) != "" {
		sectorID, err = uuid.Parse(param.SectorID)
		if err != nil {
			return nil, apperrors.BadRequest("sector_id must be a valid uuid")
		}

		_, err = s.businessSectorRepo.GetBusinessSector(tx, model.GetBusinessSectorParam{SectorID: sectorID})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.BadRequest("business sector not found")
			}
			return nil, apperrors.InternalServer("failed to get business sector")
		}
	}

	var uploadedURLs []string
	for _, photo := range param.Photos {
		url, err := s.supabase.UploadFile(photo)
		if err != nil {
			s.supabase.DeleteMultipleFiles(uploadedURLs)
			return nil, apperrors.InternalServer("failed to upload business photo")
		}
		uploadedURLs = append(uploadedURLs, url)
	}

	profile, err := s.umkmProfileRepo.GetUMKMProfile(tx, model.GetUMKMProfileParam{UserID: user.UserID})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.supabase.DeleteMultipleFiles(uploadedURLs)
		return nil, apperrors.InternalServer("failed to get business profile")
	}

	var oldPhotoURLs []string
	var currentPhotoURLs []string
	if errors.Is(err, gorm.ErrRecordNotFound) {
		profile = &entity.UMKMProfile{
			ProfileID:           uuid.New(),
			UserID:              user.UserID,
			SectorID:            sectorID,
			BusinessName:        param.BusinessName,
			BusinessDescription: param.BusinessDescription,
			BusinessAddressLine: param.BusinessAddressLine,
			BusinessProvince:    param.BusinessProvince,
			BusinessCity:        param.BusinessCity,
			IsServiceBusiness:   param.IsServiceBusiness,
			WhatsappNumber:      param.WhatsappNumber,
		}
		profile.ProfileCompletionScore = calculateBusinessProfileCompletionScore(param, len(uploadedURLs))

		err = s.umkmProfileRepo.CreateUMKMProfile(tx, profile)
		if err != nil {
			s.supabase.DeleteMultipleFiles(uploadedURLs)
			return nil, apperrors.InternalServer("failed to create business profile")
		}
	} else {
		oldPhotos, err := s.locationPhotoRepo.GetLocationPhotosByProfileID(tx, profile.ProfileID)
		if err != nil {
			s.supabase.DeleteMultipleFiles(uploadedURLs)
			return nil, apperrors.InternalServer("failed to get existing business photos")
		}

		for _, photo := range oldPhotos {
			oldPhotoURLs = append(oldPhotoURLs, photo.URL)
		}
		currentPhotoURLs = oldPhotoURLs

		profile.SectorID = sectorID
		profile.BusinessName = param.BusinessName
		profile.BusinessDescription = param.BusinessDescription
		profile.BusinessAddressLine = param.BusinessAddressLine
		profile.BusinessProvince = param.BusinessProvince
		profile.BusinessCity = param.BusinessCity
		profile.IsServiceBusiness = param.IsServiceBusiness
		profile.WhatsappNumber = param.WhatsappNumber
		photoCountForScore := len(uploadedURLs)
		if photoCountForScore == 0 {
			photoCountForScore = len(oldPhotos)
		}
		profile.ProfileCompletionScore = calculateBusinessProfileCompletionScore(param, photoCountForScore)

		err = s.umkmProfileRepo.UpdateUMKMProfile(tx, profile)
		if err != nil {
			s.supabase.DeleteMultipleFiles(uploadedURLs)
			return nil, apperrors.InternalServer("failed to update business profile")
		}

		if len(uploadedURLs) > 0 {
			err = s.locationPhotoRepo.DeleteLocationPhotosByProfileID(tx, profile.ProfileID)
			if err != nil {
				s.supabase.DeleteMultipleFiles(uploadedURLs)
				return nil, apperrors.InternalServer("failed to replace business photos")
			}
			currentPhotoURLs = uploadedURLs
		}
	}

	if len(uploadedURLs) > 0 {
		locationPhotos := make([]entity.LocationPhoto, 0, len(uploadedURLs))
		for _, url := range uploadedURLs {
			locationPhotos = append(locationPhotos, entity.LocationPhoto{
				PhotoID:   uuid.New(),
				ProfileID: profile.ProfileID,
				URL:       url,
			})
		}

		err = s.locationPhotoRepo.CreateLocationPhotos(tx, locationPhotos)
		if err != nil {
			s.supabase.DeleteMultipleFiles(uploadedURLs)
			return nil, apperrors.InternalServer("failed to create business photos")
		}
		currentPhotoURLs = uploadedURLs
	}

	err = tx.Commit().Error
	if err != nil {
		s.supabase.DeleteMultipleFiles(uploadedURLs)
		return nil, apperrors.InternalServer("failed to commit transaction")
	}

	if len(uploadedURLs) > 0 && len(oldPhotoURLs) > 0 {
		_ = s.supabase.DeleteMultipleFiles(oldPhotoURLs)
	}

	return &model.SubmitBusinessProfileResponse{
		ProfileID: profile.ProfileID,
		PhotoURLs: currentPhotoURLs,
		Message:   "business profile submitted successfully",
	}, nil
}

func (s *UserService) LoginUser(param model.LoginUserParam) (*model.LoginResponse, error) {
	user, err := s.userRepo.GetUser(s.db, model.GetUserParam{
		Email: param.Email,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Unauthorized("invalid email or password")
		}

		return nil, apperrors.InternalServer("failed to get user")
	}

	err = s.bcrypt.CompareAndHashPassword(user.Password, param.Password)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid email or password")
	}

	if user.Status != "active" {
		return nil, apperrors.Forbidden("account is not verified")
	}

	roleName := "umkm"
	if user.RoleID == constants.RoleInvestor {
		roleName = "investor"
	}

	token, err := s.jwtAuth.CreateJWTToken(user.UserID, roleName)
	if err != nil {
		return nil, apperrors.InternalServer("failed to generate token")
	}

	return &model.LoginResponse{
		Token: token,
	}, nil
}

func (s *UserService) GetUser(param model.GetUserParam) (*entity.User, error) {
	return s.userRepo.GetUser(s.db, param)
}

func calculateBusinessProfileCompletionScore(param model.SubmitBusinessProfileParam, photoCount int) float64 {
	totalFields := 8
	completedFields := 0

	if strings.TrimSpace(param.BusinessName) != "" {
		completedFields++
	}
	if strings.TrimSpace(param.SectorID) != "" {
		completedFields++
	}
	if strings.TrimSpace(param.BusinessDescription) != "" {
		completedFields++
	}
	if strings.TrimSpace(param.BusinessAddressLine) != "" {
		completedFields++
	}
	if strings.TrimSpace(param.BusinessProvince) != "" {
		completedFields++
	}
	if strings.TrimSpace(param.BusinessCity) != "" {
		completedFields++
	}
	if strings.TrimSpace(param.WhatsappNumber) != "" {
		completedFields++
	}
	if photoCount > 0 {
		completedFields++
	}

	return (float64(completedFields) / float64(totalFields)) * 100
}

func isDuplicateEntryError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func getExpiredOTPDuration() time.Duration {
	expiredOTP, err := strconv.Atoi(os.Getenv("EXPIRED_OTP"))
	if err != nil || expiredOTP <= 0 {
		expiredOTP = 5
	}

	return time.Duration(expiredOTP) * time.Minute
}
