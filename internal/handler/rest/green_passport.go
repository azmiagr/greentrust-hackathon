package rest

import (
	"github.com/gin-gonic/gin"
	"greentrust-hackathon/pkg/helper"
	"greentrust-hackathon/pkg/response"
	"net/http"
)

func (r *Rest) IssueGreenPassport(c *gin.Context) {
	userID := helper.GetAuthenticatedUserID(c)

	result, err := r.service.GreenPassportService.IssueGreenPassport(userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to issue green passport", result)
}
