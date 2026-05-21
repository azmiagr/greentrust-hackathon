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

type ProposalActionParam struct {
	Reason string `json:"reason"`
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
