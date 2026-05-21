package model

import (
	"time"

	"github.com/google/uuid"
)

type InvestorDashboardResponse struct {
	Greeting          InvestorDashboardGreeting         `json:"greeting"`
	Summary           InvestorDashboardSummary          `json:"summary"`
	RecommendedUMKMs  []InvestorDashboardUMKM           `json:"recommended_umkm"`
	RecentActivities  []InvestorDashboardActivity       `json:"recent_activities"`
	InvestmentProfile InvestorInvestmentProfileResponse `json:"investment_profile"`
}

type InvestorDashboardGreeting struct {
	Name      string `json:"name"`
	Subtitle  string `json:"subtitle"`
	Highlight string `json:"highlight"`
}

type InvestorDashboardSummary struct {
	WatchedUMKMCount           int     `json:"watched_umkm_count"`
	WatchedUMKMGrowthThisWeek  int     `json:"watched_umkm_growth_this_week"`
	ActiveProposalsCount       int     `json:"active_proposals_count"`
	AcceptedProposalsCount     int     `json:"accepted_proposals_count"`
	WaitingConfirmationCount   int     `json:"waiting_confirmation_count"`
	AveragePortfolioGRS        float64 `json:"average_portfolio_grs"`
	PortfolioMajorityTier      string  `json:"portfolio_majority_tier"`
	NewUnggulUMKMThisWeekCount int     `json:"new_unggul_umkm_this_week_count"`
}

type InvestorDashboardUMKM struct {
	ProfileID            uuid.UUID `json:"profile_id"`
	BusinessName         string    `json:"business_name"`
	SectorName           string    `json:"sector_name"`
	City                 string    `json:"city"`
	GRSScore             float64   `json:"grs_score"`
	Tier                 string    `json:"tier"`
	OnChainDocumentCount int       `json:"on_chain_document_count"`
}

type InvestorDashboardActivity struct {
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type InvestorProposalDashboardStats struct {
	ActiveProposalsCount     int64
	AcceptedProposalsCount   int64
	WaitingConfirmationCount int64
}
