package model

import (
	"time"

	"github.com/google/uuid"
)

type GreenPassportCategoryScore struct {
	CategoryID string  `json:"category_id"`
	Score      float64 `json:"score"`
	Percent    float64 `json:"percent"`
	DocCount   int     `json:"doc_count"`
	DocsNeeded int     `json:"docs_needed"`
}

type IssueGreenPassportResponse struct {
	PassportID       uuid.UUID                    `json:"passport_id"`
	ProfileID        uuid.UUID                    `json:"profile_id"`
	GRSScore         float64                      `json:"grs_score"`
	Status           string                       `json:"status"`
	PublicSlug       string                       `json:"public_slug"`
	PassportURL      string                       `json:"passport_url"`
	QRCodeURL        string                       `json:"qr_code_url"`
	DocumentCount    int                          `json:"document_count"`
	CategoryScores   []GreenPassportCategoryScore `json:"category_scores"`
	Network          string                       `json:"network"`
	ChainID          int64                        `json:"chain_id"`
	ContractAddress  string                       `json:"contract_address"`
	BlockchainTxHash string                       `json:"blockchain_tx_hash"`
	BlockNumber      uint64                       `json:"block_number"`
	IssuedAt         time.Time                    `json:"issued_at"`
}
