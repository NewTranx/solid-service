package server

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"fmt"
)

type ServiceEndpoint struct {
	Host     string
	Port     int
	WorkPath string
	srv      *http.Server
}

func (s *ServiceEndpoint) Start() {
	router := gin.Default()

	v1 := router.Group("v1")

	{
		v1.POST("/convert/upload", s.handleUpload)
	}

	router.Run(fmt.Sprintf("%s:%d", s.Host, s.Port))
}

func (s *ServiceEndpoint) handleUpload(c *gin.Context) {

}
