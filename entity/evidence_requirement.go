package entity

type EvidenceRequirement struct {
	RequirementID string `json:"requirement_id" gorm:"type:varchar(36);primaryKey"`
	CategoryID    string `json:"category_id" gorm:"type:varchar(10);not null;index"`
	Name          string `json:"name" gorm:"type:varchar(200);not null"`
	Description   string `json:"description" gorm:"type:text"`
	IsRequired    bool   `json:"is_required" gorm:"type:boolean;default:true"`
	SortOrder     int    `json:"sort_order" gorm:"not null"`

	EvidenceDocuments []EvidenceDocument `json:"evidence_documents" gorm:"foreignKey:RequirementID;constraint:onDelete:CASCADE"`
}
