package middleware

import (
	"greentrust-hackathon/internal/service"
	"greentrust-hackathon/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type Interface interface {
	Cors() gin.HandlerFunc
}

type middleware struct {
	service *service.Service
	jwtAuth jwt.Interface
}

func Init(service *service.Service, jwtAuth jwt.Interface) Interface {
	return &middleware{
		service: service,
		jwtAuth: jwtAuth,
	}
}
