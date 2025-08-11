package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fabianoflorentino/mr-robot/adapters/inbound/http/controllers"
	"github.com/fabianoflorentino/mr-robot/internal/app"
)

var (
	APP_PORT string = os.Getenv("APP_PORT")
)

func InitHTTPServer(container app.Container) {
	mux := http.NewServeMux()

	// Register routes
	registerPaymentRoutes(mux, container)
	registerHealthCheckRoutes(mux)

	// Add middleware
	handler := loggingMiddleware(mux)

	server := &http.Server{
		Addr:    ":" + APP_PORT,
		Handler: handler,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on port %s", APP_PORT)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func registerPaymentRoutes(mux *http.ServeMux, container app.Container) {
	paymentController := controllers.NewPaymentController(container.GetPaymentQueue(), container.GetPaymentService())

	mux.HandleFunc("POST /payments", paymentController.PaymentProcess)
	mux.HandleFunc("GET /payments-summary", paymentController.PaymentsSummary)
	mux.HandleFunc("DELETE /payments-purge", paymentController.PurgePayments)
}

func registerHealthCheckRoutes(mux *http.ServeMux) {
	healthCheckController := controllers.NewHealthCheckController()

	mux.HandleFunc("GET /health", healthCheckController.HealthCheck)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)

		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
