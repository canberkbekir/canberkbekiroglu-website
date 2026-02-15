package handlers

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/canberkbekiroglu/website/internal/models"
	"github.com/canberkbekiroglu/website/internal/templates"
	"github.com/go-chi/chi/v5"
)

// Pages holds page route handlers.
type Pages struct {
	renderer   *templates.Renderer
	data       *models.PortfolioData
	contentDir string
}

// NewPages creates a new Pages handler group.
func NewPages(r *templates.Renderer, data *models.PortfolioData, contentDir string) *Pages {
	return &Pages{renderer: r, data: data, contentDir: contentDir}
}

// HandleHome renders the portfolio home page.
func (p *Pages) HandleHome(w http.ResponseWriter, r *http.Request) {
	// Build year groups for each section
	type SectionData struct {
		Name       string
		Slug       string
		YearGroups []models.YearGroup
	}
	sections := make([]SectionData, len(p.data.Sections))
	for i, s := range p.data.Sections {
		sections[i] = SectionData{
			Name:       s.Name,
			Slug:       s.Slug,
			YearGroups: models.GroupByYear(s.Projects),
		}
	}

	data := map[string]any{
		"Title":    "Portfolio",
		"Sections": sections,
	}
	if err := p.renderer.Render(w, "home", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// HandleResume renders the resume page.
func (p *Pages) HandleResume(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":      "Resume",
		"Experience": models.SampleResumeEntries(),
		"Skills":     models.SampleSkills(),
	}
	if err := p.renderer.Render(w, "resume", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// HandleContact renders the contact form page.
func (p *Pages) HandleContact(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Contact",
	}
	if err := p.renderer.Render(w, "contact", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// HandleProject renders an individual project page.
func (p *Pages) HandleProject(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	project := p.data.GetProjectBySlug(slug)
	if project == nil {
		http.NotFound(w, r)
		return
	}

	// If project has an external URL, redirect there
	if project.ExternalURL != "" {
		http.Redirect(w, r, project.ExternalURL, http.StatusFound)
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
