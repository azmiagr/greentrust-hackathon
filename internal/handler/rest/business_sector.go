package rest

import (
	"greentrust-hackathon/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Rest) GetBusinessSectors(c *gin.Context) {
	result, err := r.service.UserService.GetBusinessSectors()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "success to get business sectors", result)
}
