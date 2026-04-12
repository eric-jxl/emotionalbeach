// Package service contains the application's unified business-logic layer.
// A single Service struct holds every capability (user, relation, github, health, email)
// mirroring the pattern in arip-samp/internal/service/service.go.
// Handlers receive *Service directly — one dependency, zero fragmentation.
package service

import (
	"emotionalBeach/config"
	"emotionalBeach/internal/dao"
	"net/http"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Provider is the Wire provider set for the unified service layer.
// A single line replaces five separate service providers in wire.go.
var Provider = wire.NewSet(New)

// Service holds every business-logic capability of the application.
type Service struct {
	dao dao.Dao

	// ── GitHub OAuth ────────────────────────────────────────────────────────
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client

	// ── Health ──────────────────────────────────────────────────────────────
	db        *gorm.DB
	rdb       *redis.Client
	startTime time.Time

	// ── Email / SMTP ────────────────────────────────────────────────────────
	mailFrom string
	smtpUser string
	smtpPwd  string
}

// New constructs a Service from all infrastructure dependencies.
func New(cfg *config.Config, d dao.Dao, db *gorm.DB, rdb *redis.Client) *Service {
	return &Service{
		dao:          d,
		clientID:     cfg.Server.ClientID,
		clientSecret: cfg.Server.ClientSecret,
		redirectURI:  defaultRedirectURI,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		db:           db,
		rdb:          rdb,
		startTime:    time.Now(),
		mailFrom:     cfg.MailConfig.MailFrom,
		smtpUser:     cfg.MailConfig.SmtpUser,
		smtpPwd:      cfg.MailConfig.SmtpPassword,
	}
}

