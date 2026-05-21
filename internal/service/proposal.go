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
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxProposalAttachmentSize = 15 * 1024 * 1024

type IProposalService interface {
	CreateProposal(userID uuid.UUID, param model.CreateProposalParam) (*model.ProposalResponse, error)
	GetProposals(userID uuid.UUID, query model.ProposalListQuery) ([]model.ProposalResponse, error)
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
