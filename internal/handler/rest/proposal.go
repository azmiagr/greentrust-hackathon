package rest

import (
	"greentrust-hackathon/model"
	"greentrust-hackathon/pkg/helper"
	"greentrust-hackathon/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (r *Rest) CreateProposal(c *gin.Context) {
	param, err := bindProposalMultipart(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	userID := helper.GetAuthenticatedUserID(c)
	result, err := r.service.ProposalService.CreateProposal(userID, param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to create proposal", result)
}

func (r *Rest) GetProposals(c *gin.Context) {
	userID := helper.GetAuthenticatedUserID(c)

	query := model.ProposalListQuery{
		Box:    c.Query("box"),
		Status: c.Query("status"),
	}

	result, err := r.service.ProposalService.GetProposals(userID, query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get proposals", result)
}

func (r *Rest) GetInvestorProposals(c *gin.Context) {
	var query model.InvestorProposalListQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	userID := helper.GetAuthenticatedUserID(c)
	result, err := r.service.ProposalService.GetInvestorProposals(userID, query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get investor proposals", result)
}

func (r *Rest) GetUMKMProposals(c *gin.Context) {
	var query model.UMKMProposalListQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind query", err)
		return
	}

	userID := helper.GetAuthenticatedUserID(c)
	result, err := r.service.ProposalService.GetUMKMProposals(userID, query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get umkm proposals", result)
}

func (r *Rest) GetProposalDetail(c *gin.Context) {
	proposalID, err := uuid.Parse(c.Param("proposal_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "proposal_id must be a valid uuid", err)
		return
	}

	userID := helper.GetAuthenticatedUserID(c)
	result, err := r.service.ProposalService.GetProposalDetail(userID, proposalID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get proposal detail", result)
}

func (r *Rest) UpdateProposal(c *gin.Context) {
	proposalID, err := uuid.Parse(c.Param("proposal_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "proposal_id must be a valid uuid", err)
		return
	}

	createParam, err := bindProposalMultipart(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	param := model.UpdateProposalParam{
		ReceiverRole:      createParam.ReceiverRole,
		ReceiverProfileID: createParam.ReceiverProfileID,
		ProposalType:      createParam.ProposalType,
		Title:             createParam.Title,
		Amount:            createParam.Amount,
		TenorMonths:       createParam.TenorMonths,
		Scheme:            createParam.Scheme,
		Message:           createParam.Message,
		Action:            createParam.Action,
		Attachments:       createParam.Attachments,
	}

	userID := helper.GetAuthenticatedUserID(c)
	result, err := r.service.ProposalService.UpdateProposal(userID, proposalID, param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to update proposal", result)
}

func (r *Rest) SendProposal(c *gin.Context) {
	r.handleProposalAction(c, "send")
}

func (r *Rest) AcceptProposal(c *gin.Context) {
	r.handleProposalAction(c, "accept")
}

func (r *Rest) RejectProposal(c *gin.Context) {
	r.handleProposalAction(c, "reject")
}

func (r *Rest) WithdrawProposal(c *gin.Context) {
	r.handleProposalAction(c, "withdraw")
}

func (r *Rest) handleProposalAction(c *gin.Context, action string) {
	proposalID, err := uuid.Parse(c.Param("proposal_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "proposal_id must be a valid uuid", err)
		return
	}

	userID := helper.GetAuthenticatedUserID(c)

	var result *model.ProposalResponse
	switch action {
	case "send":
		result, err = r.service.ProposalService.SendProposal(userID, proposalID)
	case "accept":
		result, err = r.service.ProposalService.AcceptProposal(userID, proposalID)
	case "reject":
		result, err = r.service.ProposalService.RejectProposal(userID, proposalID)
	case "withdraw":
		result, err = r.service.ProposalService.WithdrawProposal(userID, proposalID)
	default:
		response.Error(c, http.StatusBadRequest, "invalid proposal action", nil)
		return
	}

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to "+action+" proposal", result)
}

func bindProposalMultipart(c *gin.Context) (model.CreateProposalParam, error) {
	err := c.Request.ParseMultipartForm(32 << 20)
	if err != nil {
		return model.CreateProposalParam{}, err
	}

	amount, err := strconv.ParseInt(c.PostForm("amount"), 10, 64)
	if err != nil {
		return model.CreateProposalParam{}, err
	}

	tenorMonths := 0
	if c.PostForm("tenor_months") != "" {
		tenorMonths, err = strconv.Atoi(c.PostForm("tenor_months"))
		if err != nil {
			return model.CreateProposalParam{}, err
		}
	}

	var attachments = c.Request.MultipartForm.File["attachments"]

	return model.CreateProposalParam{
		ReceiverRole:      c.PostForm("receiver_role"),
		ReceiverProfileID: c.PostForm("receiver_profile_id"),
		ProposalType:      c.PostForm("proposal_type"),
		Title:             c.PostForm("title"),
		Amount:            amount,
		TenorMonths:       tenorMonths,
		Scheme:            c.PostForm("scheme"),
		Message:           c.PostForm("message"),
		Action:            c.PostForm("action"),
		Attachments:       attachments,
	}, nil
}
