package template

import (
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// InitRepository initializes the global template repository singleton
func InitRepository(dbClient *sqlx.DB) {
	_globalRepositoryMu.Lock()
	defer _globalRepositoryMu.Unlock()

	if _globalRepository != nil {
		zap.L().Info("Template repository already initialized")
		return
	}

	zap.L().Info("Initializing template repository")
	_globalRepository = NewPostgresRepository(dbClient)
}
