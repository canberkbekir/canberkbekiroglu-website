package models

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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

// SampleResumeEntries returns hardcoded resume data for the demo.
func SampleResumeEntries() []ResumeEntry {
	return []ResumeEntry{
		{
			Company:     "Cloud Imperium Games",
			Role:        "Software Engineer",
			Period:      "2022 — Present",
			Description: "Working on Star Citizen, contributing to engine systems, gameplay features, and multiplayer infrastructure for one of the most ambitious games in development.",
			Logo:        "",
		},
		{
			Company:     "Kinetic Games",
			Role:        "Game Programmer",
			Period:      "2020 — 2022",
			Description: "Developed gameplay systems and networking code for Phasmophobia, helping scale the game from indie project to millions of concurrent players.",
			Logo:        "",
		},
		{
			Company:     "Freelance",
			Role:        "Full Stack Developer",
			Period:      "2018 — 2020",
			Description: "Built web applications, tools, and interactive experiences for various clients. Specialized in real-time applications and data visualization.",
			Logo:        "",
		},
	}
}

// SampleSkills returns hardcoded skills data for the demo.
func SampleSkills() []Skill {
	return []Skill{
		{Category: "Languages", Items: []string{"Go", "C#", "C++", "TypeScript", "Python"}},
		{Category: "Frameworks", Items: []string{"Chi", "Unity", "Unreal Engine", "React", ".NET"}},
		{Category: "Tools", Items: []string{"Git", "Docker", "Linux", "CI/CD", "PostgreSQL"}},
		{Category: "Domains", Items: []string{"Game Development", "Web Development", "Networking", "Systems Programming"}},
	}
}
