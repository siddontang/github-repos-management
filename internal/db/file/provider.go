package file

import (
	"github.com/siddontang/github-repos-management/internal/config"
	"github.com/siddontang/github-repos-management/internal/db"
)

// NewProvider creates a new file database provider
func NewProvider() db.Provider {
	return func(config *config.Config) (db.DB, error) {
		// Create a new file database with the path from config
		return NewDB(config.Database.Path)
	}
}
