package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	done := make(chan bool, 1)      // Channel to tell server ready to shutdown
	quit := make(chan os.Signal, 1) // Channel to trap the OS signals
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	transitAPIClient := NewTransitAPIClient(&http.Client{
		Timeout: time.Duration(5 * time.Second),
	},
		getEnvVariable("HERE_API_TRANSIT_URL"),
		getEnvVariable("HERE_API_KEY"))
	weatherAPIClient := NewWeatherAPIClient(&http.Client{
		Timeout: time.Duration(5 * time.Second),
	},
		getEnvVariable("HERE_API_WEATHER_URL"),
		getEnvVariable("HERE_API_KEY"))
	handlers := Handler{transitAPIClient: transitAPIClient, weatherAPIClient: weatherAPIClient}
	// Routes require an authentication
	r.Group(func(r chi.Router) {
		r.Use(APIKeyAuth()) // Apply API key authentication
		// All responses in Miles and Fahrenheit
		r.Get("/api/v1/stations", handlers.getStations)
	})
	// Swagger.yaml
	r.Get("/swagger/swagger.yaml", handlers.getSwagger)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.yaml"), //The url pointing to API definition"
	))
	server := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      r,
		ErrorLog:     log.New(os.Stdout, "", log.LstdFlags),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	go gracefullShutdown(ctx, cancel, server, quit, done)
	log.Println("Server is ready to handle requests at", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", server.Addr, err)
	}
	<-done
	log.Println("Server stopped")
}

func gracefullShutdown(ctx context.Context, cancel context.CancelFunc, server *http.Server, quit <-chan os.Signal, done chan<- bool) {
	<-quit // We got Terminate signal
	log.Println("Server is shutting down...")

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
	close(done)
	cancel()
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{StatusCode: statusCode, Message: message})
}

func getEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
