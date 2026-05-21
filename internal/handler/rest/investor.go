package rest

import (
	"greentrust-hackathon/model"
	"greentrust-hackathon/pkg/helper"
	"greentrust-hackathon/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (r *Rest) CreateInvestorPosition(c *gin.Context) {
	var param model.CreateInvestorPositionParam
	err := c.ShouldBindJSON(&param)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	userID := helper.GetAuthenticatedUserID(c)
	result, err := r.service.InvestorService.CreateInvestorPosition(userID, param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create investor position", result)
}

func (r *Rest) GetInvestorPositions(c *gin.Context) {
	userID := helper.GetAuthenticatedUserID(c)
	result, err := r.service.InvestorService.GetInvestorPositions(userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get investor positions", result)
}

func (r *Rest) UpdateInvestorPosition(c *gin.Context) {
	positionID, err := uuid.Parse(c.Param("position_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "position_id must be a valid uuid", err)
		return
	}

	var param model.UpdateInvestorPositionParam
	err = c.ShouldBindJSON(&param)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	userID := helper.GetAuthenticatedUserID(c)
	result, err := r.service.InvestorService.UpdateInvestorPosition(userID, positionID, param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to update investor position", result)
}

func (r *Rest) DeleteInvestorPosition(c *gin.Context) {
	positionID, err := uuid.Parse(c.Param("position_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "position_id must be a valid uuid", err)
		return
	}

	userID := helper.GetAuthenticatedUserID(c)
	err = r.service.InvestorService.DeleteInvestorPosition(userID, positionID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to delete investor position", nil)
}

func (r *Rest) SearchSkills(c *gin.Context) {
	result, err := r.service.InvestorService.SearchSkills(c.Query("query"))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to search skills", result)
}

func (r *Rest) CreateOnboardingInvestorPosition(c *gin.Context) {
	var param model.CreateInvestorPositionParam
	err := c.ShouldBindJSON(&param)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	result, err := r.service.InvestorService.CreateInvestorPositionWithSession(c.GetHeader("X-Session-Token"), param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create investor position", result)
}

func (r *Rest) GetOnboardingInvestorPositions(c *gin.Context) {
	result, err := r.service.InvestorService.GetInvestorPositionsWithSession(c.GetHeader("X-Session-Token"))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get investor positions", result)
}

func (r *Rest) UpdateOnboardingInvestorPosition(c *gin.Context) {
	positionID, err := uuid.Parse(c.Param("position_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "position_id must be a valid uuid", err)
		return
	}

	var param model.UpdateInvestorPositionParam
	err = c.ShouldBindJSON(&param)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	result, err := r.service.InvestorService.UpdateInvestorPositionWithSession(c.GetHeader("X-Session-Token"), positionID, param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to update investor position", result)
}

func (r *Rest) DeleteOnboardingInvestorPosition(c *gin.Context) {
	positionID, err := uuid.Parse(c.Param("position_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "position_id must be a valid uuid", err)
		return
	}

	err = r.service.InvestorService.DeleteInvestorPositionWithSession(c.GetHeader("X-Session-Token"), positionID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to delete investor position", nil)
}
