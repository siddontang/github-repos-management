package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

// Pagination represents pagination information
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// RepositoryRequest represents a request to add a repository
type RepositoryRequest struct {
	FullName string `json:"full_name"`
}

// Bind validates the request
func (r *RepositoryRequest) Bind(req *http.Request) error {
	if r.FullName == "" {
		return errors.New("full_name is required")
	}
	return nil
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Err            error `json:"-"` // Application-specific error
	HTTPStatusCode int   `json:"-"` // HTTP status code

	Code    string      `json:"code"`              // Error code
	Message string      `json:"message"`           // Error message
	Details interface{} `json:"details,omitempty"` // Additional error details
}

// Render implements the render.Renderer interface
func (e *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrInvalidRequest returns a 400 Bad Request error
func ErrInvalidRequest(err error) render.Renderer {
	return &ErrorResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		Code:           "invalid_request",
		Message:        err.Error(),
	}
}

// ErrNotFound returns a 404 Not Found error
func ErrNotFound(err error) render.Renderer {
	return &ErrorResponse{
		Err:            err,
		HTTPStatusCode: http.StatusNotFound,
		Code:           "not_found",
		Message:        err.Error(),
	}
}

// ErrConflict returns a 409 Conflict error
func ErrConflict(err error) render.Renderer {
	return &ErrorResponse{
		Err:            err,
		HTTPStatusCode: http.StatusConflict,
		Code:           "conflict",
		Message:        err.Error(),
	}
}

// ErrInternalServer returns a 500 Internal Server Error
func ErrInternalServer(err error) render.Renderer {
	return &ErrorResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		Code:           "internal_server_error",
		Message:        "Internal server error",
		Details:        err.Error(),
	}
}
