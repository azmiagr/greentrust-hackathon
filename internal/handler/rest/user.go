package rest

import (
	"greentrust-hackathon/model"
	"greentrust-hackathon/pkg/response"
	"net/http"

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
