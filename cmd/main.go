package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpDelivery "github.com/handokobeni/agent-openhands/internal/delivery/http"
	"github.com/handokobeni/agent-openhands/internal/delivery/http/handler"
	"github.com/handokobeni/agent-openhands/internal/delivery/http/middleware"
	"github.com/handokobeni/agent-openhands/internal/infrastructure/config"
	"github.com/handokobeni/agent-openhands/internal/infrastructure/database"
	"github.com/handokobeni/agent-openhands/internal/infrastructure/jwt"
	"github.com/handokobeni/agent-openhands/internal/usecase/auth"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewSQLiteDB(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := database.NewUserRepository(db)
	refreshTokenRepo := database.NewRefreshTokenRepository(db)

	// Initialize infrastructure services
	tokenGenerator := jwt.NewTokenGenerator(jwt.Config{
		SecretKey:         cfg.JWT.SecretKey,
		AccessTokenExpiry: cfg.JWT.AccessTokenExpiry,
	})
	passwordHasher := jwt.NewPasswordHasher()

	// Initialize use cases
	authUseCase := auth.NewAuthUseCase(
		userRepo,
		refreshTokenRepo,
		passwordHasher,
		tokenGenerator,
		auth.Config{
			RefreshTokenExpiry: cfg.JWT.RefreshTokenExpiry,
		},
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(
		authUseCase,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(tokenGenerator)

	// Setup router
	router := httpDelivery.NewRouter(authHandler, authMiddleware)
	handler := router.SetupRoutes()

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		log.Printf("Access Token Expiry: %v", cfg.JWT.AccessTokenExpiry)
		log.Printf("Refresh Token Expiry: %v", cfg.JWT.RefreshTokenExpiry)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
