package repository

import (
	"greentrust-hackathon/entity"

	"gorm.io/gorm"
)

type ISkillRepository interface {
	GetSkillByName(tx *gorm.DB, name string) (*entity.Skill, error)
	CreateSkill(tx *gorm.DB, skill *entity.Skill) error
	SearchSkills(tx *gorm.DB, query string, limit int) ([]*entity.Skill, error)
}

type SkillRepository struct {
	db *gorm.DB
}

func NewSkillRepository(db *gorm.DB) ISkillRepository {
	return &SkillRepository{db: db}
}

func (r *SkillRepository) GetSkillByName(tx *gorm.DB, name string) (*entity.Skill, error) {
	var skill entity.Skill
	err := tx.Debug().Where("LOWER(name) = LOWER(?)", name).First(&skill).Error
	if err != nil {
		return nil, err
	}

	return &skill, nil
}

func (r *SkillRepository) CreateSkill(tx *gorm.DB, skill *entity.Skill) error {
	return tx.Debug().Create(skill).Error
}

func (r *SkillRepository) SearchSkills(tx *gorm.DB, query string, limit int) ([]*entity.Skill, error) {
	var skills []*entity.Skill
	db := tx.Debug().Order("name ASC").Limit(limit)
	if query != "" {
		db = db.Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%")
	}

	err := db.Find(&skills).Error
	if err != nil {
		return nil, err
	}

	return skills, nil
}
