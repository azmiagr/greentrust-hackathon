package middleware

import (
	"greentrust-hackathon/model"
	"greentrust-hackathon/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (m *middleware) AuthenticateUser(c *gin.Context) {
	bearer := c.GetHeader("Authorization")
	if bearer == "" {
		response.Error(c, http.StatusUnauthorized, "empty token", nil)
		c.Abort()
		return
	}

	parts := strings.Split(bearer, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		response.Error(c, http.StatusUnauthorized, "invalid authorization header format", nil)
		c.Abort()
		return
	}
	token := parts[1]

	userID, err := m.jwtAuth.ValidateToken(token)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "failed to validate token", err)
		c.Abort()
		return
	}

	user, err := m.service.UserService.GetUser(model.GetUserParam{
		UserID: userID,
	})
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "failed to get user", err)
		c.Abort()
		return
	}

	if user.Status != "active" {
		response.Error(c, http.StatusUnauthorized, "your account has been deactivated. Please contact administrator", nil)
		c.Abort()
		return
	}

	c.Set("user", user)
	c.Next()
}

func (m *middleware) OptionalAuthenticateUser(c *gin.Context) {
	bearer := c.GetHeader("Authorization")
	if bearer == "" {
		c.Next()
		return
	}

	parts := strings.Split(bearer, " ")
	if len(parts) != 2 {
		c.Next()
		return
	}

	token := parts[1]
	userID, err := m.jwtAuth.ValidateToken(token)
	if err != nil {
		c.Next()
		return
	}

	user, err := m.service.UserService.GetUser(model.GetUserParam{
		UserID: userID,
	})
	if err != nil {
		c.Next()
		return
	}

	if user.Status != "active" {
		c.Next()
		return
	}

	c.Set("user", user)
	c.Next()
}
