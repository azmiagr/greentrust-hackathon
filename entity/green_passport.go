package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type GreenPassport struct {
	PassportID       uuid.UUID    `json:"passport_id" gorm:"type:varchar(36);primaryKey"`
	ProfileID        uuid.UUID    `json:"profile_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	GRSScore         float64      `json:"grs_score" gorm:"type:decimal(5,2);"`
	GRSBreakdown     GRSBreakdown `json:"grs_breakdown" gorm:"type:json;"`
	Status           string       `json:"status" gorm:"type:enum('draft','active','suspended');default:'draft'"`
	PublicSlug       string       `json:"public_slug" gorm:"type:varchar(100);uniqueIndex;"`
	QRCodeURL        string       `json:"qr_code_url" gorm:"type:text;"`
	BlockchainTxHash string       `json:"blockchain_tx_hash" gorm:"type:varchar(255);null"`
	ContractAddress  string       `json:"contract_address" gorm:"type:varchar(255);null"`
	NetworkName      string       `json:"network_name" gorm:"type:varchar(100);null"`
	ChainID          int64        `json:"chain_id" gorm:"type:bigint;default:0"`
	BlockNumber      uint64       `json:"block_number" gorm:"type:bigint unsigned;default:0"`
	IssuedAt         time.Time    `json:"issued_at" gorm:"type:timestamp;"`
	LastUpdatedAt    time.Time    `json:"last_updated_at" gorm:"type:timestamp;"`
}

type GRSCategoryBreakdown struct {
	Category   string  `json:"category"`
	Weight     float64 `json:"weight"`
	CI         float64 `json:"c_i"`
	DocCount   int     `json:"doc_count"`
	DocsNeeded int     `json:"docs_needed"`
}

type GRSBreakdown []GRSCategoryBreakdown

func (g GRSBreakdown) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GRSBreakdown) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan GRSBreakdown")
	}
	return json.Unmarshal(bytes, g)
}
