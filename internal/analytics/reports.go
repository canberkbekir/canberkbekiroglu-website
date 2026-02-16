package analytics

import (
	"fmt"

	"github.com/canberkbekiroglu/website/internal/mailer"
)

// GetWeeklyStats returns aggregated visitor data for the weekly report.
func (db *DB) GetWeeklyStats() ([]mailer.VisitorStats, error) {
	rows, err := db.conn.Query(`
		SELECT v.ip_address, v.city, v.created_at, v.updated_at
		FROM visitors v
		ORDER BY v.created_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []mailer.VisitorStats
	for rows.Next() {
		var s mailer.VisitorStats
		if err := rows.Scan(&s.IP, &s.City, &s.FirstVisit, &s.LastVisit); err != nil {
			return nil, err
		}

		// Get page views for this visitor.
		pvRows, err := db.conn.Query(`
			SELECT page_path, SUM(view_count) as total
			FROM page_views pv
			JOIN visitors v ON v.id = pv.visitor_id
			WHERE v.ip_address = ?
			GROUP BY page_path
			ORDER BY total DESC
		`, s.IP)
		if err != nil {
			return nil, err
		}
		for pvRows.Next() {
			var pv mailer.PageViewStat
			if err := pvRows.Scan(&pv.PagePath, &pv.TotalViews); err != nil {
				pvRows.Close()
				return nil, err
			}
			s.PageViews = append(s.PageViews, pv)
		}
		pvRows.Close()

		stats = append(stats, s)
	}
	return stats, rows.Err()
}

// GetDailyVisitorCounts returns the number of unique visitors per day.
func (db *DB) GetDailyVisitorCounts() ([]mailer.DailyCount, error) {
	rows, err := db.conn.Query(`
		SELECT date, COUNT(DISTINCT visitor_id) as cnt
		FROM page_views
		GROUP BY date
		ORDER BY date
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []mailer.DailyCount
	for rows.Next() {
		var c mailer.DailyCount
		if err := rows.Scan(&c.Date, &c.Count); err != nil {
			return nil, err
		}
		counts = append(counts, c)
	}
	return counts, rows.Err()
}

// GetPagePopularity returns the total views per page path.
func (db *DB) GetPagePopularity() ([]mailer.PageCount, error) {
	rows, err := db.conn.Query(`
		SELECT page_path, SUM(view_count) as total
		FROM page_views
		GROUP BY page_path
		ORDER BY total DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []mailer.PageCount
	for rows.Next() {
		var p mailer.PageCount
		if err := rows.Scan(&p.PagePath, &p.TotalViews); err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}
	return pages, rows.Err()
}

// CleanAllData deletes all visitors and page views.
func (db *DB) CleanAllData() error {
	_, err := db.conn.Exec("DELETE FROM page_views")
	if err != nil {
		return err
	}
	_, err = db.conn.Exec("DELETE FROM visitors")
	return err
}

// ResetDailyPageViews deletes page_view records older than today.
func (db *DB) ResetDailyPageViews() error {
	_, err := db.conn.Exec("DELETE FROM page_views WHERE date < date('now')")
	return err
}

// CleanOldData deletes analytics data older than the given number of days.
func (db *DB) CleanOldData(days int) error {
	cutoff := fmt.Sprintf("date('now', '-%d days')", days)
	if _, err := db.conn.Exec("DELETE FROM page_views WHERE date < " + cutoff); err != nil {
		return err
	}
	if _, err := db.conn.Exec("DELETE FROM request_metrics WHERE date < " + cutoff); err != nil {
		return err
	}
	// Delete visitors with no page views left.
	_, err := db.conn.Exec("DELETE FROM visitors WHERE id NOT IN (SELECT DISTINCT visitor_id FROM page_views)")
	return err
}

// --- Dashboard query types ---

// DashboardDailyCount holds a date + count for chart data.
type DashboardDailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// DashboardPageCount holds page path + view count.
type DashboardPageCount struct {
	PagePath   string `json:"page_path"`
	TotalViews int    `json:"total_views"`
}

// DashboardCityCount holds city + visitor count.
type DashboardCityCount struct {
	City    string `json:"city"`
	Count   int    `json:"count"`
	Percent float64 `json:"percent"`
}

// DashboardTotalStats holds summary numbers.
type DashboardTotalStats struct {
	UniqueVisitors int     `json:"unique_visitors"`
	TotalViews     int     `json:"total_views"`
	AvgResponseMs  float64 `json:"avg_response_ms"`
	ErrorRate      float64 `json:"error_rate"`
}

// DashboardDailyAvg holds date + average value.
type DashboardDailyAvg struct {
	Date string  `json:"date"`
	Avg  float64 `json:"avg"`
}

// DashboardStatusCount holds status code category + count.
type DashboardStatusCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

// DashboardPageLatency holds path + avg latency + request count.
type DashboardPageLatency struct {
	Path     string  `json:"path"`
	AvgMs    float64 `json:"avg_ms"`
	ReqCount int     `json:"req_count"`
}

// DashboardHourlyCount holds hour + request count.
type DashboardHourlyCount struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// DashboardDailyRate holds date + rate percentage.
type DashboardDailyRate struct {
	Date string  `json:"date"`
	Rate float64 `json:"rate"`
}

// --- Dashboard query methods ---

func (db *DB) GetVisitorTrend(days int) ([]DashboardDailyCount, error) {
	rows, err := db.conn.Query(fmt.Sprintf(`
		SELECT date, COUNT(DISTINCT visitor_id) as cnt
		FROM page_views
		WHERE date >= date('now', '-%d days')
		GROUP BY date ORDER BY date
	`, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DashboardDailyCount
	for rows.Next() {
		var d DashboardDailyCount
		if err := rows.Scan(&d.Date, &d.Count); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (db *DB) GetPagePopularityByRange(days int) ([]DashboardPageCount, error) {
	rows, err := db.conn.Query(fmt.Sprintf(`
		SELECT page_path, SUM(view_count) as total
		FROM page_views
		WHERE date >= date('now', '-%d days')
		GROUP BY page_path ORDER BY total DESC LIMIT 10
	`, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DashboardPageCount
	for rows.Next() {
		var p DashboardPageCount
		if err := rows.Scan(&p.PagePath, &p.TotalViews); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (db *DB) GetGeographicDistribution(days int) ([]DashboardCityCount, error) {
	rows, err := db.conn.Query(fmt.Sprintf(`
		SELECT v.city, COUNT(DISTINCT v.id) as cnt
		FROM visitors v
		JOIN page_views pv ON pv.visitor_id = v.id
		WHERE pv.date >= date('now', '-%d days')
		GROUP BY v.city ORDER BY cnt DESC
	`, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DashboardCityCount
	total := 0
	for rows.Next() {
		var c DashboardCityCount
		if err := rows.Scan(&c.City, &c.Count); err != nil {
			return nil, err
		}
		total += c.Count
		out = append(out, c)
	}
	for i := range out {
		if total > 0 {
			out[i].Percent = float64(out[i].Count) / float64(total) * 100
		}
	}
	return out, rows.Err()
}

func (db *DB) GetTotalStats(days int) (DashboardTotalStats, error) {
	var s DashboardTotalStats

	db.conn.QueryRow(fmt.Sprintf(`
		SELECT COUNT(DISTINCT visitor_id) FROM page_views WHERE date >= date('now', '-%d days')
	`, days)).Scan(&s.UniqueVisitors)

	db.conn.QueryRow(fmt.Sprintf(`
		SELECT COALESCE(SUM(view_count), 0) FROM page_views WHERE date >= date('now', '-%d days')
	`, days)).Scan(&s.TotalViews)

	db.conn.QueryRow(fmt.Sprintf(`
		SELECT COALESCE(AVG(duration_ms), 0) FROM request_metrics WHERE date >= date('now', '-%d days')
	`, days)).Scan(&s.AvgResponseMs)

	var totalReqs, errorReqs int
	db.conn.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) FROM request_metrics WHERE date >= date('now', '-%d days')
	`, days)).Scan(&totalReqs)
	db.conn.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) FROM request_metrics WHERE date >= date('now', '-%d days') AND status_code >= 400
	`, days)).Scan(&errorReqs)
	if totalReqs > 0 {
		s.ErrorRate = float64(errorReqs) / float64(totalReqs) * 100
	}

	return s, nil
}

func (db *DB) GetAvgResponseTime(days int) ([]DashboardDailyAvg, error) {
	rows, err := db.conn.Query(fmt.Sprintf(`
		SELECT date, AVG(duration_ms) as avg_ms
		FROM request_metrics
		WHERE date >= date('now', '-%d days')
		GROUP BY date ORDER BY date
	`, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DashboardDailyAvg
	for rows.Next() {
		var d DashboardDailyAvg
		if err := rows.Scan(&d.Date, &d.Avg); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (db *DB) GetStatusCodeDistribution(days int) ([]DashboardStatusCount, error) {
	rows, err := db.conn.Query(fmt.Sprintf(`
		SELECT
			CASE
				WHEN status_code >= 200 AND status_code < 300 THEN '2xx'
				WHEN status_code >= 300 AND status_code < 400 THEN '3xx'
				WHEN status_code >= 400 AND status_code < 500 THEN '4xx'
				WHEN status_code >= 500 THEN '5xx'
				ELSE 'other'
			END as category,
			COUNT(*) as cnt
		FROM request_metrics
		WHERE date >= date('now', '-%d days')
		GROUP BY category ORDER BY category
	`, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DashboardStatusCount
	for rows.Next() {
		var s DashboardStatusCount
		if err := rows.Scan(&s.Category, &s.Count); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (db *DB) GetSlowestPages(days int) ([]DashboardPageLatency, error) {
	rows, err := db.conn.Query(fmt.Sprintf(`
		SELECT path, AVG(duration_ms) as avg_ms, COUNT(*) as cnt
		FROM request_metrics
		WHERE date >= date('now', '-%d days')
		GROUP BY path ORDER BY avg_ms DESC LIMIT 10
	`, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DashboardPageLatency
	for rows.Next() {
		var p DashboardPageLatency
		if err := rows.Scan(&p.Path, &p.AvgMs, &p.ReqCount); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (db *DB) GetRequestsPerHour(days int) ([]DashboardHourlyCount, error) {
	rows, err := db.conn.Query(fmt.Sprintf(`
		SELECT hour, COUNT(*) as cnt
		FROM request_metrics
		WHERE date >= date('now', '-%d days')
		GROUP BY hour ORDER BY hour
	`, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DashboardHourlyCount
	for rows.Next() {
		var h DashboardHourlyCount
		if err := rows.Scan(&h.Hour, &h.Count); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (db *DB) GetErrorRate(days int) ([]DashboardDailyRate, error) {
	rows, err := db.conn.Query(fmt.Sprintf(`
		SELECT date,
			CAST(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) AS REAL) / COUNT(*) * 100 as rate
		FROM request_metrics
		WHERE date >= date('now', '-%d days')
		GROUP BY date ORDER BY date
	`, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DashboardDailyRate
	for rows.Next() {
		var d DashboardDailyRate
		if err := rows.Scan(&d.Date, &d.Rate); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
