package model

import (
	"time"

	"github.com/google/uuid"
)

type InvestorPortfolioQuery struct {
	SectorID string  `form:"sector_id"`
	MinGRS   float64 `form:"min_grs"`
}

type InvestorPortfolioResponse struct {
	Items []InvestorPortfolioItem `json:"items"`
}

type InvestorPortfolioItem struct {
	ProfileID              uuid.UUID  `json:"profile_id"`
	BusinessName           string     `json:"business_name"`
	SectorName             string     `json:"sector_name"`
	City                   string     `json:"city"`
	Status                 string     `json:"status"`
	TotalValue             int64      `json:"total_value"`
	FormattedTotalValue    string     `json:"formatted_total_value"`
	CurrentGRS             float64    `json:"current_grs"`
	GRSTrend               string     `json:"grs_trend"`
	FundedSince            *time.Time `json:"funded_since"`
	FundedSinceLabel       string     `json:"funded_since_label"`
	AcceptedProposalCount  int        `json:"accepted_proposal_count"`
	LatestProposalID       uuid.UUID  `json:"latest_proposal_id"`
	LatestProposalTitle    string     `json:"latest_proposal_title"`
	LatestProposalAccepted *time.Time `json:"latest_proposal_accepted_at"`
}

type InvestorPortfolioRow struct {
	ProfileID              string
	BusinessName           string
	SectorName             string
	City                   string
	TotalValue             int64
	CurrentGRS             float64
	FundedSince            *time.Time
	AcceptedProposalCount  int64
	LatestProposalID       string
	LatestProposalTitle    string
	LatestProposalAccepted *time.Time
}
