package model

import "github.com/google/uuid"

type GetInvestorProfileParam struct {
	ProfileID uuid.UUID `json:"profile_id"`
	UserID    uuid.UUID `json:"user_id"`
}

type CreateInvestorPositionParam struct {
	Title           string   `json:"title" binding:"required"`
	InstitutionName string   `json:"institution_name" binding:"required"`
	EmploymentType  string   `json:"employment_type" binding:"required,oneof=full_time part_time self_employed freelance internship"`
	Location        string   `json:"location"`
	StartDate       string   `json:"start_date" binding:"required"`
	EndDate         string   `json:"end_date"`
	IsCurrent       bool     `json:"is_current"`
	Description     string   `json:"description" binding:"max=500"`
	Skills          []string `json:"skills"`
}

type UpdateInvestorPositionParam struct {
	Title           string   `json:"title" binding:"required"`
	InstitutionName string   `json:"institution_name" binding:"required"`
	EmploymentType  string   `json:"employment_type" binding:"required,oneof=full_time part_time self_employed freelance internship"`
	Location        string   `json:"location"`
	StartDate       string   `json:"start_date" binding:"required"`
	EndDate         string   `json:"end_date"`
	IsCurrent       bool     `json:"is_current"`
	Description     string   `json:"description" binding:"max=500"`
	Skills          []string `json:"skills"`
}

type InvestorPositionResponse struct {
	PositionID      uuid.UUID       `json:"position_id"`
	Title           string          `json:"title"`
	InstitutionName string          `json:"institution_name"`
	EmploymentType  string          `json:"employment_type"`
	Location        string          `json:"location"`
	StartDate       string          `json:"start_date"`
	EndDate         *string         `json:"end_date"`
	IsCurrent       bool            `json:"is_current"`
	Description     string          `json:"description"`
	Skills          []SkillResponse `json:"skills"`
}

type SkillResponse struct {
	SkillID uuid.UUID `json:"skill_id"`
	Name    string    `json:"name"`
}
