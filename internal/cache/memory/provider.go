package memory

import "github.com/siddontang/github-repos-management/internal/cache"

// NewProvider creates a new memory cache provider
func NewProvider() cache.Provider {
	return func(config interface{}) (cache.Cache, error) {
		return NewCache(), nil
	}
}
