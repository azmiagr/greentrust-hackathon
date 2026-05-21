package rest

import (
	"greentrust-hackathon/model"
	"greentrust-hackathon/pkg/helper"
	"greentrust-hackathon/pkg/response"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (r *Rest) GetEvidenceSummary(c *gin.Context) {
	userID := helper.GetAuthenticatedUserID(c)

	result, err := r.service.EvidenceService.GetEvidenceSummary(userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get evidence summary", result)
}

func (r *Rest) GetEvidenceCategoryDetail(c *gin.Context) {
	userID := helper.GetAuthenticatedUserID(c)

	categoryID := c.Param("category_id")
	if categoryID == "" {
		response.Error(c, http.StatusBadRequest, "category_id is required", nil)
		return
	}

	result, err := r.service.EvidenceService.GetEvidenceCategoryDetail(userID, categoryID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get evidence category detail", result)
}

func (r *Rest) UploadEvidenceDocument(c *gin.Context) {
	userID := helper.GetAuthenticatedUserID(c)

	categoryID := c.Param("category_id")
	if categoryID == "" {
		response.Error(c, http.StatusBadRequest, "category_id is required", nil)
		return
	}

	requirementID := strings.TrimSpace(c.PostForm("requirement_id"))
	if requirementID == "" {
		response.Error(c, http.StatusBadRequest, "requirement_id is required", nil)
		return
	}

	aiConfidenceStr := c.PostForm("ai_confidence")
	if aiConfidenceStr == "" {
		response.Error(c, http.StatusBadRequest, "ai_confidence is required", nil)
		return
	}

	aiConfidence, err := strconv.ParseFloat(aiConfidenceStr, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ai_confidence must be a valid float64", err)
		return
	}

	file, err := c.FormFile("document")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "document is required", err)
		return
	}

	result, err := r.service.EvidenceService.UploadEvidenceDocument(userID, model.UploadEvidenceDocumentParam{
		CategoryID:    categoryID,
		RequirementID: requirementID,
		File:          file,
		AIConfidence:  aiConfidence,
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to upload evidence document", result)
}

func (r *Rest) GetEvidenceCategories(c *gin.Context) {
	result, err := r.service.EvidenceService.GetEvidenceCategories()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get evidence categories", result)
}

func (r *Rest) SubmitEvidenceAIReview(c *gin.Context) {
	userID := helper.GetAuthenticatedUserID(c)

	evidenceID, err := uuid.Parse(c.Param("evidence_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "evidence_id must be a valid uuid", err)
		return
	}

	var param model.SubmitEvidenceAIReviewParam
	err = c.ShouldBindJSON(&param)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}
	param.EvidenceID = evidenceID

	result, err := r.service.EvidenceService.SubmitEvidenceAIReview(userID, param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to submit evidence ai review", result)
}

func (r *Rest) GetEvidenceAIReviews(c *gin.Context) {
	userID := helper.GetAuthenticatedUserID(c)

	result, err := r.service.EvidenceService.GetEvidenceAIReviews(userID, c.Query("status"))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get evidence ai reviews", result)
}

func (r *Rest) ReviewEvidenceAI(c *gin.Context) {
	userID := helper.GetAuthenticatedUserID(c)

	reviewID, err := uuid.Parse(c.Param("review_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "review_id must be a valid uuid", err)
		return
	}

	var param model.ReviewEvidenceAIParam
	err = c.ShouldBindJSON(&param)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}
	param.ReviewID = reviewID

	result, err := r.service.EvidenceService.ReviewEvidenceAI(userID, param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to review evidence ai", result)
}
