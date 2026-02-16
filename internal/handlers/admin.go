package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/canberkbekiroglu/website/internal/admin"
	"github.com/canberkbekiroglu/website/internal/analytics"
	"github.com/canberkbekiroglu/website/internal/templates"
	"github.com/go-chi/chi/v5"
)

const adminSessionCookie = "admin_session"

// Admin holds admin panel route handlers.
type Admin struct {
	renderer    *templates.Renderer
	db          *admin.DB
	analyticsDB *analytics.DB
	csrf        *CSRFManager
	passHash    string // bcrypt hash of ADMIN_PASSWORD
}

// NewAdmin creates a new Admin handler group.
func NewAdmin(r *templates.Renderer, db *admin.DB, analyticsDB *analytics.DB, csrf *CSRFManager, passHash string) *Admin {
	return &Admin{renderer: r, db: db, analyticsDB: analyticsDB, csrf: csrf, passHash: passHash}
}

// RequireAdmin is middleware that checks for a valid admin session.
func (a *Admin) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(adminSessionCookie)
		if err != nil || !a.db.ValidateSession(cookie.Value) {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ─── Auth ───

func (a *Admin) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	token, _ := a.csrf.SetCookie(w)
	data := map[string]any{
		"Title":     "Login",
		"CSRFToken": token,
		"ActiveNav": "",
	}
	if err := a.renderer.Render(w, "admin_login", data); err != nil {
		log.Printf("admin login render: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

func (a *Admin) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if !a.csrf.Validate(r) {
		token, _ := a.csrf.SetCookie(w)
		a.renderer.Render(w, "admin_login", map[string]any{
			"Title": "Login", "CSRFToken": token, "ActiveNav": "",
			"Error": "Invalid form token. Please try again.",
		})
		return
	}

	password := r.FormValue("password")
	if !admin.CheckPassword(a.passHash, password) {
		token, _ := a.csrf.SetCookie(w)
		w.WriteHeader(http.StatusUnauthorized)
		a.renderer.Render(w, "admin_login", map[string]any{
			"Title": "Login", "CSRFToken": token, "ActiveNav": "",
			"Error": "Invalid password.",
		})
		return
	}

	sessionToken, err := a.db.CreateSession()
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookie,
		Value:    sessionToken,
		Path:     "/",
		MaxAge:   86400, // 24h
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/admin", http.StatusFound)
}

func (a *Admin) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(adminSessionCookie)
	if err == nil {
		a.db.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   adminSessionCookie,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusFound)
}

// ─── Dashboard ───

func (a *Admin) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	projCount, _ := a.db.ProjectCount()
	secCount, _ := a.db.SectionCount()

	stats, _ := a.analyticsDB.GetTotalStats(7)

	data := map[string]any{
		"Title":          "Dashboard",
		"ActiveNav":      "dashboard",
		"ProjectCount":   projCount,
		"SectionCount":   secCount,
		"UniqueVisitors": stats.UniqueVisitors,
		"TotalViews":     stats.TotalViews,
		"AvgResponseMs":  stats.AvgResponseMs,
		"ErrorRate":      stats.ErrorRate,
	}
	if err := a.renderer.Render(w, "admin_dashboard", data); err != nil {
		log.Printf("admin dashboard render: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// ─── Analytics API ───

func (a *Admin) parseDays(r *http.Request) int {
	switch r.URL.Query().Get("range") {
	case "30d":
		return 30
	case "90d":
		return 90
	case "6m":
		return 180
	default:
		return 7
	}
}

func (a *Admin) HandleAnalyticsVisitors(w http.ResponseWriter, r *http.Request) {
	days := a.parseDays(r)

	trend, _ := a.analyticsDB.GetVisitorTrend(days)
	pages, _ := a.analyticsDB.GetPagePopularityByRange(days)
	geo, _ := a.analyticsDB.GetGeographicDistribution(days)
	stats, _ := a.analyticsDB.GetTotalStats(days)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"trend": trend,
		"pages": pages,
		"geo":   geo,
		"stats": stats,
	})
}

func (a *Admin) HandleAnalyticsPerformance(w http.ResponseWriter, r *http.Request) {
	days := a.parseDays(r)

	avgResp, _ := a.analyticsDB.GetAvgResponseTime(days)
	statusDist, _ := a.analyticsDB.GetStatusCodeDistribution(days)
	slowest, _ := a.analyticsDB.GetSlowestPages(days)
	hourly, _ := a.analyticsDB.GetRequestsPerHour(days)
	errorRate, _ := a.analyticsDB.GetErrorRate(days)
	stats, _ := a.analyticsDB.GetTotalStats(days)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"avg_response": avgResp,
		"status_codes": statusDist,
		"slowest":      slowest,
		"hourly":       hourly,
		"error_rate":   errorRate,
		"stats":        stats,
	})
}

// ─── Projects CRUD ───

func (a *Admin) HandleProjectsList(w http.ResponseWriter, r *http.Request) {
	projects, err := a.db.AllProjects()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"Title":     "Projects",
		"ActiveNav": "projects",
		"Projects":  projects,
	}
	if err := a.renderer.Render(w, "admin_projects", data); err != nil {
		log.Printf("admin projects render: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

func (a *Admin) HandleProjectForm(w http.ResponseWriter, r *http.Request) {
	token, _ := a.csrf.SetCookie(w)
	sections, _ := a.db.AllSections()
	filterTags, _ := a.db.AllFilterTags()

	data := map[string]any{
		"Title":      "New Project",
		"ActiveNav":  "projects",
		"CSRFToken":  token,
		"Sections":   sections,
		"FilterTags": filterTags,
		"IsEdit":     false,
	}

	// Check if editing.
	idStr := chi.URLParam(r, "id")
	if idStr != "" {
		id, _ := strconv.Atoi(idStr)
		project, err := a.db.GetProject(id)
		if err != nil || project == nil {
			http.NotFound(w, r)
			return
		}
		data["Project"] = project
		data["IsEdit"] = true
		data["Title"] = "Edit Project"
	}

	if err := a.renderer.Render(w, "admin_project_form", data); err != nil {
		log.Printf("admin project form render: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

func (a *Admin) HandleProjectCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !a.csrf.Validate(r) {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	p := a.parseProjectForm(r)
	if _, err := a.db.CreateProject(p); err != nil {
		log.Printf("create project: %v", err)
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/projects", http.StatusFound)
}

func (a *Admin) HandleProjectUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !a.csrf.Validate(r) {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	p := a.parseProjectForm(r)
	p.ID = id

	if err := a.db.UpdateProject(p); err != nil {
		log.Printf("update project: %v", err)
		http.Error(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/projects", http.StatusFound)
}

func (a *Admin) HandleProjectDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := a.db.DeleteProject(id); err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (a *Admin) parseProjectForm(r *http.Request) *admin.Project {
	year, _ := strconv.Atoi(r.FormValue("year"))
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))

	var tags []string
	if t := strings.TrimSpace(r.FormValue("tags")); t != "" {
		for _, tag := range strings.Split(t, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	var sectionIDs []int
	for _, sid := range r.Form["sections"] {
		id, _ := strconv.Atoi(sid)
		if id > 0 {
			sectionIDs = append(sectionIDs, id)
		}
	}

	filterTags := r.Form["filter_tags"]

	return &admin.Project{
		Title:       strings.TrimSpace(r.FormValue("title")),
		Slug:        strings.TrimSpace(r.FormValue("slug")),
		Year:        year,
		Description: strings.TrimSpace(r.FormValue("description")),
		Image:       strings.TrimSpace(r.FormValue("image")),
		HoverImage:  strings.TrimSpace(r.FormValue("hover_image")),
		Category:    strings.TrimSpace(r.FormValue("category")),
		Company:     strings.TrimSpace(r.FormValue("company")),
		Role:        strings.TrimSpace(r.FormValue("role")),
		ExternalURL: strings.TrimSpace(r.FormValue("external_url")),
		Highlighted: r.FormValue("highlighted") == "on",
		SortOrder:   sortOrder,
		Tags:        tags,
		FilterTags:  filterTags,
		SectionIDs:  sectionIDs,
	}
}

// ─── Sections CRUD ───

func (a *Admin) HandleSectionsList(w http.ResponseWriter, r *http.Request) {
	token, _ := a.csrf.SetCookie(w)
	sections, err := a.db.AllSections()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"Title":     "Sections",
		"ActiveNav": "sections",
		"Sections":  sections,
		"CSRFToken": token,
	}
	if err := a.renderer.Render(w, "admin_sections", data); err != nil {
		log.Printf("admin sections render: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

func (a *Admin) HandleSectionCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !a.csrf.Validate(r) {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))

	if name == "" || slug == "" {
		http.Error(w, "Name and slug required", http.StatusBadRequest)
		return
	}

	if _, err := a.db.CreateSection(name, slug, sortOrder); err != nil {
		http.Error(w, "Failed to create section", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/sections", http.StatusFound)
}

func (a *Admin) HandleSectionUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !a.csrf.Validate(r) {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	name := strings.TrimSpace(r.FormValue("name"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))

	if err := a.db.UpdateSection(id, name, slug, sortOrder); err != nil {
		http.Error(w, "Failed to update section", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/sections", http.StatusFound)
}

func (a *Admin) HandleSectionDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := a.db.DeleteSection(id); err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ─── Filter Tags CRUD ───

func (a *Admin) HandleFilterTagsList(w http.ResponseWriter, r *http.Request) {
	token, _ := a.csrf.SetCookie(w)
	tags, err := a.db.AllFilterTags()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"Title":      "Filter Tags",
		"ActiveNav":  "filters",
		"FilterTags": tags,
		"CSRFToken":  token,
	}
	if err := a.renderer.Render(w, "admin_filters", data); err != nil {
		log.Printf("admin filters render: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

func (a *Admin) HandleFilterTagCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !a.csrf.Validate(r) {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	slug := strings.TrimSpace(r.FormValue("slug"))
	label := strings.TrimSpace(r.FormValue("label"))
	isGold := r.FormValue("is_gold") == "on"
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))

	if slug == "" || label == "" {
		http.Error(w, "Slug and label required", http.StatusBadRequest)
		return
	}

	if _, err := a.db.CreateFilterTag(slug, label, isGold, sortOrder); err != nil {
		http.Error(w, "Failed to create filter tag", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/filters", http.StatusFound)
}

func (a *Admin) HandleFilterTagUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !a.csrf.Validate(r) {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	label := strings.TrimSpace(r.FormValue("label"))
	isGold := r.FormValue("is_gold") == "on"
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))

	if err := a.db.UpdateFilterTag(id, slug, label, isGold, sortOrder); err != nil {
		http.Error(w, "Failed to update filter tag", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/filters", http.StatusFound)
}

func (a *Admin) HandleFilterTagDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := a.db.DeleteFilterTag(id); err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
