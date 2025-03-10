package github

import "time"

// Repository represents a GitHub repository
type Repository struct {
	Owner       User      `json:"owner"`
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	HTMLURL     string    `json:"html_url"`
	Private     bool      `json:"private"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"`
	URL       string     `json:"url"`
	HTMLURL   string     `json:"html_url"`
	User      User       `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	MergedAt  *time.Time `json:"merged_at"`
	Labels    []Label    `json:"labels"`
}

// Issue represents a GitHub issue
type Issue struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"`
	URL       string     `json:"url"`
	HTMLURL   string     `json:"html_url"`
	User      User       `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	Labels    []Label    `json:"labels"`
}

// User represents a GitHub user
type User struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	URL       string `json:"url"`
	HTMLURL   string `json:"html_url"`
}

// Label represents a GitHub label
type Label struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// RateLimit represents GitHub API rate limit information
type RateLimit struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     int64     `json:"reset"`
	ResetTime time.Time `json:"-"`
}

// PullRequestOptions represents options for listing pull requests
type PullRequestOptions struct {
	State     string
	Sort      string
	Direction string
	PerPage   int
	Page      int
}

// IssueOptions represents options for listing issues
type IssueOptions struct {
	State     string
	Sort      string
	Direction string
	PerPage   int
	Page      int
}
