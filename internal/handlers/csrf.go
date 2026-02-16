package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

const (
	csrfTokenLength = 32
	csrfCookieName  = "_csrf"
	csrfFieldName   = "_csrf"
	csrfMaxAge      = 3600 // 1 hour
)

// CSRFManager manages CSRF token generation and validation.
type CSRFManager struct {
	mu     sync.Mutex
	tokens map[string]time.Time
}

// NewCSRFManager creates a new CSRF manager and starts cleanup.
func NewCSRFManager() *CSRFManager {
	cm := &CSRFManager{tokens: make(map[string]time.Time)}
	go cm.cleanup()
	return cm
}

// GenerateToken creates a new CSRF token and stores it.
func (cm *CSRFManager) GenerateToken() (string, error) {
	b := make([]byte, csrfTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	cm.mu.Lock()
	cm.tokens[token] = time.Now().Add(time.Duration(csrfMaxAge) * time.Second)
	cm.mu.Unlock()

	return token, nil
}

// ValidateToken checks that a token exists and hasn't expired, then removes it.
func (cm *CSRFManager) ValidateToken(token string) bool {
	if token == "" {
		return false
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	expiry, ok := cm.tokens[token]
	if !ok {
		return false
	}
	delete(cm.tokens, token)

	return time.Now().Before(expiry)
}

// SetCookie generates a CSRF token, stores it, and sets it as a cookie + returns it.
func (cm *CSRFManager) SetCookie(w http.ResponseWriter) (string, error) {
	token, err := cm.GenerateToken()
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   csrfMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})
	return token, nil
}

// Validate checks that the form field token matches the cookie token and both are valid.
func (cm *CSRFManager) Validate(r *http.Request) bool {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil {
		return false
	}
	formToken := r.FormValue(csrfFieldName)
	if formToken == "" {
		return false
	}

	// Both must match.
	if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(formToken)) != 1 {
		return false
	}

	// Token must exist in our store and not be expired.
	return cm.ValidateToken(cookie.Value)
}

// cleanup removes expired tokens periodically.
func (cm *CSRFManager) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		cm.mu.Lock()
		now := time.Now()
		for token, expiry := range cm.tokens {
			if now.After(expiry) {
				delete(cm.tokens, token)
			}
		}
		cm.mu.Unlock()
	}
}
