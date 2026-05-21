package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"greentrust-hackathon/model"
	"greentrust-hackathon/pkg/helper"
	"greentrust-hackathon/pkg/response"
	"net/http"
)

func (r *Rest) GetPublicUMKMDirectory(c *gin.Context) {
	var query model.PublicUMKMDirectoryQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	result, err := r.service.GreenPassportService.GetPublicUMKMDirectory(query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get public umkm directory", result)
}

func (r *Rest) GetPublicUMKMDetail(c *gin.Context) {
	profileID, err := uuid.Parse(c.Param("profile_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "profile_id must be a valid uuid", err)
		return
	}

	result, err := r.service.GreenPassportService.GetPublicUMKMDetail(profileID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get public umkm detail", result)
}

func (r *Rest) IssueGreenPassport(c *gin.Context) {
	userID := helper.GetAuthenticatedUserID(c)

	result, err := r.service.GreenPassportService.IssueGreenPassport(userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to issue green passport", result)
}
