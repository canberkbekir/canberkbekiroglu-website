package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/canberkbekiroglu/website/internal/templates"
)

// API holds API route handlers.
type API struct {
	renderer *templates.Renderer
}

// NewAPI creates a new API handler group.
func NewAPI(r *templates.Renderer) *API {
	return &API{renderer: r}
}

// HandleContactSubmit processes the contact form submission via HTMX.
func (a *API) HandleContactSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeContactResponse(w, a.renderer, false, "Invalid form data.")
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
	} else if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
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

	// In a real app, you'd send an email or store in a database here.
	log.Printf("Contact form submission: name=%q email=%q subject=%q message=%q", name, email, subject, message)

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
