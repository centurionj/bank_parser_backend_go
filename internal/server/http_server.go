package server

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/routers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HTTPServer struct {
	cfg    *config.Config
	router *gin.Engine
	db     *gorm.DB
}

func NewHTTPServer(cfg *config.Config, db *gorm.DB) *HTTPServer {
	r := gin.Default()

	server := &HTTPServer{
		cfg:    cfg,
		router: r,
		db:     db,
	}

	routers.SetupRoutes(r, db, cfg)

	return server
}

func (s *HTTPServer) Run() error {
	return s.router.Run(":" + s.cfg.HTTPPort)
}
