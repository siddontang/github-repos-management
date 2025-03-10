package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/siddontang/github-repos-management/internal/models"
	"github.com/siddontang/github-repos-management/internal/service"
)

// Handler handles API requests
type Handler struct {
	service *service.Service
}

// NewHandler creates a new API handler
func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

// Repository handlers

// ListRepositories lists all repositories
func (h *Handler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page, perPage := getPaginationParams(r)

	// Get repositories
	repos, total, err := h.service.ListRepositories(r.Context(), page, perPage)
	if err != nil {
		render.Render(w, r, ErrInternalServer(err))
		return
	}

	// Build response
	pagination := &Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: (total + perPage - 1) / perPage,
	}

	render.JSON(w, r, map[string]interface{}{
		"data":       repos,
		"pagination": pagination,
	})
}

// AddRepository adds a new repository
func (h *Handler) AddRepository(w http.ResponseWriter, r *http.Request) {
	// Parse request
	data := &RepositoryRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Add repository
	repo, err := h.service.AddRepository(r.Context(), data.FullName)
	if err != nil {
		if errors.Is(err, service.ErrRepositoryExists) {
			render.Render(w, r, ErrConflict(err))
			return
		}
		render.Render(w, r, ErrInternalServer(err))
		return
	}

	// Return created repository
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, repo)
}

// GetRepository gets a repository by owner and name
func (h *Handler) GetRepository(w http.ResponseWriter, r *http.Request) {
	// Get owner and repo from URL
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	// Get repository
	repository, err := h.service.GetRepository(r.Context(), owner, repo)
	if err != nil {
		render.Render(w, r, ErrNotFound(err))
		return
	}

	render.JSON(w, r, repository)
}

// RemoveRepository removes a repository
func (h *Handler) RemoveRepository(w http.ResponseWriter, r *http.Request) {
	// Get owner and repo from URL
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	// Delete repository
	err := h.service.DeleteRepository(r.Context(), owner, repo)
	if err != nil {
		render.Render(w, r, ErrNotFound(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RefreshRepository forces a refresh of repository data
func (h *Handler) RefreshRepository(w http.ResponseWriter, r *http.Request) {
	// Get owner and repo from URL
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	// Refresh repository
	err := h.service.RefreshRepository(r.Context(), owner, repo)
	if err != nil {
		render.Render(w, r, ErrNotFound(err))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// Pull request handlers

// ListPullRequests lists pull requests with filtering and pagination
func (h *Handler) ListPullRequests(w http.ResponseWriter, r *http.Request) {
	// Parse filter parameters
	filter := parsePullRequestFilter(r)

	// Get pull requests
	prs, pagination, err := h.service.ListPullRequests(r.Context(), filter)
	if err != nil {
		render.Render(w, r, ErrInternalServer(err))
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"data":       prs,
		"pagination": pagination,
	})
}

// Issue handlers

// ListIssues lists issues with filtering and pagination
func (h *Handler) ListIssues(w http.ResponseWriter, r *http.Request) {
	// Parse filter parameters
	filter := parseIssueFilter(r)

	// Get issues
	issues, pagination, err := h.service.ListIssues(r.Context(), filter)
	if err != nil {
		render.Render(w, r, ErrInternalServer(err))
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"data":       issues,
		"pagination": pagination,
	})
}

// Service handlers

// RefreshAll forces a refresh of all repository data
func (h *Handler) RefreshAll(w http.ResponseWriter, r *http.Request) {
	err := h.service.RefreshAll(r.Context())
	if err != nil {
		render.Render(w, r, ErrInternalServer(err))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetStatus returns the current status of the service
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.service.GetStatus(r.Context())
	if err != nil {
		render.Render(w, r, ErrInternalServer(err))
		return
	}

	render.JSON(w, r, status)
}

// Helper functions

// getPaginationParams extracts pagination parameters from the request
func getPaginationParams(r *http.Request) (int, int) {
	page := 1
	perPage := 30

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	return page, perPage
}

// parsePullRequestFilter extracts pull request filter parameters from the request
func parsePullRequestFilter(r *http.Request) *models.PullRequestFilter {
	page, perPage := getPaginationParams(r)

	filter := &models.PullRequestFilter{
		State:     r.URL.Query().Get("state"),
		Author:    r.URL.Query().Get("author"),
		Repo:      r.URL.Query().Get("repo"),
		Label:     r.URL.Query().Get("label"),
		SortBy:    r.URL.Query().Get("sort"),
		Direction: r.URL.Query().Get("direction"),
		GroupBy:   r.URL.Query().Get("group_by"),
		Page:      page,
		PerPage:   perPage,
	}

	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		if since, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			filter.Since = since
		}
	}

	return filter
}

// parseIssueFilter extracts issue filter parameters from the request
func parseIssueFilter(r *http.Request) *models.IssueFilter {
	page, perPage := getPaginationParams(r)

	filter := &models.IssueFilter{
		State:     r.URL.Query().Get("state"),
		Author:    r.URL.Query().Get("author"),
		Repo:      r.URL.Query().Get("repo"),
		Label:     r.URL.Query().Get("label"),
		SortBy:    r.URL.Query().Get("sort"),
		Direction: r.URL.Query().Get("direction"),
		GroupBy:   r.URL.Query().Get("group_by"),
		Page:      page,
		PerPage:   perPage,
	}

	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		if since, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			filter.Since = since
		}
	}

	return filter
}
