package model

type InvestorWatchedUMKMSummary struct {
	WatchedUMKMCount          int64
	WatchedUMKMGrowthThisWeek int64
	AveragePortfolioGRS       float64
	ReadyOrUnggulCount        int64
	NewUnggulUMKMThisWeek     int64
}

type InvestorDashboardUMKMRow struct {
	ProfileID            string
	BusinessName         string
	SectorName           string
	City                 string
	GRSScore             float64
	OnChainDocumentCount int64
}
