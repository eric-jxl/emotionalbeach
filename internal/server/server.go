// Package server provides the *http.Server factory.
package server

import (
	"emotionalBeach/config"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// NewHTTPServer constructs an *http.Server with timeouts from configuration.
func NewHTTPServer(cfg *config.Config, router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSec) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeoutSec) * time.Second,
	}
}

