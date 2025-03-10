package service

import "errors"

// Error definitions
var (
	ErrRepositoryExists      = errors.New("repository already exists")
	ErrRepositoryNotFound    = errors.New("repository not found")
	ErrInvalidRepositoryName = errors.New("invalid repository name format")
)
