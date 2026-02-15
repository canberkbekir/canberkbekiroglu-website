package templates

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// Renderer manages HTML template loading and rendering.
type Renderer struct {
	templates map[string]*template.Template
	baseDir   string
}

// templateFuncs returns the custom template function map.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"formatDate": func(t time.Time) string {
			return t.Format("January 2006")
		},
		"join": strings.Join,
		"currentYear": func() int {
			return time.Now().Year()
		},
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
	}
}

// New creates a new Renderer by parsing all templates from the given base directory.
// baseDir should point to the web/templates/ directory.
func New(baseDir string) (*Renderer, error) {
	r := &Renderer{
		templates: make(map[string]*template.Template),
		baseDir:   baseDir,
	}

	layoutFile := filepath.Join(baseDir, "layouts", "base.html")
	partialPattern := filepath.Join(baseDir, "partials", "*.html")

	partials, err := filepath.Glob(partialPattern)
	if err != nil {
		return nil, fmt.Errorf("finding partials: %w", err)
	}

	pagePattern := filepath.Join(baseDir, "pages", "*.html")
	pages, err := filepath.Glob(pagePattern)
	if err != nil {
		return nil, fmt.Errorf("finding pages: %w", err)
	}

	// For each page, parse layout + partials + that page together.
	for _, page := range pages {
		name := strings.TrimSuffix(filepath.Base(page), ".html")

		files := []string{layoutFile}
		files = append(files, partials...)
		files = append(files, page)

		t, err := template.New(filepath.Base(layoutFile)).Funcs(templateFuncs()).ParseFiles(files...)
		if err != nil {
			return nil, fmt.Errorf("parsing template %q: %w", name, err)
		}

		r.templates[name] = t
	}

	// Also parse partials standalone for HTMX partial responses.
	for _, partial := range partials {
		name := "partial_" + strings.TrimSuffix(filepath.Base(partial), ".html")
		t, err := template.New(filepath.Base(partial)).Funcs(templateFuncs()).ParseFiles(partial)
		if err != nil {
			return nil, fmt.Errorf("parsing partial %q: %w", name, err)
		}
		r.templates[name] = t
	}

	return r, nil
}

// Render writes the named template to w with the given data.
// For full pages, name is the page name (e.g., "home").
// For HTMX partials, prefix with "partial_" (e.g., "partial_contact_form").
func (r *Renderer) Render(w io.Writer, name string, data any) error {
	t, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %q not found", name)
	}
	return t.Execute(w, data)
}
