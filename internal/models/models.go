package models

import (
	"encoding/json"
	"time"
)

// Repository represents a GitHub repository in the database
type Repository struct {
	Owner        string    `db:"owner"`
	Name         string    `db:"name"`
	FullName     string    `db:"full_name"`
	Description  string    `db:"description"`
	URL          string    `db:"url"`
	HTMLURL      string    `db:"html_url"`
	IsPrivate    bool      `db:"is_private"`
	LastSyncedAt time.Time `db:"last_synced_at"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// MarshalJSON customizes JSON marshaling for Repository
func (r *Repository) MarshalJSON() ([]byte, error) {
	type Alias Repository
	return json.Marshal(&struct {
		*Alias
		LastSyncedAt string `json:"last_synced_at"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
	}{
		Alias:        (*Alias)(r),
		LastSyncedAt: r.LastSyncedAt.Format(time.RFC3339),
		CreatedAt:    r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    r.UpdatedAt.Format(time.RFC3339),
	})
}

// PullRequest represents a GitHub pull request in the database
type PullRequest struct {
	RepositoryFullName string     `db:"repository_full_name"`
	Number             int        `db:"number"`
	Title              string     `db:"title"`
	Body               string     `db:"body"`
	State              string     `db:"state"`
	URL                string     `db:"url"`
	HTMLURL            string     `db:"html_url"`
	UserLogin          string     `db:"user_login"`
	UserAvatarURL      string     `db:"user_avatar_url"`
	UserURL            string     `db:"user_url"`
	UserHTMLURL        string     `db:"user_html_url"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
	ClosedAt           *time.Time `db:"closed_at"`
	MergedAt           *time.Time `db:"merged_at"`
}

// MarshalJSON customizes JSON marshaling for PullRequest
func (pr *PullRequest) MarshalJSON() ([]byte, error) {
	type Alias PullRequest

	// Format time fields
	createdAt := pr.CreatedAt.Format(time.RFC3339)
	updatedAt := pr.UpdatedAt.Format(time.RFC3339)

	// Handle nullable time fields
	var closedAt, mergedAt *string
	if pr.ClosedAt != nil {
		t := pr.ClosedAt.Format(time.RFC3339)
		closedAt = &t
	}
	if pr.MergedAt != nil {
		t := pr.MergedAt.Format(time.RFC3339)
		mergedAt = &t
	}

	return json.Marshal(&struct {
		*Alias
		CreatedAt string  `json:"created_at"`
		UpdatedAt string  `json:"updated_at"`
		ClosedAt  *string `json:"closed_at,omitempty"`
		MergedAt  *string `json:"merged_at,omitempty"`
	}{
		Alias:     (*Alias)(pr),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		ClosedAt:  closedAt,
		MergedAt:  mergedAt,
	})
}

// Issue represents a GitHub issue in the database
type Issue struct {
	RepositoryFullName string     `db:"repository_full_name"`
	Number             int        `db:"number"`
	Title              string     `db:"title"`
	Body               string     `db:"body"`
	State              string     `db:"state"`
	URL                string     `db:"url"`
	HTMLURL            string     `db:"html_url"`
	UserLogin          string     `db:"user_login"`
	UserAvatarURL      string     `db:"user_avatar_url"`
	UserURL            string     `db:"user_url"`
	UserHTMLURL        string     `db:"user_html_url"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
	ClosedAt           *time.Time `db:"closed_at"`
}

// MarshalJSON customizes JSON marshaling for Issue
func (issue *Issue) MarshalJSON() ([]byte, error) {
	type Alias Issue

	// Format time fields
	createdAt := issue.CreatedAt.Format(time.RFC3339)
	updatedAt := issue.UpdatedAt.Format(time.RFC3339)

	// Handle nullable time fields
	var closedAt *string
	if issue.ClosedAt != nil {
		t := issue.ClosedAt.Format(time.RFC3339)
		closedAt = &t
	}

	return json.Marshal(&struct {
		*Alias
		CreatedAt string  `json:"created_at"`
		UpdatedAt string  `json:"updated_at"`
		ClosedAt  *string `json:"closed_at,omitempty"`
	}{
		Alias:     (*Alias)(issue),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		ClosedAt:  closedAt,
	})
}

// Label represents a GitHub label in the database
type Label struct {
	Name        string `db:"name"`
	Color       string `db:"color"`
	Description string `db:"description"`
}

// PullRequestLabel represents a many-to-many relationship between pull requests and labels
type PullRequestLabel struct {
	RepositoryFullName string `db:"repository_full_name"`
	PullRequestNumber  int    `db:"pull_request_number"`
	LabelName          string `db:"label_name"`
}

// IssueLabel represents a many-to-many relationship between issues and labels
type IssueLabel struct {
	RepositoryFullName string `db:"repository_full_name"`
	IssueNumber        int    `db:"issue_number"`
	LabelName          string `db:"label_name"`
}

// PullRequestFilter represents filter options for pull requests
type PullRequestFilter struct {
	State     string
	Author    string
	Repo      string
	Label     string
	SortBy    string
	Direction string
	Since     time.Time
	GroupBy   string
	Page      int
	PerPage   int
}

// IssueFilter represents filter options for issues
type IssueFilter struct {
	State     string
	Author    string
	Repo      string
	Label     string
	SortBy    string
	Direction string
	Since     time.Time
	GroupBy   string
	Page      int
	PerPage   int
}

// Pagination represents pagination information
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
