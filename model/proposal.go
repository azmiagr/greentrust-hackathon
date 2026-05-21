package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type CreateProposalParam struct {
	ReceiverRole      string                  `form:"receiver_role" binding:"required,oneof=umkm investor"`
	ReceiverProfileID string                  `form:"receiver_profile_id" binding:"required"`
	ProposalType      string                  `form:"proposal_type" binding:"required,oneof=funding supply"`
	Title             string                  `form:"title" binding:"required,max=200"`
	Amount            int64                   `form:"amount" binding:"required,min=1"`
	TenorMonths       int                     `form:"tenor_months" binding:"min=0"`
	Scheme            string                  `form:"scheme" binding:"max=100"`
	Message           string                  `form:"message" binding:"max=1000"`
	Action            string                  `form:"action" binding:"required,oneof=draft send"`
	Attachments       []*multipart.FileHeader `form:"attachments"`
}

type UpdateProposalParam struct {
	ReceiverRole      string                  `form:"receiver_role" binding:"required,oneof=umkm investor"`
	ReceiverProfileID string                  `form:"receiver_profile_id" binding:"required"`
	ProposalType      string                  `form:"proposal_type" binding:"required,oneof=funding supply"`
	Title             string                  `form:"title" binding:"required,max=200"`
	Amount            int64                   `form:"amount" binding:"required,min=1"`
	TenorMonths       int                     `form:"tenor_months" binding:"min=0"`
	Scheme            string                  `form:"scheme" binding:"max=100"`
	Message           string                  `form:"message" binding:"max=1000"`
	Action            string                  `form:"action" binding:"required,oneof=draft send"`
	Attachments       []*multipart.FileHeader `form:"attachments"`
}

type ProposalListQuery struct {
	Box    string `form:"box"`    // inbox, sent
	Status string `form:"status"` // draft, sent, accepted, rejected, withdrawn
}

type InvestorProposalListQuery struct {
	Tab string `form:"tab"` // sent, requests, approved, rejected
}

type UMKMProposalListQuery struct {
	Tab  string `form:"tab"`  // incoming, sent, approved, rejected, all
	Sort string `form:"sort"` // newest, oldest
}

type ProposalActionParam struct {
	Reason string `json:"reason"`
}

type InvestorProposalListResponse struct {
	Summary InvestorProposalSummary `json:"summary"`
	Tabs    InvestorProposalTabs    `json:"tabs"`
	Items   []InvestorProposalItem  `json:"items"`
}

type InvestorProposalSummary struct {
	ActiveProposalsCount      int    `json:"active_proposals_count"`
	ApprovedUMKMCount         int    `json:"approved_umkm_count"`
	ApprovalRate              int    `json:"approval_rate"`
	IncomingUMKMRequestsCount int    `json:"incoming_umkm_requests_count"`
	ApprovedTotalValue        int64  `json:"approved_total_value"`
	ApprovedTotalValueLabel   string `json:"approved_total_value_label"`
	ApprovedTotalPeriodLabel  string `json:"approved_total_period_label"`
}

type InvestorProposalTabs struct {
	Sent     int `json:"sent"`
	Requests int `json:"requests"`
	Approved int `json:"approved"`
	Rejected int `json:"rejected"`
}

type InvestorProposalItem struct {
	ProposalID        uuid.UUID                    `json:"proposal_id"`
	ProposalCode      string                       `json:"proposal_code"`
	Direction         string                       `json:"direction"`
	Counterparty      ProposalCounterpartyResponse `json:"counterparty"`
	ProposalType      string                       `json:"proposal_type"`
	ProposalTypeLabel string                       `json:"proposal_type_label"`
	Title             string                       `json:"title"`
	Amount            int64                        `json:"amount"`
	AmountLabel       string                       `json:"amount_label"`
	TenorMonths       int                          `json:"tenor_months"`
	Scheme            string                       `json:"scheme"`
	Message           string                       `json:"message"`
	Status            string                       `json:"status"`
	StatusLabel       string                       `json:"status_label"`
	StatusTone        string                       `json:"status_tone"`
	SentAt            *time.Time                   `json:"sent_at"`
	AcceptedAt        *time.Time                   `json:"accepted_at"`
	RejectedAt        *time.Time                   `json:"rejected_at"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
	Attachments       []ProposalAttachmentResponse `json:"attachments"`
	CanEdit           bool                         `json:"can_edit"`
	CanWithdraw       bool                         `json:"can_withdraw"`
	CanAccept         bool                         `json:"can_accept"`
	CanReject         bool                         `json:"can_reject"`
}

type UMKMProposalListResponse struct {
	Summary UMKMProposalSummary `json:"summary"`
	Tabs    UMKMProposalTabs    `json:"tabs"`
	Items   []UMKMProposalItem  `json:"items"`
}

type UMKMProposalSummary struct {
	IncomingOffersCount     int    `json:"incoming_offers_count"`
	SentProposalsCount      int    `json:"sent_proposals_count"`
	SentRejectedCount       int    `json:"sent_rejected_count"`
	SentApprovedCount       int    `json:"sent_approved_count"`
	PendingTotalValue       int64  `json:"pending_total_value"`
	PendingTotalValueLabel  string `json:"pending_total_value_label"`
	ResponseRate            int    `json:"response_rate"`
	AverageResponseTimeDays int    `json:"average_response_time_days"`
}

type UMKMProposalTabs struct {
	Incoming int `json:"incoming"`
	Sent     int `json:"sent"`
	Approved int `json:"approved"`
	Rejected int `json:"rejected"`
	All      int `json:"all"`
}

type UMKMProposalItem struct {
	ProposalID        uuid.UUID                    `json:"proposal_id"`
	ProposalCode      string                       `json:"proposal_code"`
	Direction         string                       `json:"direction"`
	Counterparty      ProposalCounterpartyResponse `json:"counterparty"`
	ProposalType      string                       `json:"proposal_type"`
	ProposalTypeLabel string                       `json:"proposal_type_label"`
	Title             string                       `json:"title"`
	Amount            int64                        `json:"amount"`
	AmountLabel       string                       `json:"amount_label"`
	TenorMonths       int                          `json:"tenor_months"`
	Scheme            string                       `json:"scheme"`
	Message           string                       `json:"message"`
	Status            string                       `json:"status"`
	StatusLabel       string                       `json:"status_label"`
	StatusTone        string                       `json:"status_tone"`
	SentAt            *time.Time                   `json:"sent_at"`
	AcceptedAt        *time.Time                   `json:"accepted_at"`
	RejectedAt        *time.Time                   `json:"rejected_at"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
	Attachments       []ProposalAttachmentResponse `json:"attachments"`
	CanReview         bool                         `json:"can_review"`
	CanEdit           bool                         `json:"can_edit"`
	CanWithdraw       bool                         `json:"can_withdraw"`
	CanAccept         bool                         `json:"can_accept"`
	CanReject         bool                         `json:"can_reject"`
}

type ProposalResponse struct {
	ProposalID        uuid.UUID                    `json:"proposal_id"`
	SenderUserID      uuid.UUID                    `json:"sender_user_id"`
	ReceiverUserID    uuid.UUID                    `json:"receiver_user_id"`
	SenderRole        string                       `json:"sender_role"`
	ReceiverRole      string                       `json:"receiver_role"`
	SenderProfileID   uuid.UUID                    `json:"sender_profile_id"`
	ReceiverProfileID uuid.UUID                    `json:"receiver_profile_id"`
	ProposalType      string                       `json:"proposal_type"`
	Title             string                       `json:"title"`
	Amount            int64                        `json:"amount"`
	TenorMonths       int                          `json:"tenor_months"`
	Scheme            string                       `json:"scheme"`
	Message           string                       `json:"message"`
	Status            string                       `json:"status"`
	SentAt            *time.Time                   `json:"sent_at"`
	AcceptedAt        *time.Time                   `json:"accepted_at"`
	RejectedAt        *time.Time                   `json:"rejected_at"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
	Attachments       []ProposalAttachmentResponse `json:"attachments"`
	Counterparty      ProposalCounterpartyResponse `json:"counterparty"`
}

type ProposalAttachmentResponse struct {
	AttachmentID uuid.UUID `json:"attachment_id"`
	FilePath     string    `json:"file_path"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	FileSize     int64     `json:"file_size"`
}

type ProposalCounterpartyResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	ProfileID uuid.UUID `json:"profile_id"`
	Name      string    `json:"name"`
	Subtitle  string    `json:"subtitle"`
}
