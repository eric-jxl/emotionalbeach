package server

import (
	"emotionalBeach/config"
	"emotionalBeach/internal/dao"
	"emotionalBeach/internal/service"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Router *gin.Engine
	DB     *dao.Database
	Mail   *service.MailService
}

func NewServer(cfg *config.Config, db *dao.Database) (*Server, error) {
	r := NewRouter()
	mail := service.NewMailService(cfg.MailConfig)
	server := &Server{
		Router: r,
		DB:     db,
		Mail:   mail,
	}

	server.setupRoutes()
	return server, nil
}

func (s *Server) setupRoutes() {
	//Webhook Service
	s.Router.POST("/webhook", s.Mail.WebhookEmail)
}
