package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/canberkbekiroglu/website/internal/admin"
	"github.com/canberkbekiroglu/website/internal/analytics"
	"github.com/canberkbekiroglu/website/internal/handlers"
	"github.com/canberkbekiroglu/website/internal/mailer"
	"github.com/canberkbekiroglu/website/internal/templates"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if present (ignored if missing).
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — using environment variables")
	}

	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	// Resolve paths relative to the working directory.
	templateDir := filepath.Join("web", "templates")
	staticDir := filepath.Join("web", "static")
	dataFile := filepath.Join("data", "projects.json")
	contentDir := filepath.Join("data", "content")

	// Initialize admin database.
	adminDB, err := admin.NewDB(filepath.Join("data", "admin.db"))
	if err != nil {
		log.Fatalf("Failed to initialize admin database: %v", err)
	}
	defer adminDB.Close()

	// Run JSON → SQLite migration if needed.
	if err := adminDB.MigrateFromJSON(dataFile); err != nil {
		log.Fatalf("Failed to migrate data: %v", err)
	}

	// Start session cleanup.
	adminDB.StartSessionCleanup()

	// Initialize template renderer.
	renderer, err := templates.New(templateDir)
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	// Configure mailer (optional — app works without it).
	var m *mailer.Mailer
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")
	contactEmail := os.Getenv("CONTACT_EMAIL")

	if smtpHost != "" && smtpUser != "" && smtpPass != "" {
		if smtpPort == "" {
			smtpPort = "587"
		}
		if smtpFrom == "" {
			smtpFrom = smtpUser
		}
		if contactEmail == "" {
			contactEmail = smtpUser
		}
		m = mailer.New(mailer.Config{
			Host:         smtpHost,
			Port:         smtpPort,
			Username:     smtpUser,
			Password:     smtpPass,
			FromAddress:  smtpFrom,
			ContactEmail: contactEmail,
		})
		log.Printf("SMTP mailer configured (host=%s, from=%s)", smtpHost, smtpFrom)
	} else {
		log.Println("SMTP not configured — contact form emails disabled")
	}

	// Trust proxy headers only if explicitly enabled (e.g. behind nginx/Cloudflare).
	if os.Getenv("TRUST_PROXY") == "true" {
		analytics.TrustProxy = true
		log.Println("Trusting X-Forwarded-For / X-Real-IP proxy headers")
	}

	// Initialize analytics database.
	analyticsDB, err := analytics.NewDB(filepath.Join("data", "analytics.db"))
	if err != nil {
		log.Fatalf("Failed to initialize analytics: %v", err)
	}
	defer analyticsDB.Close()
	analyticsDB.StartScheduler(m)

	// Admin password setup.
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	var adminPassHash string
	if adminPassword != "" {
		hash, err := admin.HashPassword(adminPassword)
		if err != nil {
			log.Fatalf("Failed to hash admin password: %v", err)
		}
		adminPassHash = hash
		log.Println("Admin panel enabled (password set via ADMIN_PASSWORD)")
	} else {
		log.Println("WARNING: ADMIN_PASSWORD not set — admin panel login disabled")
	}

	// Create handler groups.
	csrf := handlers.NewCSRFManager()
	pages := handlers.NewPages(renderer, adminDB, contentDir, csrf)
	api := handlers.NewAPI(renderer, m, csrf)
	adminHandler := handlers.NewAdmin(renderer, adminDB, analyticsDB, csrf, adminPassHash)

	// Set up router.
	r := chi.NewRouter()

	// Middleware
	r.Use(handlers.Recoverer)
	r.Use(handlers.Logger)
	r.Use(handlers.SecurityHeaders)
	r.Use(analyticsDB.Tracker)

	// Static files (directory listing disabled).
	fileServer := http.FileServer(http.Dir(staticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", noDirectoryListing(fileServer)))

	// Page routes
	r.Get("/", pages.HandleHome)
	r.Get("/resume", pages.HandleResume)
	r.Get("/contact", pages.HandleContact)
	r.Get("/portfolio/{slug}", pages.HandleProject)

	// API routes (rate-limited: 5 requests per 15 minutes per IP)
	contactLimiter := handlers.NewRateLimiter(5, 15*time.Minute)
	r.With(handlers.RateLimit(contactLimiter)).Post("/api/contact", api.HandleContactSubmit)
	r.Get("/api/health", api.HandleHealthCheck)

	// Admin routes (unauthenticated)
	r.Get("/admin/login", adminHandler.HandleLoginPage)
	r.Post("/admin/login", adminHandler.HandleLogin)

	// Admin routes (authenticated)
	r.Group(func(r chi.Router) {
		r.Use(adminHandler.RequireAdmin)

		r.Get("/admin", adminHandler.HandleDashboard)
		r.Get("/admin/logout", adminHandler.HandleLogout)

		// Projects
		r.Get("/admin/projects", adminHandler.HandleProjectsList)
		r.Get("/admin/projects/new", adminHandler.HandleProjectForm)
		r.Post("/admin/projects/new", adminHandler.HandleProjectCreate)
		r.Get("/admin/projects/{id}/edit", adminHandler.HandleProjectForm)
		r.Post("/admin/projects/{id}/edit", adminHandler.HandleProjectUpdate)
		r.Delete("/admin/projects/{id}", adminHandler.HandleProjectDelete)

		// Sections
		r.Get("/admin/sections", adminHandler.HandleSectionsList)
		r.Post("/admin/sections", adminHandler.HandleSectionCreate)
		r.Post("/admin/sections/{id}", adminHandler.HandleSectionUpdate)
		r.Delete("/admin/sections/{id}", adminHandler.HandleSectionDelete)

		// Filter Tags
		r.Get("/admin/filters", adminHandler.HandleFilterTagsList)
		r.Post("/admin/filters", adminHandler.HandleFilterTagCreate)
		r.Post("/admin/filters/{id}", adminHandler.HandleFilterTagUpdate)
		r.Delete("/admin/filters/{id}", adminHandler.HandleFilterTagDelete)

		// Analytics API
		r.Get("/api/admin/analytics/visitors", adminHandler.HandleAnalyticsVisitors)
		r.Get("/api/admin/analytics/performance", adminHandler.HandleAnalyticsPerformance)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// noDirectoryListing wraps a file server to return 404 for directory paths.
func noDirectoryListing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") || r.URL.Path == "" {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
