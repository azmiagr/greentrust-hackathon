package helper

import (
	"greentrust-hackathon/entity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetAuthenticatedUserID(c *gin.Context) uuid.UUID {
	user, exists := c.Get("user")
	if !exists {
		return uuid.Nil
	}

	authenticatedUser, ok := user.(*entity.User)
	if !ok {
		return uuid.Nil
	}

	return authenticatedUser.UserID
}
