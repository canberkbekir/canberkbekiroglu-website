package handlers

import (
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/canberkbekiroglu/website/internal/admin"
	"github.com/canberkbekiroglu/website/internal/models"
	"github.com/canberkbekiroglu/website/internal/templates"
	"github.com/go-chi/chi/v5"
)

// Pages holds page route handlers.
type Pages struct {
	renderer   *templates.Renderer
	db         *admin.DB
	contentDir string
	csrf       *CSRFManager
}

// NewPages creates a new Pages handler group.
func NewPages(r *templates.Renderer, db *admin.DB, contentDir string, csrf *CSRFManager) *Pages {
	return &Pages{renderer: r, db: db, contentDir: contentDir, csrf: csrf}
}

// HandleHome renders the portfolio home page.
func (p *Pages) HandleHome(w http.ResponseWriter, r *http.Request) {
	pd, err := p.db.LoadPortfolioData()
	if err != nil {
		http.Error(w, "Error loading data", http.StatusInternalServerError)
		return
	}
	filterTags, err := p.db.LoadFilterTagsForTemplate()
	if err != nil {
		http.Error(w, "Error loading filter tags", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":      "Portfolio",
		"YearGroups": models.GroupByYear(pd.AllProjects()),
		"FilterTags": filterTags,
	}
	if err := p.renderer.Render(w, "home", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// HandleResume renders the resume page.
func (p *Pages) HandleResume(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Resume",
	}
	if err := p.renderer.Render(w, "resume", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// HandleContact renders the contact form page.
func (p *Pages) HandleContact(w http.ResponseWriter, r *http.Request) {
	token, err := p.csrf.SetCookie(w)
	if err != nil {
		http.Error(w, "Error generating form token", http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"Title":     "Contact",
		"CSRFToken": token,
	}
	if err := p.renderer.Render(w, "contact", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// HandleProject renders an individual project page.
func (p *Pages) HandleProject(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	pd, err := p.db.LoadPortfolioData()
	if err != nil {
		http.Error(w, "Error loading data", http.StatusInternalServerError)
		return
	}

	project := pd.GetProjectBySlug(slug)
	if project == nil {
		http.NotFound(w, r)
		return
	}

	// If project has an external URL, redirect there (after validation).
	if project.ExternalURL != "" {
		if isSafeRedirectURL(project.ExternalURL) {
			http.Redirect(w, r, project.ExternalURL, http.StatusFound)
		} else {
			http.Error(w, "Invalid redirect URL", http.StatusBadRequest)
		}
		return
	}

	// Try to load custom HTML content from data/content/{slug}.html
	var content template.HTML
	contentPath := filepath.Join(p.contentDir, slug+".html")
	if raw, err := os.ReadFile(contentPath); err == nil {
		content = template.HTML(raw)
	}

	data := map[string]any{
		"Title":   project.Title,
		"Project": project,
		"Content": content,
	}
	if err := p.renderer.Render(w, "project", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// isSafeRedirectURL validates that a URL is an absolute HTTP(S) URL
// to prevent open redirect attacks via protocol-relative or javascript: URLs.
func isSafeRedirectURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "https" || u.Scheme == "http"
}
