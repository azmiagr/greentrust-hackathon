package jwt

import (
	"errors"
	"greentrust-hackathon/entity"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Interface interface {
	CreateJWTToken(userID uuid.UUID, roleName string) (string, error)
	ValidateToken(tokenString string) (uuid.UUID, error)
	CreateRegistrationSessionToken(email string, userID *uuid.UUID, verified bool, identityCompleted bool) (string, error)
	ValidateRegistrationSessionToken(tokenString string) (*RegistrationSessionClaims, error)
	GetLoginUser(c *gin.Context) (*entity.User, error)
}

type jsonWebToken struct {
	SecretKey   string
	ExpiredTime time.Duration
}

type RegistrationSessionClaims struct {
	Email             string     `json:"email"`
	UserID            *uuid.UUID `json:"user_id,omitempty"`
	Verified          bool       `json:"verified"`
	IdentityCompleted bool       `json:"identity_completed"`
	Purpose           string     `json:"purpose"`
	jwt.RegisteredClaims
}

type Claims struct {
	UserID   uuid.UUID
	RoleName string
	jwt.RegisteredClaims
}

func Init() Interface {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	expiredTime, err := strconv.Atoi(os.Getenv("JWT_EXP_TIME"))
	if err != nil {
		log.Fatalf("error init jwt %v", err)
	}

	return &jsonWebToken{
		SecretKey:   secretKey,
		ExpiredTime: time.Duration(expiredTime) * time.Hour,
	}
}

func (j *jsonWebToken) CreateJWTToken(userID uuid.UUID, roleName string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.ExpiredTime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j *jsonWebToken) ValidateToken(tokenString string) (uuid.UUID, error) {
	var (
		claim  Claims
		userID uuid.UUID
	)

	token, err := jwt.ParseWithClaims(tokenString, &claim, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.SecretKey), nil
	})

	if err != nil {
		return userID, err
	}

	if !token.Valid {
		return userID, errors.New("token is not valid")
	}

	userID = claim.UserID
	return userID, nil
}

func (j *jsonWebToken) CreateRegistrationSessionToken(email string, userID *uuid.UUID, verified bool, identityCompleted bool) (string, error) {
	claims := RegistrationSessionClaims{
		Email:             email,
		UserID:            userID,
		Verified:          verified,
		IdentityCompleted: identityCompleted,
		Purpose:           "registration_session",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

func (j *jsonWebToken) ValidateRegistrationSessionToken(tokenString string) (*RegistrationSessionClaims, error) {
	var claims RegistrationSessionClaims

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid session token")
	}
	if claims.Purpose != "registration_session" {
		return nil, errors.New("invalid token purpose")
	}
	if claims.Email == "" || claims.UserID == nil {
		return nil, errors.New("invalid session token claims")
	}

	return &claims, nil
}

func (j *jsonWebToken) GetLoginUser(c *gin.Context) (*entity.User, error) {
	user, ok := c.Get("user")
	if !ok {
		return &entity.User{}, errors.New("failed to get user login")
	}

	return user.(*entity.User), nil
}
