package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client represents an API client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

// Repository represents a GitHub repository
type Repository struct {
	Owner        string `json:"owner"`
	Name         string `json:"name"`
	FullName     string `json:"full_name"`
	Description  string `json:"description"`
	URL          string `json:"url"`
	HTMLURL      string `json:"html_url"`
	IsPrivate    bool   `json:"is_private"`
	LastSyncedAt string `json:"last_synced_at"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	UserLogin string `json:"user_login"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Issue represents a GitHub issue
type Issue struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	UserLogin string `json:"user_login"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Pagination represents pagination information
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ListRepositoriesResponse represents a response for listing repositories
type ListRepositoriesResponse struct {
	Data       []*Repository `json:"data"`
	Pagination *Pagination   `json:"pagination"`
}

// ListPullRequestsResponse represents a response for listing pull requests
type ListPullRequestsResponse struct {
	Data       []*PullRequest `json:"data"`
	Pagination *Pagination    `json:"pagination"`
}

// ListIssuesResponse represents a response for listing issues
type ListIssuesResponse struct {
	Data       []*Issue    `json:"data"`
	Pagination *Pagination `json:"pagination"`
}

// ListRepositories lists all repositories
func (c *Client) ListRepositories(page, perPage int) (*ListRepositoriesResponse, error) {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/repositories", c.baseURL))
	if err != nil {
		return nil, err
	}

	// Add query parameters
	q := u.Query()
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("per_page", fmt.Sprintf("%d", perPage))
	u.RawQuery = q.Encode()

	// Make request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	// Parse response
	var response ListRepositoriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// AddRepository adds a new repository
func (c *Client) AddRepository(fullName string) (*Repository, error) {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/repositories", c.baseURL))
	if err != nil {
		return nil, err
	}

	// Build request body
	body := map[string]string{
		"full_name": fullName,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Make request
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	// Parse response
	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, err
	}

	return &repository, nil
}

// GetRepository gets a repository by owner and name
func (c *Client) GetRepository(owner, name string) (*Repository, error) {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/repositories/%s/%s", c.baseURL, owner, name))
	if err != nil {
		return nil, err
	}

	// Make request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	// Parse response
	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, err
	}

	return &repository, nil
}

// RemoveRepository removes a repository
func (c *Client) RemoveRepository(owner, name string) error {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/repositories/%s/%s", c.baseURL, owner, name))
	if err != nil {
		return err
	}

	// Make request
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return err
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// RefreshRepository forces a refresh of repository data
func (c *Client) RefreshRepository(owner, name string) error {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/repositories/%s/%s/refresh", c.baseURL, owner, name))
	if err != nil {
		return err
	}

	// Make request
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return err
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusAccepted {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// ListPullRequests lists pull requests with filtering and pagination
func (c *Client) ListPullRequests(params map[string]string) (*ListPullRequestsResponse, error) {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/pulls", c.baseURL))
	if err != nil {
		return nil, err
	}

	// Add query parameters
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	// Make request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	// Parse response
	var response ListPullRequestsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ListIssues lists issues with filtering and pagination
func (c *Client) ListIssues(params map[string]string) (*ListIssuesResponse, error) {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/issues", c.baseURL))
	if err != nil {
		return nil, err
	}

	// Add query parameters
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	// Make request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	// Parse response
	var response ListIssuesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// RefreshAll forces a refresh of all repository data
func (c *Client) RefreshAll() error {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/refresh", c.baseURL))
	if err != nil {
		return err
	}

	// Make request
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return err
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusAccepted {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// GetStatus returns the current status of the service
func (c *Client) GetStatus() (map[string]interface{}, error) {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/status", c.baseURL))
	if err != nil {
		return nil, err
	}

	// Make request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	// Parse response
	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return status, nil
}

// handleErrorResponse handles error responses
func (c *Client) handleErrorResponse(resp *http.Response) error {
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response: %w", err)
	}

	// Try to parse as JSON
	var errorResponse ErrorResponse
	if err := json.Unmarshal(body, &errorResponse); err != nil {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return fmt.Errorf("%s: %s", errorResponse.Code, errorResponse.Message)
}
