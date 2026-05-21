package service

import (
	"errors"
	"greentrust-hackathon/entity"
	"greentrust-hackathon/internal/repository"
	"greentrust-hackathon/model"
	constants "greentrust-hackathon/pkg/constant"
	"greentrust-hackathon/pkg/database/mariadb"
	apperrors "greentrust-hackathon/pkg/errors"
	"greentrust-hackathon/pkg/supabase"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxProposalAttachmentSize = 15 * 1024 * 1024

type IProposalService interface {
	CreateProposal(userID uuid.UUID, param model.CreateProposalParam) (*model.ProposalResponse, error)
	GetProposals(userID uuid.UUID, query model.ProposalListQuery) ([]model.ProposalResponse, error)
	GetInvestorProposals(userID uuid.UUID, query model.InvestorProposalListQuery) (*model.InvestorProposalListResponse, error)
	GetUMKMProposals(userID uuid.UUID, query model.UMKMProposalListQuery) (*model.UMKMProposalListResponse, error)
	GetProposalDetail(userID uuid.UUID, proposalID uuid.UUID) (*model.ProposalResponse, error)
	UpdateProposal(userID uuid.UUID, proposalID uuid.UUID, param model.UpdateProposalParam) (*model.ProposalResponse, error)
	SendProposal(userID uuid.UUID, proposalID uuid.UUID) (*model.ProposalResponse, error)
	AcceptProposal(userID uuid.UUID, proposalID uuid.UUID) (*model.ProposalResponse, error)
	RejectProposal(userID uuid.UUID, proposalID uuid.UUID) (*model.ProposalResponse, error)
	WithdrawProposal(userID uuid.UUID, proposalID uuid.UUID) (*model.ProposalResponse, error)
}

type ProposalService struct {
	db                  *gorm.DB
	userRepo            repository.IUserRepository
	userIdentityRepo    repository.IUserIdentityRepository
	umkmProfileRepo     repository.IUMKMProfileRepository
	investorProfileRepo repository.IInvestorProfileRepository
	proposalRepo        repository.IProposalRepository
	supabase            supabase.Interface
}

func NewProposalService(
	userRepo repository.IUserRepository,
	userIdentityRepo repository.IUserIdentityRepository,
	umkmProfileRepo repository.IUMKMProfileRepository,
	investorProfileRepo repository.IInvestorProfileRepository,
	proposalRepo repository.IProposalRepository,
	supabase supabase.Interface,
) IProposalService {
	return &ProposalService{
		db:                  mariadb.Connection,
		userRepo:            userRepo,
		userIdentityRepo:    userIdentityRepo,
		umkmProfileRepo:     umkmProfileRepo,
		investorProfileRepo: investorProfileRepo,
		proposalRepo:        proposalRepo,
		supabase:            supabase,
	}
}

func (s *ProposalService) CreateProposal(userID uuid.UUID, param model.CreateProposalParam) (*model.ProposalResponse, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	sender, err := s.resolveSender(tx, userID)
	if err != nil {
		return nil, err
	}

	receiver, err := s.resolveReceiver(tx, param.ReceiverRole, param.ReceiverProfileID)
	if err != nil {
		return nil, err
	}

	if sender.Role == receiver.Role {
		return nil, apperrors.BadRequest("proposal must be sent across different roles")
	}

	status := "draft"
	var sentAt *time.Time
	if param.Action == "send" {
		now := time.Now()
		status = "sent"
		sentAt = &now
	}

	proposal := &entity.Proposal{
		ProposalID:        uuid.New(),
		SenderUserID:      sender.UserID,
		ReceiverUserID:    receiver.UserID,
		SenderRole:        sender.Role,
		ReceiverRole:      receiver.Role,
		SenderProfileID:   sender.ProfileID,
		ReceiverProfileID: receiver.ProfileID,
		ProposalType:      param.ProposalType,
		Title:             strings.TrimSpace(param.Title),
		Amount:            param.Amount,
		TenorMonths:       param.TenorMonths,
		Scheme:            strings.TrimSpace(param.Scheme),
		Message:           strings.TrimSpace(param.Message),
		Status:            status,
		SentAt:            sentAt,
	}

	err = s.proposalRepo.CreateProposal(tx, proposal)
	if err != nil {
		return nil, apperrors.InternalServer("failed to create proposal")
	}

	attachments, uploadedURLs, err := s.uploadProposalAttachments(proposal.ProposalID, param.Attachments)
	if err != nil {
		return nil, err
	}

	err = s.proposalRepo.CreateProposalAttachments(tx, attachments)
	if err != nil {
		s.supabase.DeleteMultipleFiles(uploadedURLs)
		return nil, apperrors.InternalServer("failed to save proposal attachments")
	}

	err = tx.Commit().Error
	if err != nil {
		s.supabase.DeleteMultipleFiles(uploadedURLs)
		return nil, apperrors.InternalServer("failed to commit proposal")
	}

	proposal.Attachments = attachments
	return s.buildProposalResponse(*proposal, userID), nil
}

func (s *ProposalService) GetProposals(userID uuid.UUID, query model.ProposalListQuery) ([]model.ProposalResponse, error) {
	proposals, err := s.proposalRepo.GetProposalsByUserID(s.db, userID, query.Box, query.Status)
	if err != nil {
		return nil, apperrors.InternalServer("failed to get proposals")
	}

	responses := make([]model.ProposalResponse, 0, len(proposals))
	for _, proposal := range proposals {
		responses = append(responses, *s.buildProposalResponse(*proposal, userID))
	}

	return responses, nil
}

func (s *ProposalService) GetInvestorProposals(userID uuid.UUID, query model.InvestorProposalListQuery) (*model.InvestorProposalListResponse, error) {
	user, err := s.userRepo.GetUser(s.db, model.GetUserParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.InternalServer("failed to get user")
	}
	if user.RoleID != constants.RoleInvestor {
		return nil, apperrors.Forbidden("investor role required")
	}

	proposals, err := s.proposalRepo.GetProposalsByUserID(s.db, userID, "", "")
	if err != nil {
		return nil, apperrors.InternalServer("failed to get investor proposals")
	}

	tab := strings.TrimSpace(query.Tab)
	if tab == "" {
		tab = "sent"
	}
	if tab != "sent" && tab != "requests" && tab != "approved" && tab != "rejected" {
		return nil, apperrors.BadRequest("tab must be sent, requests, approved, or rejected")
	}

	items := make([]model.InvestorProposalItem, 0, len(proposals))
	for _, proposal := range proposals {
		if !matchesInvestorProposalTab(*proposal, userID, tab) {
			continue
		}

		items = append(items, s.buildInvestorProposalItem(*proposal, userID))
	}

	return &model.InvestorProposalListResponse{
		Summary: buildInvestorProposalSummary(proposals, userID),
		Tabs:    buildInvestorProposalTabs(proposals, userID),
		Items:   items,
	}, nil
}

func (s *ProposalService) GetUMKMProposals(userID uuid.UUID, query model.UMKMProposalListQuery) (*model.UMKMProposalListResponse, error) {
	user, err := s.userRepo.GetUser(s.db, model.GetUserParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.InternalServer("failed to get user")
	}
	if user.RoleID != constants.RoleUMKM {
		return nil, apperrors.Forbidden("umkm role required")
	}

	proposals, err := s.proposalRepo.GetProposalsByUserID(s.db, userID, "", "")
	if err != nil {
		return nil, apperrors.InternalServer("failed to get umkm proposals")
	}

	tab := strings.TrimSpace(query.Tab)
	if tab == "" {
		tab = "incoming"
	}
	if tab != "incoming" && tab != "sent" && tab != "approved" && tab != "rejected" && tab != "all" {
		return nil, apperrors.BadRequest("tab must be incoming, sent, approved, rejected, or all")
	}

	sortOrder := strings.TrimSpace(query.Sort)
	if sortOrder == "" {
		sortOrder = "newest"
	}
	if sortOrder != "newest" && sortOrder != "oldest" {
		return nil, apperrors.BadRequest("sort must be newest or oldest")
	}

	items := make([]model.UMKMProposalItem, 0, len(proposals))
	for _, proposal := range proposals {
		if !matchesUMKMProposalTab(*proposal, userID, tab) {
			continue
		}

		items = append(items, s.buildUMKMProposalItem(*proposal, userID))
	}

	if sortOrder == "oldest" {
		reverseUMKMProposalItems(items)
	}

	return &model.UMKMProposalListResponse{
		Summary: buildUMKMProposalSummary(proposals, userID),
		Tabs:    buildUMKMProposalTabs(proposals, userID),
		Items:   items,
	}, nil
}

func (s *ProposalService) GetProposalDetail(userID uuid.UUID, proposalID uuid.UUID) (*model.ProposalResponse, error) {
	proposal, err := s.proposalRepo.GetProposalByIDAndUserID(s.db, proposalID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("proposal not found")
		}
		return nil, apperrors.InternalServer("failed to get proposal")
	}

	return s.buildProposalResponse(*proposal, userID), nil
}

func (s *ProposalService) UpdateProposal(userID uuid.UUID, proposalID uuid.UUID, param model.UpdateProposalParam) (*model.ProposalResponse, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	proposal, err := s.proposalRepo.GetProposalByIDAndUserID(tx, proposalID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("proposal not found")
		}
		return nil, apperrors.InternalServer("failed to get proposal")
	}

	if proposal.SenderUserID != userID {
		return nil, apperrors.Forbidden("only sender can update proposal")
	}
	if proposal.Status != "draft" {
		return nil, apperrors.BadRequest("only draft proposal can be updated")
	}

	receiver, err := s.resolveReceiver(tx, param.ReceiverRole, param.ReceiverProfileID)
	if err != nil {
		return nil, err
	}
	if proposal.SenderRole == receiver.Role {
		return nil, apperrors.BadRequest("proposal must be sent across different roles")
	}

	status := "draft"
	var sentAt *time.Time
	if param.Action == "send" {
		now := time.Now()
		status = "sent"
		sentAt = &now
	}

	proposal.ReceiverUserID = receiver.UserID
	proposal.ReceiverRole = receiver.Role
	proposal.ReceiverProfileID = receiver.ProfileID
	proposal.ProposalType = param.ProposalType
	proposal.Title = strings.TrimSpace(param.Title)
	proposal.Amount = param.Amount
	proposal.TenorMonths = param.TenorMonths
	proposal.Scheme = strings.TrimSpace(param.Scheme)
	proposal.Message = strings.TrimSpace(param.Message)
	proposal.Status = status
	proposal.SentAt = sentAt

	err = s.proposalRepo.UpdateProposal(tx, proposal)
	if err != nil {
		return nil, apperrors.InternalServer("failed to update proposal")
	}

	oldURLs := extractProposalAttachmentURLs(proposal.Attachments)
	attachments, uploadedURLs, err := s.uploadProposalAttachments(proposal.ProposalID, param.Attachments)
	if err != nil {
		return nil, err
	}

	if len(param.Attachments) > 0 {
		err = s.proposalRepo.DeleteAttachmentsByProposalID(tx, proposal.ProposalID)
		if err != nil {
			s.supabase.DeleteMultipleFiles(uploadedURLs)
			return nil, apperrors.InternalServer("failed to replace proposal attachments")
		}
		err = s.proposalRepo.CreateProposalAttachments(tx, attachments)
		if err != nil {
			s.supabase.DeleteMultipleFiles(uploadedURLs)
			return nil, apperrors.InternalServer("failed to save proposal attachments")
		}
		proposal.Attachments = attachments
	}

	err = tx.Commit().Error
	if err != nil {
		s.supabase.DeleteMultipleFiles(uploadedURLs)
		return nil, apperrors.InternalServer("failed to commit proposal")
	}

	if len(param.Attachments) > 0 {
		s.supabase.DeleteMultipleFiles(oldURLs)
	}

	return s.buildProposalResponse(*proposal, userID), nil
}

func (s *ProposalService) SendProposal(userID uuid.UUID, proposalID uuid.UUID) (*model.ProposalResponse, error) {
	return s.transitionProposal(userID, proposalID, "send")
}

func (s *ProposalService) AcceptProposal(userID uuid.UUID, proposalID uuid.UUID) (*model.ProposalResponse, error) {
	return s.transitionProposal(userID, proposalID, "accept")
}

func (s *ProposalService) RejectProposal(userID uuid.UUID, proposalID uuid.UUID) (*model.ProposalResponse, error) {
	return s.transitionProposal(userID, proposalID, "reject")
}

func (s *ProposalService) WithdrawProposal(userID uuid.UUID, proposalID uuid.UUID) (*model.ProposalResponse, error) {
	return s.transitionProposal(userID, proposalID, "withdraw")
}

func (s *ProposalService) transitionProposal(userID uuid.UUID, proposalID uuid.UUID, action string) (*model.ProposalResponse, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, apperrors.InternalServer("failed to start transaction")
	}
	defer tx.Rollback()

	proposal, err := s.proposalRepo.GetProposalByIDAndUserID(tx, proposalID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("proposal not found")
		}
		return nil, apperrors.InternalServer("failed to get proposal")
	}

	now := time.Now()
	switch action {
	case "send":
		if proposal.SenderUserID != userID {
			return nil, apperrors.Forbidden("only sender can send proposal")
		}
		if proposal.Status != "draft" {
			return nil, apperrors.BadRequest("only draft proposal can be sent")
		}
		proposal.Status = "sent"
		proposal.SentAt = &now
	case "accept":
		if proposal.ReceiverUserID != userID {
			return nil, apperrors.Forbidden("only receiver can accept proposal")
		}
		if proposal.Status != "sent" {
			return nil, apperrors.BadRequest("only sent proposal can be accepted")
		}
		proposal.Status = "accepted"
		proposal.AcceptedAt = &now
	case "reject":
		if proposal.ReceiverUserID != userID {
			return nil, apperrors.Forbidden("only receiver can reject proposal")
		}
		if proposal.Status != "sent" {
			return nil, apperrors.BadRequest("only sent proposal can be rejected")
		}
		proposal.Status = "rejected"
		proposal.RejectedAt = &now
	case "withdraw":
		if proposal.SenderUserID != userID {
			return nil, apperrors.Forbidden("only sender can withdraw proposal")
		}
		if proposal.Status != "sent" {
			return nil, apperrors.BadRequest("only sent proposal can be withdrawn")
		}
		proposal.Status = "withdrawn"
	default:
		return nil, apperrors.BadRequest("invalid proposal action")
	}

	if err := s.proposalRepo.UpdateProposal(tx, proposal); err != nil {
		return nil, apperrors.InternalServer("failed to update proposal status")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, apperrors.InternalServer("failed to commit proposal status")
	}

	return s.buildProposalResponse(*proposal, userID), nil
}

type proposalParty struct {
	UserID    uuid.UUID
	Role      string
	ProfileID uuid.UUID
	Name      string
	Subtitle  string
}

func (s *ProposalService) resolveSender(tx *gorm.DB, userID uuid.UUID) (*proposalParty, error) {
	user, err := s.userRepo.GetUser(tx, model.GetUserParam{UserID: userID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.InternalServer("failed to get user")
	}

	if user.RoleID == constants.RoleUMKM {
		profile, err := s.umkmProfileRepo.GetUMKMProfile(tx, model.GetUMKMProfileParam{UserID: userID})
		if err != nil {
			return nil, apperrors.Forbidden("business profile must be completed first")
		}
		return &proposalParty{UserID: user.UserID, Role: "umkm", ProfileID: profile.ProfileID, Name: profile.BusinessName, Subtitle: profile.BusinessCity}, nil
	}

	if user.RoleID == constants.RoleInvestor {
		profile, err := s.investorProfileRepo.GetInvestorProfile(tx, model.GetInvestorProfileParam{UserID: userID})
		if err != nil {
			return nil, apperrors.Forbidden("investor profile must be completed first")
		}

		name, subtitle := s.getInvestorDisplay(tx, user.UserID)
		if name == "" {
			name = user.Email
		}
		return &proposalParty{UserID: user.UserID, Role: "investor", ProfileID: profile.ProfileID, Name: name, Subtitle: subtitle}, nil
	}

	return nil, apperrors.Forbidden("unsupported user role")
}

func (s *ProposalService) resolveReceiver(tx *gorm.DB, receiverRole string, receiverProfileID string) (*proposalParty, error) {
	profileID, err := uuid.Parse(receiverProfileID)
	if err != nil {
		return nil, apperrors.BadRequest("receiver_profile_id must be a valid uuid")
	}

	switch receiverRole {
	case "umkm":
		profile, err := s.umkmProfileRepo.GetUMKMProfile(tx, model.GetUMKMProfileParam{ProfileID: profileID})
		if err != nil {
			return nil, apperrors.NotFound("receiver business profile not found")
		}
		return &proposalParty{UserID: profile.UserID, Role: "umkm", ProfileID: profile.ProfileID, Name: profile.BusinessName, Subtitle: profile.BusinessCity}, nil
	case "investor":
		profile, err := s.investorProfileRepo.GetInvestorProfile(tx, model.GetInvestorProfileParam{ProfileID: profileID})
		if err != nil {
			return nil, apperrors.NotFound("receiver investor profile not found")
		}

		name, subtitle := s.getInvestorDisplay(tx, profile.UserID)
		return &proposalParty{UserID: profile.UserID, Role: "investor", ProfileID: profile.ProfileID, Name: name, Subtitle: subtitle}, nil
	default:
		return nil, apperrors.BadRequest("receiver_role must be umkm or investor")
	}
}

func (s *ProposalService) uploadProposalAttachments(proposalID uuid.UUID, files []*multipart.FileHeader) ([]entity.ProposalAttachment, []string, error) {
	attachments := make([]entity.ProposalAttachment, 0, len(files))
	uploadedURLs := make([]string, 0, len(files))

	for _, file := range files {
		if file == nil {
			continue
		}
		if file.Size > maxProposalAttachmentSize {
			s.supabase.DeleteMultipleFiles(uploadedURLs)
			return nil, nil, apperrors.BadRequest("attachment must not exceed 15MB")
		}

		url, err := s.supabase.UploadPDF(file)
		if err != nil {
			s.supabase.DeleteMultipleFiles(uploadedURLs)
			return nil, nil, apperrors.BadRequest("attachment must be a valid PDF")
		}

		uploadedURLs = append(uploadedURLs, url)
		attachments = append(attachments, entity.ProposalAttachment{
			AttachmentID: uuid.New(),
			ProposalID:   proposalID,
			FilePath:     url,
			OriginalName: file.Filename,
			MimeType:     file.Header.Get("Content-Type"),
			FileSize:     file.Size,
		})
	}

	return attachments, uploadedURLs, nil
}

func extractProposalAttachmentURLs(attachments []entity.ProposalAttachment) []string {
	urls := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		urls = append(urls, attachment.FilePath)
	}
	return urls
}

func (s *ProposalService) buildProposalResponse(proposal entity.Proposal, viewerUserID uuid.UUID) *model.ProposalResponse {
	counterparty := model.ProposalCounterpartyResponse{}

	if viewerUserID == proposal.SenderUserID {
		counterparty = model.ProposalCounterpartyResponse{
			UserID:    proposal.ReceiverUserID,
			Role:      proposal.ReceiverRole,
			ProfileID: proposal.ReceiverProfileID,
		}
	} else {
		counterparty = model.ProposalCounterpartyResponse{
			UserID:    proposal.SenderUserID,
			Role:      proposal.SenderRole,
			ProfileID: proposal.SenderProfileID,
		}
	}
	counterparty.Name, counterparty.Subtitle = s.getCounterpartyDisplay(s.db, counterparty)

	attachments := make([]model.ProposalAttachmentResponse, 0, len(proposal.Attachments))
	for _, attachment := range proposal.Attachments {
		attachments = append(attachments, model.ProposalAttachmentResponse{
			AttachmentID: attachment.AttachmentID,
			FilePath:     attachment.FilePath,
			OriginalName: attachment.OriginalName,
			MimeType:     attachment.MimeType,
			FileSize:     attachment.FileSize,
		})
	}

	return &model.ProposalResponse{
		ProposalID:        proposal.ProposalID,
		SenderUserID:      proposal.SenderUserID,
		ReceiverUserID:    proposal.ReceiverUserID,
		SenderRole:        proposal.SenderRole,
		ReceiverRole:      proposal.ReceiverRole,
		SenderProfileID:   proposal.SenderProfileID,
		ReceiverProfileID: proposal.ReceiverProfileID,
		ProposalType:      proposal.ProposalType,
		Title:             proposal.Title,
		Amount:            proposal.Amount,
		TenorMonths:       proposal.TenorMonths,
		Scheme:            proposal.Scheme,
		Message:           proposal.Message,
		Status:            proposal.Status,
		SentAt:            proposal.SentAt,
		AcceptedAt:        proposal.AcceptedAt,
		RejectedAt:        proposal.RejectedAt,
		CreatedAt:         proposal.CreatedAt,
		UpdatedAt:         proposal.UpdatedAt,
		Attachments:       attachments,
		Counterparty:      counterparty,
	}
}

func (s *ProposalService) buildInvestorProposalItem(proposal entity.Proposal, viewerUserID uuid.UUID) model.InvestorProposalItem {
	base := s.buildProposalResponse(proposal, viewerUserID)
	direction := "incoming"
	if proposal.SenderUserID == viewerUserID {
		direction = "sent"
	}

	return model.InvestorProposalItem{
		ProposalID:        proposal.ProposalID,
		ProposalCode:      buildProposalCode(proposal),
		Direction:         direction,
		Counterparty:      base.Counterparty,
		ProposalType:      proposal.ProposalType,
		ProposalTypeLabel: buildProposalTypeLabel(proposal.ProposalType),
		Title:             proposal.Title,
		Amount:            proposal.Amount,
		AmountLabel:       formatRupiahCompact(proposal.Amount),
		TenorMonths:       proposal.TenorMonths,
		Scheme:            proposal.Scheme,
		Message:           proposal.Message,
		Status:            proposal.Status,
		StatusLabel:       buildInvestorProposalStatusLabel(proposal, viewerUserID),
		StatusTone:        buildInvestorProposalStatusTone(proposal.Status),
		SentAt:            proposal.SentAt,
		AcceptedAt:        proposal.AcceptedAt,
		RejectedAt:        proposal.RejectedAt,
		CreatedAt:         proposal.CreatedAt,
		UpdatedAt:         proposal.UpdatedAt,
		Attachments:       base.Attachments,
		CanEdit:           proposal.SenderUserID == viewerUserID && proposal.Status == "draft",
		CanWithdraw:       proposal.SenderUserID == viewerUserID && proposal.Status == "sent",
		CanAccept:         proposal.ReceiverUserID == viewerUserID && proposal.Status == "sent",
		CanReject:         proposal.ReceiverUserID == viewerUserID && proposal.Status == "sent",
	}
}

func (s *ProposalService) buildUMKMProposalItem(proposal entity.Proposal, viewerUserID uuid.UUID) model.UMKMProposalItem {
	base := s.buildProposalResponse(proposal, viewerUserID)
	direction := "incoming"
	if proposal.SenderUserID == viewerUserID {
		direction = "sent"
	}

	return model.UMKMProposalItem{
		ProposalID:        proposal.ProposalID,
		ProposalCode:      buildProposalCode(proposal),
		Direction:         direction,
		Counterparty:      base.Counterparty,
		ProposalType:      proposal.ProposalType,
		ProposalTypeLabel: buildProposalTypeLabel(proposal.ProposalType),
		Title:             proposal.Title,
		Amount:            proposal.Amount,
		AmountLabel:       formatRupiahCompact(proposal.Amount),
		TenorMonths:       proposal.TenorMonths,
		Scheme:            proposal.Scheme,
		Message:           proposal.Message,
		Status:            proposal.Status,
		StatusLabel:       buildUMKMProposalStatusLabel(proposal, viewerUserID),
		StatusTone:        buildInvestorProposalStatusTone(proposal.Status),
		SentAt:            proposal.SentAt,
		AcceptedAt:        proposal.AcceptedAt,
		RejectedAt:        proposal.RejectedAt,
		CreatedAt:         proposal.CreatedAt,
		UpdatedAt:         proposal.UpdatedAt,
		Attachments:       base.Attachments,
		CanReview:         proposal.ReceiverUserID == viewerUserID && proposal.Status == "sent",
		CanEdit:           proposal.SenderUserID == viewerUserID && proposal.Status == "draft",
		CanWithdraw:       proposal.SenderUserID == viewerUserID && proposal.Status == "sent",
		CanAccept:         proposal.ReceiverUserID == viewerUserID && proposal.Status == "sent",
		CanReject:         proposal.ReceiverUserID == viewerUserID && proposal.Status == "sent",
	}
}

func matchesInvestorProposalTab(proposal entity.Proposal, viewerUserID uuid.UUID, tab string) bool {
	switch tab {
	case "sent":
		return proposal.SenderUserID == viewerUserID
	case "requests":
		return proposal.ReceiverUserID == viewerUserID && proposal.SenderRole == "umkm" && proposal.Status == "sent"
	case "approved":
		return proposal.Status == "accepted"
	case "rejected":
		return proposal.Status == "rejected"
	default:
		return false
	}
}

func matchesUMKMProposalTab(proposal entity.Proposal, viewerUserID uuid.UUID, tab string) bool {
	switch tab {
	case "incoming":
		return proposal.ReceiverUserID == viewerUserID && proposal.SenderRole == "investor" && proposal.Status == "sent"
	case "sent":
		return proposal.SenderUserID == viewerUserID
	case "approved":
		return proposal.Status == "accepted"
	case "rejected":
		return proposal.Status == "rejected"
	case "all":
		return true
	default:
		return false
	}
}

func buildInvestorProposalSummary(proposals []*entity.Proposal, viewerUserID uuid.UUID) model.InvestorProposalSummary {
	approvedUMKMs := map[uuid.UUID]bool{}
	activeCount := 0
	incomingRequestsCount := 0
	acceptedCount := 0
	rejectedCount := 0
	approvedTotalValue := int64(0)
	approvedSince := time.Now().AddDate(-1, 0, 0)

	for _, proposal := range proposals {
		if proposal.Status == "sent" {
			activeCount++
		}
		if proposal.ReceiverUserID == viewerUserID && proposal.SenderRole == "umkm" && proposal.Status == "sent" {
			incomingRequestsCount++
		}
		if proposal.Status == "accepted" {
			acceptedCount++
			if proposal.SenderRole == "umkm" {
				approvedUMKMs[proposal.SenderProfileID] = true
			}
			if proposal.ReceiverRole == "umkm" {
				approvedUMKMs[proposal.ReceiverProfileID] = true
			}

			approvedAt := proposal.UpdatedAt
			if proposal.AcceptedAt != nil {
				approvedAt = *proposal.AcceptedAt
			}
			if !approvedAt.Before(approvedSince) {
				approvedTotalValue += proposal.Amount
			}
		}
		if proposal.Status == "rejected" {
			rejectedCount++
		}
	}

	approvalRate := 0
	decidedCount := acceptedCount + rejectedCount
	if decidedCount > 0 {
		approvalRate = ((acceptedCount * 100) + (decidedCount / 2)) / decidedCount
	}

	return model.InvestorProposalSummary{
		ActiveProposalsCount:      activeCount,
		ApprovedUMKMCount:         len(approvedUMKMs),
		ApprovalRate:              approvalRate,
		IncomingUMKMRequestsCount: incomingRequestsCount,
		ApprovedTotalValue:        approvedTotalValue,
		ApprovedTotalValueLabel:   formatRupiahCompact(approvedTotalValue),
		ApprovedTotalPeriodLabel:  "12 bulan terakhir",
	}
}

func buildUMKMProposalSummary(proposals []*entity.Proposal, viewerUserID uuid.UUID) model.UMKMProposalSummary {
	incomingOffersCount := 0
	sentProposalsCount := 0
	sentRejectedCount := 0
	sentApprovedCount := 0
	pendingTotalValue := int64(0)
	respondedCount := 0
	responseDurationHours := 0.0

	for _, proposal := range proposals {
		if proposal.ReceiverUserID == viewerUserID && proposal.SenderRole == "investor" && proposal.Status == "sent" {
			incomingOffersCount++
			pendingTotalValue += proposal.Amount
		}
		if proposal.SenderUserID == viewerUserID {
			sentProposalsCount++
			if proposal.Status == "rejected" {
				sentRejectedCount++
			}
			if proposal.Status == "accepted" {
				sentApprovedCount++
			}
		}

		respondedAt := proposal.AcceptedAt
		if proposal.RejectedAt != nil {
			respondedAt = proposal.RejectedAt
		}
		if proposal.ReceiverUserID == viewerUserID && proposal.SentAt != nil && respondedAt != nil {
			respondedCount++
			responseDurationHours += respondedAt.Sub(*proposal.SentAt).Hours()
		}
	}

	averageResponseTimeDays := 0
	if respondedCount > 0 {
		averageResponseTimeDays = int((responseDurationHours / float64(respondedCount) / 24) + 0.5)
	}

	return model.UMKMProposalSummary{
		IncomingOffersCount:     incomingOffersCount,
		SentProposalsCount:      sentProposalsCount,
		SentRejectedCount:       sentRejectedCount,
		SentApprovedCount:       sentApprovedCount,
		PendingTotalValue:       pendingTotalValue,
		PendingTotalValueLabel:  formatRupiahCompact(pendingTotalValue),
		ResponseRate:            buildUMKMResponseRate(proposals, viewerUserID),
		AverageResponseTimeDays: averageResponseTimeDays,
	}
}

func buildInvestorProposalTabs(proposals []*entity.Proposal, viewerUserID uuid.UUID) model.InvestorProposalTabs {
	tabs := model.InvestorProposalTabs{}
	for _, proposal := range proposals {
		if proposal.SenderUserID == viewerUserID {
			tabs.Sent++
		}
		if proposal.ReceiverUserID == viewerUserID && proposal.SenderRole == "umkm" && proposal.Status == "sent" {
			tabs.Requests++
		}
		if proposal.Status == "accepted" {
			tabs.Approved++
		}
		if proposal.Status == "rejected" {
			tabs.Rejected++
		}
	}

	return tabs
}

func buildUMKMProposalTabs(proposals []*entity.Proposal, viewerUserID uuid.UUID) model.UMKMProposalTabs {
	tabs := model.UMKMProposalTabs{}
	for _, proposal := range proposals {
		tabs.All++
		if proposal.ReceiverUserID == viewerUserID && proposal.SenderRole == "investor" && proposal.Status == "sent" {
			tabs.Incoming++
		}
		if proposal.SenderUserID == viewerUserID {
			tabs.Sent++
		}
		if proposal.Status == "accepted" {
			tabs.Approved++
		}
		if proposal.Status == "rejected" {
			tabs.Rejected++
		}
	}

	return tabs
}

func buildUMKMResponseRate(proposals []*entity.Proposal, viewerUserID uuid.UUID) int {
	incomingCount := 0
	respondedCount := 0
	for _, proposal := range proposals {
		if proposal.ReceiverUserID != viewerUserID || proposal.SenderRole != "investor" {
			continue
		}

		incomingCount++
		if proposal.Status == "accepted" || proposal.Status == "rejected" {
			respondedCount++
		}
	}

	if incomingCount == 0 {
		return 0
	}

	return ((respondedCount * 100) + (incomingCount / 2)) / incomingCount
}

func buildProposalCode(proposal entity.Proposal) string {
	id := strings.ReplaceAll(proposal.ProposalID.String(), "-", "")
	suffix := id
	if len(suffix) > 4 {
		suffix = suffix[:4]
	}

	return "PRP-" + strconv.Itoa(proposal.CreatedAt.Year()) + "-" + strings.ToUpper(suffix)
}

func buildProposalTypeLabel(proposalType string) string {
	switch proposalType {
	case "funding":
		return "Pendanaan"
	case "supply":
		return "Pengadaan"
	default:
		return proposalType
	}
}

func buildUMKMProposalStatusLabel(proposal entity.Proposal, viewerUserID uuid.UUID) string {
	switch proposal.Status {
	case "draft":
		return "Draft"
	case "sent":
		if proposal.ReceiverUserID == viewerUserID {
			return "Menunggu Anda"
		}
		return "Sedang Ditinjau"
	case "accepted":
		return "Disetujui"
	case "rejected":
		return "Ditolak"
	case "withdrawn":
		return "Ditarik"
	default:
		return proposal.Status
	}
}

func reverseUMKMProposalItems(items []model.UMKMProposalItem) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}

func buildInvestorProposalStatusLabel(proposal entity.Proposal, viewerUserID uuid.UUID) string {
	switch proposal.Status {
	case "draft":
		return "Draft"
	case "sent":
		if proposal.ReceiverUserID == viewerUserID {
			return "Butuh Tinjauan"
		}
		return "Sedang Ditinjau"
	case "accepted":
		return "Disetujui"
	case "rejected":
		return "Ditolak"
	case "withdrawn":
		return "Ditarik"
	default:
		return proposal.Status
	}
}

func buildInvestorProposalStatusTone(status string) string {
	switch status {
	case "accepted":
		return "success"
	case "rejected", "withdrawn":
		return "danger"
	case "sent":
		return "warning"
	case "draft":
		return "muted"
	default:
		return "neutral"
	}
}

func (s *ProposalService) getCounterpartyDisplay(tx *gorm.DB, counterparty model.ProposalCounterpartyResponse) (string, string) {
	switch counterparty.Role {
	case "umkm":
		profile, err := s.umkmProfileRepo.GetUMKMProfile(tx, model.GetUMKMProfileParam{ProfileID: counterparty.ProfileID})
		if err != nil {
			return "", ""
		}
		return profile.BusinessName, profile.BusinessCity
	case "investor":
		name, subtitle := s.getInvestorDisplay(tx, counterparty.UserID)
		if name != "" {
			return name, subtitle
		}

		user, err := s.userRepo.GetUser(tx, model.GetUserParam{UserID: counterparty.UserID})
		if err != nil {
			return "", subtitle
		}
		return user.Email, subtitle
	default:
		return "", ""
	}
}

func (s *ProposalService) getInvestorDisplay(tx *gorm.DB, userID uuid.UUID) (string, string) {
	subtitle := "Investor"

	identity, err := s.userIdentityRepo.GetUserIdentity(tx, model.GetUserIdentityParam{UserID: userID})
	if err != nil {
		return "", subtitle
	}

	name := strings.TrimSpace(strings.Join([]string{identity.FirstName, identity.LastName}, " "))
	return name, subtitle
}
