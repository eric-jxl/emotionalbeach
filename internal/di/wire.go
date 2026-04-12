//go:build wireinject
// +build wireinject

// The build constraint above tells the Go compiler to ignore this file;
// Wire reads it instead to generate wire_gen.go.

package di

import (
	"emotionalBeach/config"
	"emotionalBeach/internal/dao"
	"emotionalBeach/internal/infra"
	"emotionalBeach/internal/server"
	"emotionalBeach/internal/service"

	"github.com/google/wire"
)

// InitializeApp is the Wire injector.
// Wire reads this function signature, resolves the provider graph, and writes
// the concrete implementation into wire_gen.go.
func InitializeApp(cfg *config.Config) (*App, func(), error) {
	wire.Build(infra.Provider, dao.Provider, service.Provider, server.New, NewApp)
	return nil, nil, nil
}
