package mariadb

import (
	"greentrust-hackathon/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entity.Role{},
	)

	if err != nil {
		return err
	}

	return nil
}
