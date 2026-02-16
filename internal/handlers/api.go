package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/canberkbekiroglu/website/internal/mailer"
	"github.com/canberkbekiroglu/website/internal/templates"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// API holds API route handlers.
type API struct {
	renderer *templates.Renderer
	mailer   *mailer.Mailer // nil if SMTP is not configured
	csrf     *CSRFManager
}

// NewAPI creates a new API handler group.
func NewAPI(r *templates.Renderer, m *mailer.Mailer, csrf *CSRFManager) *API {
	return &API{renderer: r, mailer: m, csrf: csrf}
}

// HandleContactSubmit processes the contact form submission via HTMX.
func (a *API) HandleContactSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeContactResponse(w, a.renderer, false, "Invalid form data.")
		return
	}

	// CSRF validation.
	if !a.csrf.Validate(r) {
		w.WriteHeader(http.StatusForbidden)
		writeContactResponse(w, a.renderer, false, "Invalid or expired form token. Please reload the page and try again.")
		return
	}

	// Honeypot check — if filled, it's a bot.
	if r.FormValue("website") != "" {
		// Silently accept to not alert the bot.
		writeContactResponse(w, a.renderer, true, "Thank you! Your message has been sent.")
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	subject := strings.TrimSpace(r.FormValue("subject"))
	message := strings.TrimSpace(r.FormValue("message"))

	// Validation
	var errors []string
	if name == "" {
		errors = append(errors, "Name is required.")
	}
	if email == "" {
		errors = append(errors, "Email is required.")
	} else if !emailRegex.MatchString(email) {
		errors = append(errors, "Please enter a valid email address.")
	}
	if subject == "" {
		errors = append(errors, "Subject is required.")
	}
	if message == "" {
		errors = append(errors, "Message is required.")
	}

	if len(errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		writeContactResponse(w, a.renderer, false, strings.Join(errors, " "))
		return
	}

	log.Printf("Contact form submission: name=%q email=%q subject=%q", name, email, subject)

	if a.mailer != nil {
		if err := a.mailer.SendContactNotification(name, email, subject, message); err != nil {
			log.Printf("Failed to send contact notification email: %v", err)
		}
		if err := a.mailer.SendConfirmation(email, name); err != nil {
			log.Printf("Failed to send confirmation email to %s: %v", email, err)
		}
	}

	writeContactResponse(w, a.renderer, true, "Thank you! Your message has been sent successfully.")
}

func writeContactResponse(w http.ResponseWriter, r *templates.Renderer, success bool, message string) {
	data := map[string]any{
		"Success": success,
		"Message": message,
	}
	if err := r.Render(w, "partial_contact_form", data); err != nil {
		log.Printf("Error rendering contact response: %v", err)
		http.Error(w, message, http.StatusInternalServerError)
	}
}

// HandleHealthCheck returns a simple health check response.
func (a *API) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
