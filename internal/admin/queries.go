package admin

import (
	"database/sql"

	"github.com/canberkbekiroglu/website/internal/models"
)

// --- Sections ---

// Section represents a DB section row.
type Section struct {
	ID        int
	Name      string
	Slug      string
	SortOrder int
}

func (db *DB) AllSections() ([]Section, error) {
	rows, err := db.conn.Query("SELECT id, name, slug, sort_order FROM sections ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Section
	for rows.Next() {
		var s Section
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (db *DB) GetSection(id int) (*Section, error) {
	var s Section
	err := db.conn.QueryRow("SELECT id, name, slug, sort_order FROM sections WHERE id = ?", id).
		Scan(&s.ID, &s.Name, &s.Slug, &s.SortOrder)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (db *DB) CreateSection(name, slug string, sortOrder int) (int64, error) {
	res, err := db.conn.Exec("INSERT INTO sections (name, slug, sort_order) VALUES (?, ?, ?)", name, slug, sortOrder)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) UpdateSection(id int, name, slug string, sortOrder int) error {
	_, err := db.conn.Exec("UPDATE sections SET name=?, slug=?, sort_order=? WHERE id=?", name, slug, sortOrder, id)
	return err
}

func (db *DB) DeleteSection(id int) error {
	_, err := db.conn.Exec("DELETE FROM sections WHERE id=?", id)
	return err
}

// --- Filter Tags ---

// FilterTag represents a DB filter_tags row.
type FilterTag struct {
	ID        int
	Slug      string
	Label     string
	IsGold    bool
	SortOrder int
}

func (db *DB) AllFilterTags() ([]FilterTag, error) {
	rows, err := db.conn.Query("SELECT id, slug, label, is_gold, sort_order FROM filter_tags ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []FilterTag
	for rows.Next() {
		var f FilterTag
		if err := rows.Scan(&f.ID, &f.Slug, &f.Label, &f.IsGold, &f.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (db *DB) GetFilterTag(id int) (*FilterTag, error) {
	var f FilterTag
	err := db.conn.QueryRow("SELECT id, slug, label, is_gold, sort_order FROM filter_tags WHERE id = ?", id).
		Scan(&f.ID, &f.Slug, &f.Label, &f.IsGold, &f.SortOrder)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &f, err
}

func (db *DB) CreateFilterTag(slug, label string, isGold bool, sortOrder int) (int64, error) {
	res, err := db.conn.Exec("INSERT INTO filter_tags (slug, label, is_gold, sort_order) VALUES (?, ?, ?, ?)", slug, label, isGold, sortOrder)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) UpdateFilterTag(id int, slug, label string, isGold bool, sortOrder int) error {
	_, err := db.conn.Exec("UPDATE filter_tags SET slug=?, label=?, is_gold=?, sort_order=? WHERE id=?", slug, label, isGold, sortOrder, id)
	return err
}

func (db *DB) DeleteFilterTag(id int) error {
	_, err := db.conn.Exec("DELETE FROM filter_tags WHERE id=?", id)
	return err
}

// --- Projects ---

// Project represents a DB project row.
type Project struct {
	ID          int
	Title       string
	Slug        string
	Year        int
	Description string
	Image       string
	HoverImage  string
	Category    string
	Company     string
	Role        string
	ExternalURL string
	Highlighted bool
	SortOrder   int
	Tags        []string // tech tags
	FilterTags  []string // filter tag slugs
	SectionIDs  []int    // associated section IDs
}

func (db *DB) AllProjects() ([]Project, error) {
	rows, err := db.conn.Query(`
		SELECT id, title, slug, year, description, image, hover_image,
		       category, company, role, external_url, highlighted, sort_order
		FROM projects ORDER BY sort_order, year DESC, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load tags, filter tags, sections for each project.
	for i := range out {
		if err := db.loadProjectRelations(&out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (db *DB) GetProject(id int) (*Project, error) {
	row := db.conn.QueryRow(`
		SELECT id, title, slug, year, description, image, hover_image,
		       category, company, role, external_url, highlighted, sort_order
		FROM projects WHERE id = ?
	`, id)

	var p Project
	err := row.Scan(&p.ID, &p.Title, &p.Slug, &p.Year, &p.Description,
		&p.Image, &p.HoverImage, &p.Category, &p.Company, &p.Role,
		&p.ExternalURL, &p.Highlighted, &p.SortOrder)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := db.loadProjectRelations(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (db *DB) CreateProject(p *Project) (int64, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		INSERT INTO projects (title, slug, year, description, image, hover_image,
		                      category, company, role, external_url, highlighted, sort_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, p.Title, p.Slug, p.Year, p.Description, p.Image, p.HoverImage,
		p.Category, p.Company, p.Role, p.ExternalURL, p.Highlighted, p.SortOrder)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()

	if err := saveProjectRelations(tx, id, p.Tags, p.FilterTags, p.SectionIDs); err != nil {
		return 0, err
	}

	return id, tx.Commit()
}

func (db *DB) UpdateProject(p *Project) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE projects SET title=?, slug=?, year=?, description=?, image=?, hover_image=?,
		                    category=?, company=?, role=?, external_url=?, highlighted=?, sort_order=?
		WHERE id=?
	`, p.Title, p.Slug, p.Year, p.Description, p.Image, p.HoverImage,
		p.Category, p.Company, p.Role, p.ExternalURL, p.Highlighted, p.SortOrder, p.ID)
	if err != nil {
		return err
	}

	// Clear existing relations and re-insert.
	tx.Exec("DELETE FROM project_tags WHERE project_id=?", p.ID)
	tx.Exec("DELETE FROM project_filter_tags WHERE project_id=?", p.ID)
	tx.Exec("DELETE FROM project_sections WHERE project_id=?", p.ID)

	if err := saveProjectRelations(tx, int64(p.ID), p.Tags, p.FilterTags, p.SectionIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) DeleteProject(id int) error {
	_, err := db.conn.Exec("DELETE FROM projects WHERE id=?", id)
	return err
}

// ProjectCount returns the total number of projects.
func (db *DB) ProjectCount() (int, error) {
	var n int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM projects").Scan(&n)
	return n, err
}

// SectionCount returns the total number of sections.
func (db *DB) SectionCount() (int, error) {
	var n int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM sections").Scan(&n)
	return n, err
}

// --- Portfolio Data for public site ---

// LoadPortfolioData builds a *models.PortfolioData from the database.
func (db *DB) LoadPortfolioData() (*models.PortfolioData, error) {
	sections, err := db.AllSections()
	if err != nil {
		return nil, err
	}

	pd := &models.PortfolioData{}
	for _, sec := range sections {
		ms := models.Section{
			Name: sec.Name,
			Slug: sec.Slug,
		}

		rows, err := db.conn.Query(`
			SELECT p.id, p.title, p.slug, p.year, p.description, p.image, p.hover_image,
			       p.category, p.company, p.role, p.external_url, p.highlighted, p.sort_order
			FROM projects p
			JOIN project_sections ps ON ps.project_id = p.id
			WHERE ps.section_id = ?
			ORDER BY p.sort_order, p.year DESC, p.id
		`, sec.ID)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			ap, err := scanProject(rows)
			if err != nil {
				rows.Close()
				return nil, err
			}
			if err := db.loadProjectRelations(&ap); err != nil {
				rows.Close()
				return nil, err
			}
			ms.Projects = append(ms.Projects, models.Project{
				Title:       ap.Title,
				Slug:        ap.Slug,
				Year:        ap.Year,
				Description: ap.Description,
				Image:       ap.Image,
				HoverImage:  ap.HoverImage,
				Tags:        ap.Tags,
				Category:    ap.Category,
				Company:     ap.Company,
				Role:        ap.Role,
				ExternalURL: ap.ExternalURL,
				Highlighted: ap.Highlighted,
				FilterTags:  ap.FilterTags,
			})
		}
		rows.Close()

		pd.Sections = append(pd.Sections, ms)
	}

	return pd, nil
}

// LoadFilterTagsForTemplate returns filter tags formatted for the home template.
func (db *DB) LoadFilterTagsForTemplate() ([]models.FilterTagView, error) {
	tags, err := db.AllFilterTags()
	if err != nil {
		return nil, err
	}
	var out []models.FilterTagView
	for _, t := range tags {
		out = append(out, models.FilterTagView{
			Slug:   t.Slug,
			Label:  t.Label,
			IsGold: t.IsGold,
		})
	}
	return out, nil
}

// HasProjects returns true if the projects table has any rows.
func (db *DB) HasProjects() (bool, error) {
	var n int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM projects").Scan(&n)
	return n > 0, err
}

// --- Helpers ---

type scannable interface {
	Scan(dest ...any) error
}

func scanProject(row scannable) (Project, error) {
	var p Project
	err := row.Scan(&p.ID, &p.Title, &p.Slug, &p.Year, &p.Description,
		&p.Image, &p.HoverImage, &p.Category, &p.Company, &p.Role,
		&p.ExternalURL, &p.Highlighted, &p.SortOrder)
	return p, err
}

func (db *DB) loadProjectRelations(p *Project) error {
	// Tags
	rows, err := db.conn.Query("SELECT tag FROM project_tags WHERE project_id=? ORDER BY sort_order, id", p.ID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			rows.Close()
			return err
		}
		p.Tags = append(p.Tags, tag)
	}
	rows.Close()

	// Filter tags
	rows, err = db.conn.Query(`
		SELECT ft.slug FROM filter_tags ft
		JOIN project_filter_tags pft ON pft.filter_tag_id = ft.id
		WHERE pft.project_id=?
		ORDER BY ft.sort_order, ft.id
	`, p.ID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			rows.Close()
			return err
		}
		p.FilterTags = append(p.FilterTags, slug)
	}
	rows.Close()

	// Section IDs
	rows, err = db.conn.Query("SELECT section_id FROM project_sections WHERE project_id=?", p.ID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var sid int
		if err := rows.Scan(&sid); err != nil {
			rows.Close()
			return err
		}
		p.SectionIDs = append(p.SectionIDs, sid)
	}
	rows.Close()

	return nil
}

func saveProjectRelations(tx *sql.Tx, projectID int64, tags []string, filterTagSlugs []string, sectionIDs []int) error {
	for i, tag := range tags {
		if _, err := tx.Exec("INSERT INTO project_tags (project_id, tag, sort_order) VALUES (?, ?, ?)", projectID, tag, i); err != nil {
			return err
		}
	}

	for _, slug := range filterTagSlugs {
		_, err := tx.Exec(`
			INSERT INTO project_filter_tags (project_id, filter_tag_id)
			SELECT ?, id FROM filter_tags WHERE slug = ?
		`, projectID, slug)
		if err != nil {
			return err
		}
	}

	for _, sid := range sectionIDs {
		if _, err := tx.Exec("INSERT INTO project_sections (project_id, section_id) VALUES (?, ?)", projectID, sid); err != nil {
			return err
		}
	}

	return nil
}
