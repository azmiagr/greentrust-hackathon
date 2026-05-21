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
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IUserService interface {
	RegisterUser(param model.RegisterUserParam) (*model.RegisterResponse, error)
	VerifyOTP(sessionToken string, param model.VerifyOTPParam) (*model.VerifyOTPResponse, error)
}

type UserService struct {
	db       *gorm.DB
	userRepo repository.IUserRepository
	otpRepo  repository.IOTPRepository
	jwtAuth  jwt.Interface
	bcrypt   bcrypt.Interface
}

func NewUserService(
	userRepo repository.IUserRepository,
	otpRepo repository.IOTPRepository,
	jwtAuth jwt.Interface,
	bcrypt bcrypt.Interface,
) IUserService {
	return &UserService{
		db:       mariadb.Connection,
		userRepo: userRepo,
		otpRepo:  otpRepo,
		jwtAuth:  jwtAuth,
		bcrypt:   bcrypt,
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

	sessionToken, err := s.jwtAuth.CreateRegistrationSessionToken(user.Email, &user.UserID, false)
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
		newSessionToken, err := s.jwtAuth.CreateRegistrationSessionToken(user.Email, &user.UserID, true)
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

	newSessionToken, err := s.jwtAuth.CreateRegistrationSessionToken(user.Email, &user.UserID, true)
	if err != nil {
		return nil, apperrors.InternalServer("failed to generate registration session token")
	}

	return &model.VerifyOTPResponse{
		SessionToken: newSessionToken,
		Message:      "account verified successfully",
	}, nil
}

func getExpiredOTPDuration() time.Duration {
	expiredOTP, err := strconv.Atoi(os.Getenv("EXPIRED_OTP"))
	if err != nil || expiredOTP <= 0 {
		expiredOTP = 5
	}

	return time.Duration(expiredOTP) * time.Minute
}
