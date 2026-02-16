package analytics

import (
	"log"
	"time"
)

// RecordVisit logs a page visit for the given IP and path.
// If the visitor is new, their city is resolved via geolocation.
func (db *DB) RecordVisit(ip, pagePath string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get or create visitor.
	var visitorID int64
	err = tx.QueryRow("SELECT id FROM visitors WHERE ip_address = ?", ip).Scan(&visitorID)
	if err != nil {
		// New visitor — look up city.
		city := FetchCityFromAPI(ip)
		result, err := tx.Exec(
			"INSERT INTO visitors (ip_address, city) VALUES (?, ?)",
			ip, city,
		)
		if err != nil {
			return err
		}
		visitorID, _ = result.LastInsertId()
		log.Printf("New visitor: %s (%s)", ip, city)
	} else {
		// Update timestamp for returning visitor.
		tx.Exec("UPDATE visitors SET updated_at = datetime('now') WHERE id = ?", visitorID)
	}

	// Upsert page view: increment count if the row already exists for today.
	now := time.Now().Format("2006-01-02")
	_, err = tx.Exec(`
		INSERT INTO page_views (visitor_id, page_path, date)
		VALUES (?, ?, ?)
		ON CONFLICT(visitor_id, page_path, date)
		DO UPDATE SET view_count = view_count + 1
	`, visitorID, pagePath, now)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RecordMetrics inserts a request_metrics row.
func (db *DB) RecordMetrics(path, method string, statusCode, durationMs int) error {
	_, err := db.conn.Exec(`
		INSERT INTO request_metrics (path, method, status_code, duration_ms)
		VALUES (?, ?, ?, ?)
	`, path, method, statusCode, durationMs)
	return err
}
