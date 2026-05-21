package entity

import "github.com/google/uuid"

type Role struct {
	RoleID   uuid.UUID `json:"role_id" gorm:"type:varchar(36);primaryKey"`
	RoleName string    `json:"role_name" gorm:"type:varchar(100);not null"`

	Users []User `json:"user" gorm:"foreignKey:RoleID"`
}
