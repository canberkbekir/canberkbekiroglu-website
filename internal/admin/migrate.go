package admin

import (
	"encoding/json"
	"log"
	"os"
)

// jsonPortfolioData mirrors the JSON structure.
type jsonPortfolioData struct {
	Sections []jsonSection `json:"sections"`
}

type jsonSection struct {
	Name     string        `json:"name"`
	Slug     string        `json:"slug"`
	Projects []jsonProject `json:"projects"`
}

type jsonProject struct {
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

// MigrateFromJSON reads projects.json and imports data into the admin database.
// It only runs if the projects table is empty.
func (db *DB) MigrateFromJSON(jsonPath string) error {
	hasData, err := db.HasProjects()
	if err != nil {
		return err
	}
	if hasData {
		log.Println("Admin DB already has projects, skipping JSON migration")
		return nil
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Printf("No projects.json found at %s, skipping migration", jsonPath)
		return nil
	}

	var pd jsonPortfolioData
	if err := json.Unmarshal(data, &pd); err != nil {
		return err
	}

	// Collect unique filter tags from all projects.
	filterTagSet := make(map[string]bool)
	for _, sec := range pd.Sections {
		for _, p := range sec.Projects {
			for _, ft := range p.FilterTags {
				filterTagSet[ft] = true
			}
		}
	}

	// Insert default filter tags with sensible ordering.
	defaultTags := []struct {
		slug  string
		label string
		gold  bool
	}{
		{"highlighted", "HIGHLIGHTED", true},
		{"awards", "AWARDS & NOMINEES", true},
		{"all", "ALL", false},
		{"games", "GAMES", false},
		{"web", "WEB", false},
		{"android", "ANDROID", false},
		{"windows", "WINDOWS", false},
		{"others", "OTHERS", false},
	}
	for i, dt := range defaultTags {
		db.CreateFilterTag(dt.slug, dt.label, dt.gold, i)
	}
	// Insert any filter tags from projects that aren't in the defaults.
	for ft := range filterTagSet {
		found := false
		for _, dt := range defaultTags {
			if dt.slug == ft {
				found = true
				break
			}
		}
		if !found {
			db.CreateFilterTag(ft, ft, false, len(defaultTags))
		}
	}

	// Create sections.
	sectionIDMap := make(map[string]int) // slug -> DB id
	for i, sec := range pd.Sections {
		id, err := db.CreateSection(sec.Name, sec.Slug, i)
		if err != nil {
			return err
		}
		sectionIDMap[sec.Slug] = int(id)
	}

	// Create projects (deduplicate by slug).
	projectIDMap := make(map[string]int) // slug -> DB id
	for _, sec := range pd.Sections {
		secID := sectionIDMap[sec.Slug]
		for _, jp := range sec.Projects {
			if pid, exists := projectIDMap[jp.Slug]; exists {
				// Just link to this section.
				db.conn.Exec("INSERT OR IGNORE INTO project_sections (project_id, section_id) VALUES (?, ?)", pid, secID)
				continue
			}

			p := &Project{
				Title:       jp.Title,
				Slug:        jp.Slug,
				Year:        jp.Year,
				Description: jp.Description,
				Image:       jp.Image,
				HoverImage:  jp.HoverImage,
				Category:    jp.Category,
				Company:     jp.Company,
				Role:        jp.Role,
				ExternalURL: jp.ExternalURL,
				Highlighted: jp.Highlighted,
				Tags:        jp.Tags,
				FilterTags:  jp.FilterTags,
				SectionIDs:  []int{secID},
			}

			id, err := db.CreateProject(p)
			if err != nil {
				return err
			}
			projectIDMap[jp.Slug] = int(id)
		}
	}

	log.Printf("Migrated %d projects and %d sections from JSON", len(projectIDMap), len(sectionIDMap))
	return nil
}
