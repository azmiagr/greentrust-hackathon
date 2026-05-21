package middleware

import (
	"greentrust-hackathon/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (m *middleware) checkRole(c *gin.Context, allowedRoles []uuid.UUID) bool {
	user, err := m.jwtAuth.GetLoginUser(c)
	if err != nil {
		response.Error(c, http.StatusForbidden, "access denied", err)
		c.Abort()
		return false
	}

	for _, roleID := range allowedRoles {
		if user.RoleID == roleID {
			return true
		}
	}

	response.Error(c, http.StatusForbidden, "access denied", nil)
	c.Abort()
	return false
}
