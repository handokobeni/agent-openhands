package http

import (
	"net/http"

	"github.com/handokobeni/agent-openhands/internal/delivery/http/handler"
	"github.com/handokobeni/agent-openhands/internal/delivery/http/middleware"
)

// Router handles HTTP routing
type Router struct {
	mux            *http.ServeMux
	authHandler    *handler.AuthHandler
	authMiddleware *middleware.AuthMiddleware
}

// NewRouter creates a new HTTP router
func NewRouter(authHandler *handler.AuthHandler, authMiddleware *middleware.AuthMiddleware) *Router {
	return &Router{
		mux:            http.NewServeMux(),
		authHandler:    authHandler,
		authMiddleware: authMiddleware,
	}
}

// SetupRoutes configures all routes
func (r *Router) SetupRoutes() http.Handler {
	// Public routes (no authentication required)
	r.mux.HandleFunc("/api/auth/register", r.authHandler.Register)
	r.mux.HandleFunc("/api/auth/login", r.authHandler.Login)
	r.mux.HandleFunc("/api/auth/refresh", r.authHandler.RefreshToken)
	r.mux.HandleFunc("/api/auth/logout", r.authHandler.Logout)

	// Protected routes (authentication required)
	r.mux.Handle("/api/auth/logout-all", r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.LogoutAll)))
	r.mux.Handle("/api/auth/sessions", r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.GetSessions)))
	r.mux.Handle("/api/auth/me", r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.Me)))

	// Health check
	r.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Apply global middleware
	var handler http.Handler = r.mux
	handler = middleware.Logging(handler)
	handler = middleware.CORS(handler)

	return handler
}
