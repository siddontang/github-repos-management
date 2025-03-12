package db

import (
	"context"

	"github.com/siddontang/github-repos-management/internal/config"
	"github.com/siddontang/github-repos-management/internal/models"
)

// DB defines the interface for storing GitHub data
type DB interface {
	// Repository operations
	AddRepository(ctx context.Context, repo *models.Repository) error
	GetRepository(ctx context.Context, owner, name string) (*models.Repository, error)
	ListRepositories(ctx context.Context, page, perPage int) ([]*models.Repository, int, error)
	UpdateRepository(ctx context.Context, repo *models.Repository) error
	DeleteRepository(ctx context.Context, owner, name string) error

	// Pull request operations
	AddPullRequest(ctx context.Context, pr *models.PullRequest) error
	GetPullRequest(ctx context.Context, repoFullName string, number int) (*models.PullRequest, error)
	ListPullRequests(ctx context.Context, repoFullName string, page, perPage int) ([]*models.PullRequest, int, error)
	UpdatePullRequest(ctx context.Context, pr *models.PullRequest) error
	DeletePullRequest(ctx context.Context, repoFullName string, number int) error

	// Issue operations
	AddIssue(ctx context.Context, issue *models.Issue) error
	GetIssue(ctx context.Context, repoFullName string, number int) (*models.Issue, error)
	ListIssues(ctx context.Context, repoFullName string, page, perPage int) ([]*models.Issue, int, error)
	UpdateIssue(ctx context.Context, issue *models.Issue) error
	DeleteIssue(ctx context.Context, repoFullName string, number int) error

	// Label operations
	AddLabel(ctx context.Context, label *models.Label) error
	GetLabel(ctx context.Context, name string) (*models.Label, error)
	ListLabels(ctx context.Context, page, perPage int) ([]*models.Label, int, error)
	UpdateLabel(ctx context.Context, label *models.Label) error
	DeleteLabel(ctx context.Context, name string) error

	// Pull request label operations
	AddPullRequestLabel(ctx context.Context, repoFullName string, prNumber int, labelName string) error
	ListPullRequestLabels(ctx context.Context, repoFullName string, prNumber int) ([]*models.Label, error)
	RemovePullRequestLabel(ctx context.Context, repoFullName string, prNumber int, labelName string) error

	// Issue label operations
	AddIssueLabel(ctx context.Context, repoFullName string, issueNumber int, labelName string) error
	ListIssueLabels(ctx context.Context, repoFullName string, issueNumber int) ([]*models.Label, error)
	RemoveIssueLabel(ctx context.Context, repoFullName string, issueNumber int, labelName string) error

	// Maintenance operations
	Close() error
	Ping(ctx context.Context) error

	// Sync operations
	Sync() error
}

// Provider is a function that creates a new db instance
type Provider func(config *config.Config) (DB, error)
