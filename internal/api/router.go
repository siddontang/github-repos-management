package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/siddontang/github-repos-management/internal/service"
)

// NewRouter creates a new HTTP router
func NewRouter(svc *service.Service) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60))
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Create API handler
	h := NewHandler(svc)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Repository routes
		r.Route("/repositories", func(r chi.Router) {
			r.Get("/", h.ListRepositories)
			r.Post("/", h.AddRepository)
			r.Route("/{owner}/{repo}", func(r chi.Router) {
				r.Get("/", h.GetRepository)
				r.Delete("/", h.RemoveRepository)
				r.Post("/refresh", h.RefreshRepository)
			})
		})

		// Pull request routes
		r.Get("/pulls", h.ListPullRequests)

		// Issue routes
		r.Get("/issues", h.ListIssues)

		// Service routes
		r.Post("/refresh", h.RefreshAll)
		r.Get("/status", h.GetStatus)
	})

	// Serve OpenAPI documentation
	r.Get("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "api/openapi.yaml")
	})

	return r
}
