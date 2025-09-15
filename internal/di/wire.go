//go:build wireinject
// +build wireinject

package di

// wire.go
import (
	"emotionalBeach/config"
	"emotionalBeach/internal/dao"
	"emotionalBeach/internal/server"

	"github.com/google/wire"
)

//go:generate wire
func InitializeApp() (*server.Server, error) {
	wire.Build(server.NewServer, dao.NewDatabase, config.LoadConfig)
	return &server.Server{}, nil
}
