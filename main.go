package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/noel-vega/deployment-agent/auth"
	"github.com/noel-vega/deployment-agent/docker"
	"github.com/noel-vega/deployment-agent/handlers"
	"github.com/noel-vega/deployment-agent/middleware"
	"github.com/noel-vega/deployment-agent/registry"
)

func main() {
	// Initialize users from environment
	if err := auth.InitializeUsers(); err != nil {
		log.Fatalf("Failed to initialize users: %v", err)
	}

	// Initialize authentication service
	if err := auth.Initialize(); err != nil {
		log.Fatalf("Failed to initialize auth service: %v", err)
	}
	log.Printf("Authentication service initialized")
	log.Printf("Access token duration: %v", auth.AccessTokenDuration)
	log.Printf("Refresh token duration: %v", auth.RefreshTokenDuration)

	// Initialize docker service
	dockerService, err := docker.NewService()
	if err != nil {
		log.Fatalf("Failed to initialize docker service: %v", err)
	}
	defer dockerService.Close()

	// Initialize registry client
	registryClient, err := registry.NewClient()
	if err != nil {
		log.Printf("Warning: Failed to initialize registry client: %v", err)
		log.Printf("Registry endpoints will not be available")
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	containersHandler := handlers.NewContainersHandler(dockerService)

	// Initialize registry handler if registry client is available
	var registryHandler *handlers.RegistryHandler
	if registryClient != nil {
		registryHandler = handlers.NewRegistryHandler(registryClient)
		defer registryClient.Close()
	}
	imagesHandler := handlers.NewImagesHandler(dockerService)

	// Setup router
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Public routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("deployment-agent"))
	})

	// Auth routes (public)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		r.Post("/refresh", authHandler.Refresh)
	})

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Protected)

		r.Get("/auth/me", authHandler.Me)

		// Registry endpoints (if registry client is configured)
		if registryHandler != nil {
			r.Get("/registry/repositories", registryHandler.ListRepositories)
			r.Get("/registry/repositories/{name}/tags", registryHandler.ListTags)
			r.Get("/registry/catalog", registryHandler.ListRepositoriesWithTags)
		}
		r.Get("/containers", containersHandler.List)
		r.Post("/containers/{id}/stop", containersHandler.Stop)
		r.Post("/containers/{id}/start", containersHandler.Start)
		r.Get("/images", imagesHandler.List)
	})

	// Start server
	log.Println("Starting server on :5000")
	if err := http.ListenAndServe(":5000", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
