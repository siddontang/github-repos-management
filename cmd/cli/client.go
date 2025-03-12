package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/siddontang/github-repos-management/internal/config"
	"github.com/siddontang/github-repos-management/internal/models"
	"github.com/siddontang/github-repos-management/internal/service"
)

// Client represents a service client wrapper
type Client struct {
	service *service.Service
	ctx     context.Context
}

// NewClient creates a new service client wrapper
func NewClient() (*Client, error) {
	// Load default configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: config.DBTypeFile, // Use file database by default
			Path: "data/github-repos.db",
		},
	}

	// Create service
	svc, err := service.NewService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return &Client{
		service: svc,
		ctx:     context.Background(),
	}, nil
}

// Pagination represents pagination information
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListRepositoriesResponse represents a response for listing repositories
type ListRepositoriesResponse struct {
	Data       []*models.Repository `json:"data"`
	Pagination *Pagination          `json:"pagination"`
}

// ListPullRequestsResponse represents a response for listing pull requests
type ListPullRequestsResponse struct {
	Data       []*models.PullRequest `json:"data"`
	Pagination *Pagination           `json:"pagination"`
}

// ListIssuesResponse represents a response for listing issues
type ListIssuesResponse struct {
	Data       []*models.Issue `json:"data"`
	Pagination *Pagination     `json:"pagination"`
}

// ListRepositories lists repositories that have been added
func (c *Client) ListRepositories(page, perPage int) (*ListRepositoriesResponse, error) {
	// Get repositories from service
	repos, total, err := c.service.ListRepositories(c.ctx, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	// Create pagination
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	return &ListRepositoriesResponse{
		Data: repos,
		Pagination: &Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// AddRepository adds a new repository to track
func (c *Client) AddRepository(fullName string) (*models.Repository, error) {
	// Add repository using service
	repo, err := c.service.AddRepository(c.ctx, fullName)
	if err != nil {
		return nil, fmt.Errorf("failed to add repository: %w", err)
	}

	return repo, nil
}

// GetRepository gets a repository by owner and name
func (c *Client) GetRepository(owner, name string) (*models.Repository, error) {
	// Get repository using service
	repo, err := c.service.GetRepository(c.ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	return repo, nil
}

// RemoveRepository removes a repository from tracking
func (c *Client) RemoveRepository(owner, name string) error {
	// Remove repository using service
	err := c.service.DeleteRepository(c.ctx, owner, name)
	if err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}

	return nil
}

// RefreshRepository forces a refresh of repository data
func (c *Client) RefreshRepository(owner, name string) error {
	// Refresh repository using service
	err := c.service.RefreshRepository(c.ctx, owner, name)
	if err != nil {
		return fmt.Errorf("failed to refresh repository: %w", err)
	}

	return nil
}

// ListPullRequests lists pull requests with filtering and pagination
func (c *Client) ListPullRequests(params map[string]string) (*ListPullRequestsResponse, error) {
	// Create filter
	filter := &models.PullRequestFilter{
		State:     params["state"],
		Author:    params["author"],
		Repo:      params["repo"],
		Label:     params["label"],
		SortBy:    params["sort"],
		Direction: params["direction"],
	}

	// Parse pagination
	page := 1
	perPage := 30

	if pageStr, ok := params["page"]; ok && pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr, ok := params["per_page"]; ok && perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
			perPage = pp
		}
	}

	filter.Page = page
	filter.PerPage = perPage

	// Parse since date
	if sinceStr, ok := params["since"]; ok && sinceStr != "" {
		if since, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			filter.Since = since
		}
	}

	// Get pull requests from service
	prs, pagination, err := c.service.ListPullRequests(c.ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	return &ListPullRequestsResponse{
		Data: prs,
		Pagination: &Pagination{
			Page:       pagination.Page,
			PerPage:    pagination.PerPage,
			Total:      pagination.Total,
			TotalPages: pagination.TotalPages,
		},
	}, nil
}

// ListIssues lists issues with filtering and pagination
func (c *Client) ListIssues(params map[string]string) (*ListIssuesResponse, error) {
	// Create filter
	filter := &models.IssueFilter{
		State:     params["state"],
		Author:    params["author"],
		Repo:      params["repo"],
		Label:     params["label"],
		SortBy:    params["sort"],
		Direction: params["direction"],
	}

	// Parse pagination
	page := 1
	perPage := 30

	if pageStr, ok := params["page"]; ok && pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr, ok := params["per_page"]; ok && perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
			perPage = pp
		}
	}

	filter.Page = page
	filter.PerPage = perPage

	// Parse since date
	if sinceStr, ok := params["since"]; ok && sinceStr != "" {
		if since, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			filter.Since = since
		}
	}

	// Get issues from service
	issues, pagination, err := c.service.ListIssues(c.ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	return &ListIssuesResponse{
		Data: issues,
		Pagination: &Pagination{
			Page:       pagination.Page,
			PerPage:    pagination.PerPage,
			Total:      pagination.Total,
			TotalPages: pagination.TotalPages,
		},
	}, nil
}

// RefreshAll forces a refresh of all repository data
func (c *Client) RefreshAll() error {
	// Get all repositories
	err := c.service.RefreshAll(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh all repositories: %w", err)
	}
	return nil
}

// GetStatus returns the current status of the client
func (c *Client) GetStatus() (map[string]interface{}, error) {
	// Get status from service
	status, err := c.service.GetStatus(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	return status, nil
}
