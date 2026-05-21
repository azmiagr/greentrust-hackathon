package mariadb

import (
	"greentrust-hackathon/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entity.Role{},
		&entity.User{},
		&entity.UserIdentity{},
		&entity.BusinessSector{},
		&entity.UMKMProfile{},
		&entity.GreenPassport{},
		&entity.LocationPhoto{},
		&entity.EvidenceDocument{},
		&entity.UMKMProduct{},
	)

	if err != nil {
		return err
	}

	return nil
}
