package model

import (
	"mime/multipart"

	"github.com/google/uuid"
)

type SubmitBusinessProfileParam struct {
	BusinessName        string                  `form:"business_name"`
	SectorID            string                  `form:"sector_id"`
	BusinessDescription string                  `form:"business_description"`
	IsServiceBusiness   bool                    `form:"is_service_business"`
	BusinessAddressLine string                  `form:"business_address_line"`
	BusinessProvince    string                  `form:"business_province"`
	BusinessCity        string                  `form:"business_city"`
	WhatsappNumber      string                  `form:"whatsapp_number"`
	Photos              []*multipart.FileHeader `form:"photos"`
}

type GetBusinessSectorParam struct {
	SectorID uuid.UUID `json:"sector_id"`
}

type BusinessSectorResponse struct {
	SectorID   uuid.UUID `json:"sector_id"`
	SectorName string    `json:"sector_name"`
}

type GetUMKMProfileParam struct {
	ProfileID uuid.UUID `json:"profile_id"`
	UserID    uuid.UUID `json:"user_id"`
}

type SubmitBusinessProfileResponse struct {
	ProfileID uuid.UUID `json:"profile_id"`
	PhotoURLs []string  `json:"photo_urls"`
	Message   string    `json:"message"`
}
