package models

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Project represents a portfolio project.
type Project struct {
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Year        int      `json:"year"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	HoverImage  string   `json:"hoverImage"`
	Tags        []string `json:"tags"`
	Category    string   `json:"category"`
	Company     string   `json:"company"`
	Role        string   `json:"role"`
	ExternalURL string   `json:"externalUrl,omitempty"`
	Highlighted bool     `json:"highlighted,omitempty"`
	FilterTags  []string `json:"filterTags,omitempty"`
}

// LinkURL returns the URL the card should link to.
func (p Project) LinkURL() string {
	if p.ExternalURL != "" {
		return p.ExternalURL
	}
	return "/portfolio/" + p.Slug
}

// IsExternal returns true if this project links to an external site.
func (p Project) IsExternal() bool {
	return p.ExternalURL != ""
}

// Section represents a group of projects under a heading.
type Section struct {
	Name     string    `json:"name"`
	Slug     string    `json:"slug"`
	Projects []Project `json:"projects"`
}

// YearGroup holds projects grouped by year for timeline rendering.
type YearGroup struct {
	Year     int
	Projects []Project
}

// ResumeEntry represents a work experience entry.
type ResumeEntry struct {
	Company     string
	Role        string
	Period      string
	Description string
	Logo        string
}

// Skill represents a skill category with items.
type Skill struct {
	Category string
	Items    []string
}

// ContactForm represents an incoming contact form submission.
type ContactForm struct {
	Name     string
	Email    string
	Subject  string
	Message  string
	Honeypot string
}

// PortfolioData holds all loaded portfolio data from JSON.
type PortfolioData struct {
	Sections []Section `json:"sections"`
}

// AllProjects returns a flat, deduplicated list of all projects across sections.
func (pd *PortfolioData) AllProjects() []Project {
	seen := make(map[string]bool)
	var result []Project
	for _, s := range pd.Sections {
		for _, p := range s.Projects {
			if !seen[p.Slug] {
				seen[p.Slug] = true
				result = append(result, p)
			}
		}
	}
	return result
}

// FilterTagsCSV returns the project's filter tags as a comma-separated string.
func (p Project) FilterTagsCSV() string {
	return strings.Join(p.FilterTags, ",")
}

// LoadPortfolioData reads and unmarshals portfolio data from a JSON file.
func LoadPortfolioData(path string) (*PortfolioData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading portfolio data: %w", err)
	}
	var pd PortfolioData
	if err := json.Unmarshal(data, &pd); err != nil {
		return nil, fmt.Errorf("parsing portfolio data: %w", err)
	}
	return &pd, nil
}

// GetProjectBySlug finds a project by its slug across all sections.
func (pd *PortfolioData) GetProjectBySlug(slug string) *Project {
	for _, s := range pd.Sections {
		for i := range s.Projects {
			if s.Projects[i].Slug == slug {
				return &s.Projects[i]
			}
		}
	}
	return nil
}

// GroupByYear takes a slice of projects and returns them grouped by year, sorted descending.
func GroupByYear(projects []Project) []YearGroup {
	m := make(map[int][]Project)
	for _, p := range projects {
		m[p.Year] = append(m[p.Year], p)
	}

	groups := make([]YearGroup, 0, len(m))
	for year, projs := range m {
		groups = append(groups, YearGroup{Year: year, Projects: projs})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Year > groups[j].Year
	})
	return groups
}

