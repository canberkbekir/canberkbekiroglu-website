package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/canberkbekiroglu/website/internal/handlers"
	"github.com/canberkbekiroglu/website/internal/models"
	"github.com/canberkbekiroglu/website/internal/templates"
	"github.com/go-chi/chi/v5"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	// Resolve paths relative to the working directory.
	templateDir := filepath.Join("web", "templates")
	staticDir := filepath.Join("web", "static")
	dataFile := filepath.Join("data", "projects.json")
	contentDir := filepath.Join("data", "content")

	// Load portfolio data from JSON.
	portfolioData, err := models.LoadPortfolioData(dataFile)
	if err != nil {
		log.Fatalf("Failed to load portfolio data: %v", err)
	}
	log.Printf("Loaded %d sections from %s", len(portfolioData.Sections), dataFile)

	// Initialize template renderer.
	renderer, err := templates.New(templateDir)
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	// Create handler groups.
	pages := handlers.NewPages(renderer, portfolioData, contentDir)
	api := handlers.NewAPI(renderer)

	// Set up router.
	r := chi.NewRouter()

	// Middleware
	r.Use(handlers.Recoverer)
	r.Use(handlers.Logger)
	r.Use(handlers.SecurityHeaders)

	// Static files
	fileServer := http.FileServer(http.Dir(staticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Page routes
	r.Get("/", pages.HandleHome)
	r.Get("/resume", pages.HandleResume)
	r.Get("/contact", pages.HandleContact)
	r.Get("/portfolio/{slug}", pages.HandleProject)

	// API routes
	r.Post("/api/contact", api.HandleContactSubmit)
	r.Get("/api/health", api.HandleHealthCheck)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
