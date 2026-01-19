package main

import (
	"log"
	"net/http"

	httpDelivery "github.com/handokobeni/agent-openhands/internal/delivery/http"
	"github.com/handokobeni/agent-openhands/internal/delivery/http/middleware"
	"github.com/handokobeni/agent-openhands/internal/repository"
	"github.com/handokobeni/agent-openhands/internal/usecase"
)

func main() {
	// Repository Layer (Dependency Inversion Principle)
	productRepo := repository.NewProductRepositoryMemory()

	// Use Case Layer (Single Responsibility Principle)
	productUseCase := usecase.NewProductUseCase(productRepo)

	// Delivery Layer (Interface Segregation Principle)
	productHandler := httpDelivery.NewProductHandler(productUseCase)

	// Router setup
	mux := http.NewServeMux()
	mux.Handle("/api/products", productHandler)
	mux.Handle("/api/products/", productHandler)

	// Apply middleware (Open/Closed Principle)
	handler := middleware.Logging(middleware.CORS(mux))

	// Start server
	port := ":12000"
	log.Printf("Server starting on port %s", port)
	log.Printf("API Endpoints:")
	log.Printf("  GET    /api/products      - Get all products")
	log.Printf("  GET    /api/products/{id} - Get product by ID")
	log.Printf("  POST   /api/products      - Create new product")
	log.Printf("  PUT    /api/products/{id} - Update product")
	log.Printf("  DELETE /api/products/{id} - Delete product")

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
