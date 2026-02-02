package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/nf-software/nf-toolbox/internal/api"
	"github.com/nf-software/nf-toolbox/internal/config"
	"github.com/nf-software/nf-toolbox/internal/db"
)

func main() {
	// Load .env in development
	godotenv.Load()

	// Load config
	cfg := config.Load()

	// Connect to database
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Create server
	server := api.NewServer(database, cfg)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://nf-software.com", "http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Post("/auth/register", server.HandleRegister)
		r.Post("/auth/login", server.HandleLogin)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(server.AuthMiddleware)

			// License
			r.Post("/activate", server.HandleActivate)
			r.Post("/renew", server.HandleRenew)
			r.Get("/license", server.HandleGetLicense)

			// Devices
			r.Get("/devices", server.HandleListDevices)
			r.Delete("/devices/{id}", server.HandleDeactivateDevice)

			// User
			r.Get("/me", server.HandleGetMe)
			r.Get("/tools", server.HandleListTools)
		})
	})

	// Webhooks (Stripe)
	r.Post("/webhooks/stripe", server.HandleStripeWebhook)

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
