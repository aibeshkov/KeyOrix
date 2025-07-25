package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/core"
	"github.com/secretlyhq/secretly/server/http/handlers"
	customMiddleware "github.com/secretlyhq/secretly/server/middleware"
)

// NewRouter creates and configures the HTTP router
func NewRouter(cfg *config.Config, coreService *core.SecretlyCore) (http.Handler, error) {
	r := chi.NewRouter()

	// Apply middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.Logger())
	r.Use(customMiddleware.Recovery())
	r.Use(middleware.Timeout(60))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure based on your needs
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize handlers
	secretHandler, err := handlers.NewSecretHandler(coreService)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret handler: %w", err)
	}
	
	shareHandler, err := handlers.NewShareHandler(coreService)
	if err != nil {
		return nil, fmt.Errorf("failed to create share handler: %w", err)
	}

	// Health check endpoint
	r.Get("/health", handlers.HealthCheck)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Authentication middleware for API routes
		r.Use(customMiddleware.Authentication())

		// Secrets endpoints
		r.Route("/secrets", func(r chi.Router) {
			// Require secrets.read permission for GET operations
			r.With(customMiddleware.RequirePermission("secrets.read")).Get("/", secretHandler.ListSecrets)
			r.With(customMiddleware.RequirePermission("secrets.read")).Get("/{id}", secretHandler.GetSecret)
			r.With(customMiddleware.RequirePermission("secrets.read")).Get("/{id}/versions", secretHandler.GetSecretVersions)
			r.With(customMiddleware.RequirePermission("secrets.read")).Get("/{id}/shares", shareHandler.ListSecretShares)

			// Require secrets.write permission for write operations
			r.With(customMiddleware.RequirePermission("secrets.write")).Post("/", secretHandler.CreateSecret)
			r.With(customMiddleware.RequirePermission("secrets.write")).Put("/{id}", secretHandler.UpdateSecret)
			r.With(customMiddleware.RequirePermission("secrets.write")).Post("/{id}/share", shareHandler.ShareSecret)

			// Require secrets.delete permission for delete operations
			r.With(customMiddleware.RequirePermission("secrets.delete")).Delete("/{id}", secretHandler.DeleteSecret)
		})
		
		// Shares endpoints
		r.Route("/shares", func(r chi.Router) {
			// Require secrets.read permission for GET operations
			r.With(customMiddleware.RequirePermission("secrets.read")).Get("/", shareHandler.ListShares)
			
			// Require secrets.write permission for write operations
			r.With(customMiddleware.RequirePermission("secrets.write")).Put("/{id}", shareHandler.UpdateSharePermission)
			
			// Require secrets.delete permission for delete operations
			r.With(customMiddleware.RequirePermission("secrets.write")).Delete("/{id}", shareHandler.RevokeShare)
		})
		
		// Shared secrets endpoint
		r.With(customMiddleware.RequirePermission("secrets.read")).Get("/shared-secrets", shareHandler.ListSharedSecrets)

		// Users endpoints (RBAC)
		r.Route("/users", func(r chi.Router) {
			r.Use(customMiddleware.RequirePermission("users.read"))
			r.Get("/", handlers.ListUsers)
			r.Post("/", handlers.CreateUser)
			r.Get("/{id}", handlers.GetUser)
			r.Put("/{id}", handlers.UpdateUser)
			r.Delete("/{id}", handlers.DeleteUser)
		})

		// Roles endpoints (RBAC)
		r.Route("/roles", func(r chi.Router) {
			r.Use(customMiddleware.RequirePermission("roles.read"))
			r.Get("/", handlers.ListRoles)
			r.Post("/", handlers.CreateRole)
			r.Get("/{id}", handlers.GetRole)
			r.Put("/{id}", handlers.UpdateRole)
			r.Delete("/{id}", handlers.DeleteRole)
		})

		// User roles endpoints
		r.Route("/user-roles", func(r chi.Router) {
			r.Use(customMiddleware.RequirePermission("roles.assign"))
			r.Post("/", handlers.AssignRole)
			r.Delete("/", handlers.RemoveRole)
			r.Get("/user/{userId}", handlers.GetUserRoles)
		})

		// Audit logs endpoints
		r.Route("/audit", func(r chi.Router) {
			r.Use(customMiddleware.RequirePermission("audit.read"))
			r.Get("/logs", handlers.GetAuditLogs)
			r.Get("/rbac-logs", handlers.GetRBACAuditLogs)
		})

		// System endpoints
		r.Route("/system", func(r chi.Router) {
			r.Use(customMiddleware.RequirePermission("system.read"))
			r.Get("/info", handlers.GetSystemInfo)
			r.Get("/metrics", handlers.GetMetrics)
		})
	})

	// Swagger UI (optional, based on config)
	if cfg.Server.HTTP.SwaggerEnabled {
		r.Mount("/swagger/", handlers.SwaggerHandler())
	}

	// OpenAPI spec endpoint
	r.Get("/openapi.yaml", handlers.OpenAPISpec)

	return r, nil
}
