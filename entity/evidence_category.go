package entity

type EvidenceCategory struct {
	CategoryID string  `json:"category_id" gorm:"type:varchar(10);primaryKey"` // BB, PP, PL, EE, SK, LK
	Name       string  `json:"name" gorm:"type:varchar(150);not null"`
	Weight     float64 `json:"weight" gorm:"type:decimal(5,2);not null"`
	SortOrder  int     `json:"sort_order" gorm:"not null"`

	Requirements []EvidenceRequirement `json:"requirements" gorm:"foreignKey:CategoryID;constraint:onDelete:CASCADE"`
}
