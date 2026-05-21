package model

import (
	"time"

	"github.com/google/uuid"
)

type GetInvestorProfileParam struct {
	ProfileID uuid.UUID `json:"profile_id"`
	UserID    uuid.UUID `json:"user_id"`
}

type CreateInvestorPositionParam struct {
	Title           string   `json:"title" binding:"required"`
	InstitutionName string   `json:"institution_name" binding:"required"`
	EmploymentType  string   `json:"employment_type" binding:"required,oneof=full_time part_time self_employed freelance internship"`
	Location        string   `json:"location"`
	StartDate       string   `json:"start_date" binding:"required"`
	EndDate         string   `json:"end_date"`
	IsCurrent       bool     `json:"is_current"`
	Description     string   `json:"description" binding:"max=500"`
	Skills          []string `json:"skills"`
}

type UpdateInvestorPositionParam struct {
	Title           string   `json:"title" binding:"required"`
	InstitutionName string   `json:"institution_name" binding:"required"`
	EmploymentType  string   `json:"employment_type" binding:"required,oneof=full_time part_time self_employed freelance internship"`
	Location        string   `json:"location"`
	StartDate       string   `json:"start_date" binding:"required"`
	EndDate         string   `json:"end_date"`
	IsCurrent       bool     `json:"is_current"`
	Description     string   `json:"description" binding:"max=500"`
	Skills          []string `json:"skills"`
}

type SubmitInvestorInvestmentProfileParam struct {
	InvestorType string   `json:"investor_type" binding:"required"`
	TicketRange  string   `json:"ticket_range" binding:"required"`
	SectorIDs    []string `json:"sector_ids" binding:"required,min=1,max=3"`
}

type SubmitInvestorProfileParam struct {
	Position          CreateInvestorPositionParam          `json:"position" binding:"required"`
	InvestmentProfile SubmitInvestorInvestmentProfileParam `json:"investment_profile" binding:"required"`
}

type SubmitInvestorProfileResponse struct {
	ProfileID         uuid.UUID                         `json:"profile_id"`
	Position          InvestorPositionResponse          `json:"position"`
	InvestmentProfile InvestorInvestmentProfileResponse `json:"investment_profile"`
	Message           string                            `json:"message"`
}

type InvestorInvestmentProfileResponse struct {
	InvestorType string                   `json:"investor_type"`
	TicketRange  string                   `json:"ticket_range"`
	FocusSectors []BusinessSectorResponse `json:"focus_sectors"`
}

type InvestorProfileResponse struct {
	ProfileID         uuid.UUID                         `json:"profile_id"`
	FullName          string                            `json:"full_name"`
	Initials          string                            `json:"initials"`
	Email             string                            `json:"email"`
	PhoneNumber       string                            `json:"phone_number"`
	Company           string                            `json:"company"`
	Title             string                            `json:"title"`
	BaseLocation      string                            `json:"base_location"`
	PublicBio         string                            `json:"public_bio"`
	InvestmentProfile InvestorInvestmentProfileResponse `json:"investment_profile"`
	Positions         []InvestorPositionResponse        `json:"positions"`
}

type PublicInvestorDirectoryQuery struct {
	Search        string `form:"search"`
	InvestorTypes string `form:"investor_types"`
	SectorIDs     string `form:"sector_ids"`
	TicketRanges  string `form:"ticket_ranges"`
	Sort          string `form:"sort"`
	Page          int    `form:"page"`
	Limit         int    `form:"limit"`
}

type PublicInvestorDirectoryResponse struct {
	Meta    PublicInvestorDirectoryMeta    `json:"meta"`
	Filters PublicInvestorDirectoryFilters `json:"filters"`
	Items   []PublicInvestorDirectoryItem  `json:"items"`
}

type PublicInvestorDirectoryMeta struct {
	Page              int `json:"page"`
	Limit             int `json:"limit"`
	Total             int `json:"total"`
	Showing           int `json:"showing"`
	ActiveFilterCount int `json:"active_filter_count"`
}

type PublicInvestorDirectoryFilters struct {
	InvestorTypes []PublicInvestorTypeFilter        `json:"investor_types"`
	FocusSectors  []PublicInvestorSectorFilter      `json:"focus_sectors"`
	TicketRanges  []PublicInvestorTicketRangeFilter `json:"ticket_ranges"`
}

type PublicInvestorTypeFilter struct {
	InvestorType string `json:"investor_type"`
	Count        int    `json:"count"`
}

type PublicInvestorSectorFilter struct {
	SectorID   uuid.UUID `json:"sector_id"`
	SectorName string    `json:"sector_name"`
	Count      int       `json:"count"`
}

type PublicInvestorTicketRangeFilter struct {
	TicketRange string `json:"ticket_range"`
	Count       int    `json:"count"`
}

type PublicInvestorDirectoryItem struct {
	ProfileID       uuid.UUID                `json:"profile_id"`
	UserID          uuid.UUID                `json:"user_id"`
	FullName        string                   `json:"full_name"`
	Initials        string                   `json:"initials"`
	Title           string                   `json:"title"`
	InstitutionName string                   `json:"institution_name"`
	InvestorType    string                   `json:"investor_type"`
	TicketRange     string                   `json:"ticket_range"`
	FocusSectors    []BusinessSectorResponse `json:"focus_sectors"`
	PortfolioCount  int                      `json:"portfolio_count"`
	ApprovalRate    int                      `json:"approval_rate"`
}

type PublicInvestorDetailResponse struct {
	Profile           PublicInvestorDetailProfile         `json:"profile"`
	InvestmentProfile InvestorInvestmentProfileResponse   `json:"investment_profile"`
	Stats             PublicInvestorDetailStats           `json:"stats"`
	Positions         []PublicInvestorDetailPosition      `json:"positions"`
	Portfolio         []PublicInvestorDetailPortfolioItem `json:"portfolio"`
	RecentActivities  []PublicInvestorDetailActivity      `json:"recent_activities"`
	ProposalCTA       PublicInvestorDetailProposalCTA     `json:"proposal_cta"`
}

type PublicInvestorDetailProfile struct {
	ProfileID       uuid.UUID `json:"profile_id"`
	UserID          uuid.UUID `json:"user_id"`
	FullName        string    `json:"full_name"`
	Initials        string    `json:"initials"`
	Email           string    `json:"email"`
	Title           string    `json:"title"`
	InstitutionName string    `json:"institution_name"`
	BaseLocation    string    `json:"base_location"`
	PublicBio       string    `json:"public_bio"`
	IsVerified      bool      `json:"is_verified"`
}

type PublicInvestorDetailStats struct {
	PortfolioCount    int    `json:"portfolio_count"`
	AcceptedCount     int    `json:"accepted_count"`
	RejectedCount     int    `json:"rejected_count"`
	ApprovalRate      int    `json:"approval_rate"`
	ApprovalRateLabel string `json:"approval_rate_label"`
	PositionCount     int    `json:"position_count"`
}

type PublicInvestorDetailPosition struct {
	PositionID      uuid.UUID       `json:"position_id"`
	Initials        string          `json:"initials"`
	Title           string          `json:"title"`
	InstitutionName string          `json:"institution_name"`
	EmploymentType  string          `json:"employment_type"`
	Location        string          `json:"location"`
	StartDate       string          `json:"start_date"`
	EndDate         *string         `json:"end_date"`
	IsCurrent       bool            `json:"is_current"`
	DurationLabel   string          `json:"duration_label"`
	Description     string          `json:"description"`
	Skills          []SkillResponse `json:"skills"`
}

type PublicInvestorDetailPortfolioItem struct {
	ProfileID         uuid.UUID `json:"profile_id"`
	BusinessName      string    `json:"business_name"`
	SectorName        string    `json:"sector_name"`
	City              string    `json:"city"`
	FundedYear        int       `json:"funded_year"`
	DurationLabel     string    `json:"duration_label"`
	ProposalType      string    `json:"proposal_type"`
	ProposalTypeLabel string    `json:"proposal_type_label"`
}

type PublicInvestorDetailActivity struct {
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type PublicInvestorDetailProposalCTA struct {
	RequiresAuth      bool      `json:"requires_auth"`
	ReceiverRole      string    `json:"receiver_role"`
	ReceiverProfileID uuid.UUID `json:"receiver_profile_id"`
	Message           string    `json:"message"`
}

type PublicInvestorDirectoryFiltersParam struct {
	Search        string
	InvestorTypes []string
	SectorIDs     []string
	TicketRanges  []string
	Sort          string
	Page          int
	Limit         int
	Offset        int
}

type PublicInvestorDirectoryRow struct {
	ProfileID       string
	UserID          string
	InvestorType    string
	TicketRange     string
	FirstName       string
	LastName        string
	Email           string
	Title           string
	InstitutionName string
	PortfolioCount  int64
	AcceptedCount   int64
	RejectedCount   int64
}

type PublicInvestorDetailRow struct {
	ProfileID       string
	UserID          string
	InvestorType    string
	TicketRange     string
	FirstName       string
	LastName        string
	Email           string
	Title           string
	InstitutionName string
	BaseLocation    string
	PublicBio       string
	IsVerified      bool
	PortfolioCount  int64
	AcceptedCount   int64
	RejectedCount   int64
}

type PublicInvestorTypeFilterRow struct {
	InvestorType string
	Count        int64
}

type PublicInvestorSectorFilterRow struct {
	SectorID   string
	SectorName string
	Count      int64
}

type PublicInvestorTicketRangeFilterRow struct {
	TicketRange string
	Count       int64
}

type InvestorPositionResponse struct {
	PositionID      uuid.UUID       `json:"position_id"`
	Title           string          `json:"title"`
	InstitutionName string          `json:"institution_name"`
	EmploymentType  string          `json:"employment_type"`
	Location        string          `json:"location"`
	StartDate       string          `json:"start_date"`
	EndDate         *string         `json:"end_date"`
	IsCurrent       bool            `json:"is_current"`
	Description     string          `json:"description"`
	Skills          []SkillResponse `json:"skills"`
}

type SkillResponse struct {
	SkillID uuid.UUID `json:"skill_id"`
	Name    string    `json:"name"`
}
