package mariadb

import (
	"greentrust-hackathon/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entity.Role{},
		&entity.User{},
		&entity.OTP{},
		&entity.UserIdentity{},
		&entity.BusinessSector{},
		&entity.UMKMProfile{},
		&entity.InvestorProfile{},
		&entity.Skill{},
		&entity.InvestorPosition{},
		&entity.GreenPassport{},
		&entity.LocationPhoto{},
		&entity.UMKMProduct{},
		&entity.EvidenceCategory{},
		&entity.EvidenceRequirement{},
		&entity.EvidenceDocument{},
		&entity.EvidenceAIReview{},
	)

	if err != nil {
		return err
	}

	return nil
}
