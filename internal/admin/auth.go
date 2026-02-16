package admin

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionDuration = 24 * time.Hour

// CheckPassword verifies a plaintext password against a bcrypt hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// HashPassword returns the bcrypt hash of a plaintext password.
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

// CreateSession generates a random session token and stores it.
func (db *DB) CreateSession() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	expiresAt := time.Now().Add(sessionDuration)

	_, err := db.conn.Exec(
		"INSERT INTO sessions (token, expires_at) VALUES (?, ?)",
		token, expiresAt.UTC().Format("2006-01-02 15:04:05"),
	)
	return token, err
}

// ValidateSession returns true if the token exists and hasn't expired.
func (db *DB) ValidateSession(token string) bool {
	if token == "" {
		return false
	}
	var expiresAt string
	err := db.conn.QueryRow("SELECT expires_at FROM sessions WHERE token = ?", token).Scan(&expiresAt)
	if err != nil {
		return false
	}
	// Try multiple formats since Go sqlite driver may store as RFC3339.
	var t time.Time
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	} {
		if parsed, err := time.Parse(layout, expiresAt); err == nil {
			t = parsed
			break
		}
	}
	if t.IsZero() {
		return false
	}
	return time.Now().UTC().Before(t)
}

// DeleteSession removes a session token.
func (db *DB) DeleteSession(token string) error {
	_, err := db.conn.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

// CleanExpiredSessions removes all expired session rows.
func (db *DB) CleanExpiredSessions() {
	_, err := db.conn.Exec("DELETE FROM sessions WHERE expires_at < datetime('now')")
	if err != nil {
		log.Printf("Failed to clean expired sessions: %v", err)
	}
}

// StartSessionCleanup launches a background goroutine to periodically clean expired sessions.
func (db *DB) StartSessionCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			db.CleanExpiredSessions()
		}
	}()
}
