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

type GreenPassportStatusResponse struct {
	Issued         bool                             `json:"issued"`
	Message        string                           `json:"message"`
	Profile        *GreenPassportStatusProfile      `json:"profile,omitempty"`
	GreenPassport  *GreenPassportStatusDetail       `json:"green_passport,omitempty"`
	Share          *GreenPassportStatusShare        `json:"share,omitempty"`
	OnChainProof   *GreenPassportStatusOnChainProof `json:"on_chain_proof,omitempty"`
	NextAction     *GreenPassportStatusNextAction   `json:"next_action,omitempty"`
	CategoryScores []GreenPassportCategoryScore     `json:"category_scores,omitempty"`
}

type GreenPassportStatusProfile struct {
	ProfileID    uuid.UUID `json:"profile_id"`
	BusinessName string    `json:"business_name"`
	OwnerName    string    `json:"owner_name,omitempty"`
	SectorName   string    `json:"sector_name"`
	Province     string    `json:"province"`
	City         string    `json:"city"`
}

type GreenPassportStatusDetail struct {
	PassportID    uuid.UUID `json:"passport_id"`
	PublicSlug    string    `json:"public_slug"`
	PassportURL   string    `json:"passport_url"`
	QRCodeURL     string    `json:"qr_code_url"`
	GRSScore      float64   `json:"grs_score"`
	Tier          string    `json:"tier"`
	TierLabel     string    `json:"tier_label"`
	Status        string    `json:"status"`
	IssuedAt      time.Time `json:"issued_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

type GreenPassportStatusShare struct {
	URL string `json:"url"`
}

type GreenPassportStatusOnChainProof struct {
	Network          string `json:"network"`
	ChainID          int64  `json:"chain_id"`
	ContractAddress  string `json:"contract_address"`
	BlockchainTxHash string `json:"blockchain_tx_hash"`
	ShortTxHash      string `json:"short_tx_hash"`
	BlockNumber      uint64 `json:"block_number"`
	ExplorerURL      string `json:"explorer_url,omitempty"`
	Confirmation     string `json:"confirmation"`
}

type GreenPassportStatusNextAction struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	TargetScore float64 `json:"target_score,omitempty"`
}

type PublicUMKMDirectoryQuery struct {
	Search    string `form:"search"`
	SectorIDs string `form:"sector_ids"`
	Tiers     string `form:"tiers"`
	Provinces string `form:"provinces"`
	Sort      string `form:"sort"`
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
}

type PublicUMKMDirectoryResponse struct {
	Meta    PublicUMKMDirectoryMeta     `json:"meta"`
	Filters *PublicUMKMDirectoryFilters `json:"filters,omitempty"`
	Items   []PublicUMKMDirectoryItem   `json:"items"`
}

type PublicUMKMDirectoryMeta struct {
	Page              int `json:"page"`
	Limit             int `json:"limit"`
	Total             int `json:"total"`
	Showing           int `json:"showing"`
	ActiveFilterCount int `json:"active_filter_count"`
}

type PublicUMKMDirectoryFilters struct {
	Sectors   []PublicUMKMSectorFilter   `json:"sectors"`
	Provinces []PublicUMKMProvinceFilter `json:"provinces"`
	Tiers     []PublicUMKMTierFilter     `json:"tiers"`
}

type PublicUMKMSectorFilter struct {
	SectorID   uuid.UUID `json:"sector_id"`
	SectorName string    `json:"sector_name"`
	Count      int       `json:"count"`
}

type PublicUMKMProvinceFilter struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type PublicUMKMTierFilter struct {
	Tier     string  `json:"tier"`
	Label    string  `json:"label"`
	MinScore float64 `json:"min_score"`
	MaxScore float64 `json:"max_score"`
	Count    int     `json:"count"`
}

type PublicUMKMDirectoryItem struct {
	ProfileID            uuid.UUID                `json:"profile_id"`
	BusinessName         string                   `json:"business_name"`
	SectorName           string                   `json:"sector_name"`
	Province             string                   `json:"province"`
	City                 string                   `json:"city"`
	Description          string                   `json:"description"`
	PhotoURL             string                   `json:"photo_url"`
	GRSScore             float64                  `json:"grs_score"`
	Tier                 string                   `json:"tier"`
	TierLabel            string                   `json:"tier_label"`
	GreenPassport        PublicGreenPassportBrief `json:"green_passport"`
	OnChainDocumentCount int                      `json:"on_chain_document_count"`
}

type PublicUMKMDetailResponse struct {
	Profile       PublicUMKMDetailProfile            `json:"profile"`
	GreenPassport PublicGreenPassportDetail          `json:"green_passport"`
	Summary       PublicUMKMDetailSummary            `json:"summary"`
	Categories    []PublicUMKMDetailEvidenceCategory `json:"categories"`
}

type PublicUMKMDetailProfile struct {
	ProfileID      uuid.UUID `json:"profile_id"`
	BusinessName   string    `json:"business_name"`
	SectorName     string    `json:"sector_name"`
	Province       string    `json:"province"`
	City           string    `json:"city"`
	Description    string    `json:"description"`
	PhotoURL       string    `json:"photo_url"`
	WhatsappNumber string    `json:"whatsapp_number"`
}

type PublicGreenPassportDetail struct {
	PassportID       uuid.UUID `json:"passport_id"`
	PublicSlug       string    `json:"public_slug"`
	PassportURL      string    `json:"passport_url"`
	QRCodeURL        string    `json:"qr_code_url"`
	GRSScore         float64   `json:"grs_score"`
	Tier             string    `json:"tier"`
	TierLabel        string    `json:"tier_label"`
	Status           string    `json:"status"`
	IssuedAt         time.Time `json:"issued_at"`
	LastUpdatedAt    time.Time `json:"last_updated_at"`
	BlockchainTxHash string    `json:"blockchain_tx_hash"`
	ContractAddress  string    `json:"contract_address"`
	Network          string    `json:"network"`
	ChainID          int64     `json:"chain_id"`
	BlockNumber      uint64    `json:"block_number"`
}

type PublicUMKMDetailSummary struct {
	VerifiedDocumentCount int `json:"verified_document_count"`
	PrivateDocumentCount  int `json:"private_document_count"`
	CategoryCount         int `json:"category_count"`
}

type PublicUMKMDetailEvidenceCategory struct {
	CategoryID      string                             `json:"category_id"`
	Code            string                             `json:"code"`
	Name            string                             `json:"name"`
	Weight          float64                            `json:"weight"`
	RequiredCount   int                                `json:"required_count"`
	FulfilledCount  int                                `json:"fulfilled_count"`
	ProgressPercent float64                            `json:"progress_percent"`
	Score           float64                            `json:"score"`
	Status          string                             `json:"status"`
	Documents       []PublicUMKMDetailEvidenceDocument `json:"documents"`
}

type PublicUMKMDetailEvidenceDocument struct {
	EvidenceID       uuid.UUID `json:"evidence_id"`
	RequirementID    *string   `json:"requirement_id,omitempty"`
	FileName         string    `json:"file_name"`
	FileHash         string    `json:"file_hash"`
	MimeType         string    `json:"mime_type"`
	FileSize         int64     `json:"file_size"`
	Status           string    `json:"status"`
	BlockchainTxHash string    `json:"blockchain_tx_hash"`
	CreatedAt        time.Time `json:"created_at"`
}

type PublicGreenPassportBrief struct {
	PassportID       uuid.UUID `json:"passport_id"`
	PublicSlug       string    `json:"public_slug"`
	PassportURL      string    `json:"passport_url"`
	Status           string    `json:"status"`
	IssuedAt         time.Time `json:"issued_at"`
	BlockchainTxHash string    `json:"blockchain_tx_hash"`
}

type PublicUMKMDirectoryFiltersParam struct {
	Search    string
	SectorIDs []string
	Tiers     []string
	Provinces []string
	Sort      string
	Page      int
	Limit     int
	Offset    int
}

type PublicUMKMDirectoryRow struct {
	ProfileID            string
	BusinessName         string
	SectorName           string
	Province             string
	City                 string
	Description          string
	PhotoURL             string
	PassportID           string
	GRSScore             float64
	PassportStatus       string
	PublicSlug           string
	BlockchainTxHash     string
	IssuedAt             time.Time
	OnChainDocumentCount int64
}

type PublicUMKMSectorFilterRow struct {
	SectorID   string
	SectorName string
	Count      int64
}

type PublicUMKMProvinceFilterRow struct {
	Name  string
	Count int64
}

type PublicUMKMTierFilterRow struct {
	Tier  string
	Count int64
}
