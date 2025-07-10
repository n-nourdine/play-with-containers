package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/n-nourdine/play-with-containers/api-gateway/handlers"
	"github.com/n-nourdine/play-with-containers/api-gateway/middleware"
)

func main() {
	logger := log.New(os.Stdout, fmt.Sprintf("api-gateway running on port %s -> ",
		os.Getenv("API_GATEWAY_PORT")), log.LstdFlags)

	// Create handlers
	h, err := handlers.NewHandler(logger)
	if err != nil {
		logger.Fatalf("Failed to create handlers: %v", err)
	}
	defer h.Close()

	// Create HTTP multiplexer
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API Gateway is healthy"))
	})

	// Inventory API routes - proxy all /api/movies requests to inventory service
	mux.HandleFunc("GET /api/movies", h.ProxyToInventory)
	mux.HandleFunc("GET /api/movies/{id}", h.ProxyToInventory)
	mux.HandleFunc("POST /api/movies", h.ProxyToInventory)
	mux.HandleFunc("PUT /api/movies/{id}", h.ProxyToInventory)
	mux.HandleFunc("DELETE /api/movies/{id}", h.ProxyToInventory)
	mux.HandleFunc("DELETE /api/movies", h.ProxyToInventory)

	// Billing API route - send messages to RabbitMQ
	mux.HandleFunc("POST /api/billing", h.HandleBilling)

	// Serve OpenAPI documentation
	mux.HandleFunc("GET /api/docs", h.ServeOpenAPIDoc)
	mux.HandleFunc("GET /", h.ServeSwaggerUI)

	// Apply middleware
	handler := middleware.LoggingMiddleware(logger)(
		middleware.CORSMiddleware()(mux))

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", os.Getenv("API_GATEWAY_PORT")),
		Handler:      handler,
		ErrorLog:     logger,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Printf("Starting API Gateway on port %s", os.Getenv("API_GATEWAY_PORT"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	sig := <-c
	logger.Printf("Received signal: %v", sig)

	// Shutdown server gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("Server shutdown error: %v", err)
	}

	logger.Println("API Gateway stopped gracefully")
}
