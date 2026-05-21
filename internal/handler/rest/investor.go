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

func (r *Rest) GetPublicInvestors(c *gin.Context) {
	var query model.PublicInvestorDirectoryQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	result, err := r.service.InvestorService.GetPublicInvestorDirectory(query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get public investors", result)
}

func (r *Rest) GetPublicInvestorDetail(c *gin.Context) {
	profileID, err := uuid.Parse(c.Param("profile_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "profile_id must be a valid uuid", err)
		return
	}

	result, err := r.service.InvestorService.GetPublicInvestorDetail(profileID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get public investor detail", result)
}

func (r *Rest) GetInvestorDashboard(c *gin.Context) {
	userID := helper.GetAuthenticatedUserID(c)
	result, err := r.service.InvestorService.GetInvestorDashboard(userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get investor dashboard", result)
}

func (r *Rest) GetInvestorProfile(c *gin.Context) {
	userID := helper.GetAuthenticatedUserID(c)
	result, err := r.service.InvestorService.GetInvestorProfile(userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get investor profile", result)
}

func (r *Rest) GetInvestorPortfolio(c *gin.Context) {
	var query model.InvestorPortfolioQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	userID := helper.GetAuthenticatedUserID(c)
	result, err := r.service.InvestorService.GetInvestorPortfolio(userID, query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get investor portfolio", result)
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

func (r *Rest) SubmitOnboardingInvestorProfile(c *gin.Context) {
	var param model.SubmitInvestorProfileParam
	err := c.ShouldBindJSON(&param)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	sessionToken := c.GetHeader("X-Session-Token")
	result, err := r.service.InvestorService.SubmitInvestorProfileWithSession(sessionToken, param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to submit investor profile", result)
}

func (r *Rest) CreateOnboardingInvestorPosition(c *gin.Context) {
	var param model.CreateInvestorPositionParam
	err := c.ShouldBindJSON(&param)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	sessionToken := c.GetHeader("X-Session-Token")
	result, err := r.service.InvestorService.CreateInvestorPositionWithSession(sessionToken, param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create investor position", result)
}

func (r *Rest) GetOnboardingInvestorPositions(c *gin.Context) {
	sessionToken := c.GetHeader("X-Session-Token")
	result, err := r.service.InvestorService.GetInvestorPositionsWithSession(sessionToken)
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

	sessionToken := c.GetHeader("X-Session-Token")
	result, err := r.service.InvestorService.UpdateInvestorPositionWithSession(sessionToken, positionID, param)
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

	sessionToken := c.GetHeader("X-Session-Token")
	err = r.service.InvestorService.DeleteInvestorPositionWithSession(sessionToken, positionID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to delete investor position", nil)
}
