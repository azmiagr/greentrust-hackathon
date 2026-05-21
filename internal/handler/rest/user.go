package rest

import (
	"greentrust-hackathon/model"
	"greentrust-hackathon/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (r *Rest) RegisterUser(c *gin.Context) {
	var param model.RegisterUserParam
	err := c.ShouldBindJSON(&param)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	result, err := r.service.UserService.RegisterUser(param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "success to register new user", result)
}

func (r *Rest) VerifyOTP(c *gin.Context) {
	var param model.VerifyOTPParam

	err := c.ShouldBindJSON(&param)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	sessionToken := c.GetHeader("X-Session-Token")

	result, err := r.service.UserService.VerifyOTP(sessionToken, param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to verify otp", result)
}

func (r *Rest) SubmitUserIdentity(c *gin.Context) {
	var param model.SubmitUserIdentityParam

	err := c.ShouldBindJSON(&param)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to bind input", err)
		return
	}

	sessionToken := c.GetHeader("X-Session-Token")

	result, err := r.service.UserService.SubmitUserIdentity(sessionToken, param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to submit identity", result)
}

func (r *Rest) SubmitBusinessProfile(c *gin.Context) {
	err := c.Request.ParseMultipartForm(32 << 20)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to parse multipart form", err)
		return
	}

	isServiceBusiness, err := strconv.ParseBool(c.PostForm("is_service_business"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "is_service_business must be boolean", err)
		return
	}

	form := c.Request.MultipartForm
	photos := form.File["photos"]

	param := model.SubmitBusinessProfileParam{
		BusinessName:        c.PostForm("business_name"),
		SectorID:            c.PostForm("sector_id"),
		BusinessDescription: c.PostForm("business_description"),
		IsServiceBusiness:   isServiceBusiness,
		BusinessAddressLine: c.PostForm("business_address_line"),
		BusinessProvince:    c.PostForm("business_province"),
		BusinessCity:        c.PostForm("business_city"),
		WhatsappNumber:      c.PostForm("whatsapp_number"),
		Photos:              photos,
	}

	sessionToken := c.GetHeader("X-Session-Token")

	result, err := r.service.UserService.SubmitBusinessProfile(sessionToken, param)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to submit business profile", result)
}
