package analytics

import (
	"log"
	"time"

	"github.com/canberkbekiroglu/website/internal/mailer"
)

// StartScheduler launches background goroutines for daily resets and weekly reports.
func (db *DB) StartScheduler(m *mailer.Mailer) {
	go db.dailyResetLoop()
	go db.weeklyReportLoop(m)
	log.Println("Analytics scheduler started")
}

// dailyResetLoop runs at midnight every day to clean up data older than 180 days.
func (db *DB) dailyResetLoop() {
	for {
		now := time.Now()
		// Next midnight.
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		time.Sleep(time.Until(next))

		if err := db.CleanOldData(180); err != nil {
			log.Printf("Daily analytics cleanup failed: %v", err)
		} else {
			log.Println("Daily analytics cleanup completed (180-day retention)")
		}
	}
}

// weeklyReportLoop runs every Sunday at 23:59 to send the weekly analytics email.
func (db *DB) weeklyReportLoop(m *mailer.Mailer) {
	for {
		now := time.Now()
		next := nextSunday2359(now)
		time.Sleep(time.Until(next))

		if m == nil {
			log.Println("Weekly analytics: mailer not configured, skipping")
			continue
		}

		db.sendWeeklyReport(m)
	}
}

// nextSunday2359 returns the next Sunday at 23:59:00 from the given time.
func nextSunday2359(now time.Time) time.Time {
	daysUntilSunday := (7 - int(now.Weekday())) % 7
	if daysUntilSunday == 0 {
		// It's Sunday — check if 23:59 has passed.
		target := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())
		if now.Before(target) {
			return target
		}
		daysUntilSunday = 7
	}
	next := now.AddDate(0, 0, daysUntilSunday)
	return time.Date(next.Year(), next.Month(), next.Day(), 23, 59, 0, 0, now.Location())
}

func (db *DB) sendWeeklyReport(m *mailer.Mailer) {
	stats, err := db.GetWeeklyStats()
	if err != nil {
		log.Printf("Weekly report: failed to get stats: %v", err)
		return
	}

	// No visitors this week — skip.
	if len(stats) == 0 {
		log.Println("Weekly report: no visitors, skipping email")
		return
	}

	daily, err := db.GetDailyVisitorCounts()
	if err != nil {
		log.Printf("Weekly report: failed to get daily counts: %v", err)
		return
	}

	pages, err := db.GetPagePopularity()
	if err != nil {
		log.Printf("Weekly report: failed to get page popularity: %v", err)
		return
	}

	report := mailer.WeeklyReport{
		Stats:          stats,
		DailyVisitors:  daily,
		PagePopularity: pages,
	}

	if err := m.SendWeeklyAnalytics(report); err != nil {
		log.Printf("Weekly report: failed to send email: %v", err)
		return
	}

	log.Printf("Weekly analytics report sent (%d visitors)", len(stats))
}
