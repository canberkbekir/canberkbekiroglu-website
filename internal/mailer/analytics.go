package mailer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"mime/multipart"
	"net/smtp"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
)

// VisitorStats holds aggregated data for a single visitor.
type VisitorStats struct {
	IP         string
	City       string
	FirstVisit string
	LastVisit  string
	PageViews  []PageViewStat
}

// PageViewStat holds page view data for a visitor.
type PageViewStat struct {
	PagePath   string
	TotalViews int
}

// DailyCount holds unique visitor count for a date.
type DailyCount struct {
	Date  string
	Count int
}

// PageCount holds total views for a page path.
type PageCount struct {
	PagePath   string
	TotalViews int
}

// WeeklyReport bundles all analytics data for the weekly email.
type WeeklyReport struct {
	Stats          []VisitorStats
	DailyVisitors  []DailyCount
	PagePopularity []PageCount
}

// SendWeeklyAnalytics builds and sends the weekly analytics email with an Excel attachment.
func (m *Mailer) SendWeeklyAnalytics(report WeeklyReport) error {
	totalVisitors := len(report.Stats)
	totalViews := 0
	for _, p := range report.PagePopularity {
		totalViews += p.TotalViews
	}
	mostPopular := "N/A"
	if len(report.PagePopularity) > 0 {
		mostPopular = report.PagePopularity[0].PagePath
	}

	subject := fmt.Sprintf("[Website Analytics] Weekly Report - %d Visitors", totalVisitors)

	body := buildAnalyticsHTML(report, totalVisitors, totalViews, mostPopular)

	xlsxData, err := generateExcel(report)
	if err != nil {
		return fmt.Errorf("generating excel: %w", err)
	}

	return m.sendWithAttachment(
		m.cfg.ContactEmail,
		subject,
		body,
		"weekly-analytics.xlsx",
		xlsxData,
	)
}

// sendWithAttachment sends an HTML email with a binary file attachment.
func (m *Mailer) sendWithAttachment(to, subject, htmlBody, attachName string, attachData []byte) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Headers
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n\r\n",
		m.cfg.FromAddress, to, subject, writer.Boundary())
	buf.Reset()
	buf.WriteString(headers)

	// We need to manually write multipart parts.
	boundary := writer.Boundary()
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\nContent-Transfer-Encoding: 7bit\r\n\r\n")
	buf.WriteString(htmlBody)
	buf.WriteString("\r\n")

	// Attachment part.
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString(fmt.Sprintf("Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet\r\n"))
	buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachName))
	buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

	encoded := base64.StdEncoding.EncodeToString(attachData)
	// Wrap at 76 chars per line.
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		buf.WriteString(encoded[i:end])
		buf.WriteString("\r\n")
	}

	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return smtp.SendMail(m.addr, m.auth, m.cfg.FromAddress, []string{to}, buf.Bytes())
}

func buildAnalyticsHTML(report WeeklyReport, totalVisitors, totalViews int, mostPopular string) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#161618;color:#d8d8d8;font-family:Inter,Arial,sans-serif;">
<div style="max-width:700px;margin:0 auto;padding:32px;">

<h1 style="color:#e2b384;margin-top:0;font-size:24px;">Weekly Analytics Report</h1>
`)

	// Summary stats.
	b.WriteString(fmt.Sprintf(`
<table style="width:100%%;border-collapse:collapse;margin-bottom:32px;">
  <tr>
    <td style="padding:16px;text-align:center;background:#1a1a1c;border-radius:8px;">
      <div style="font-size:32px;color:#e2b384;font-weight:bold;">%d</div>
      <div style="color:#a0a0a0;font-size:12px;margin-top:4px;">Unique Visitors</div>
    </td>
    <td style="width:12px;"></td>
    <td style="padding:16px;text-align:center;background:#1a1a1c;border-radius:8px;">
      <div style="font-size:32px;color:#e2b384;font-weight:bold;">%d</div>
      <div style="color:#a0a0a0;font-size:12px;margin-top:4px;">Total Page Views</div>
    </td>
    <td style="width:12px;"></td>
    <td style="padding:16px;text-align:center;background:#1a1a1c;border-radius:8px;">
      <div style="font-size:16px;color:#e2b384;font-weight:bold;">%s</div>
      <div style="color:#a0a0a0;font-size:12px;margin-top:4px;">Most Popular</div>
    </td>
  </tr>
</table>
`, totalVisitors, totalViews, html.EscapeString(mostPopular)))

	// Page popularity bar chart.
	b.WriteString(`<h2 style="color:#e2b384;font-size:18px;margin-bottom:12px;">Page Popularity</h2>
<div style="background:#1a1a1c;border-radius:8px;padding:16px;margin-bottom:32px;">
`)
	maxViews := 0
	for _, p := range report.PagePopularity {
		if p.TotalViews > maxViews {
			maxViews = p.TotalViews
		}
	}
	for _, p := range report.PagePopularity {
		pct := 0
		if maxViews > 0 {
			pct = (p.TotalViews * 100) / maxViews
		}
		if pct < 3 {
			pct = 3
		}
		b.WriteString(fmt.Sprintf(`
  <div style="margin-bottom:8px;">
    <div style="color:#d8d8d8;font-size:13px;margin-bottom:4px;">%s</div>
    <div style="display:flex;align-items:center;">
      <div style="background:#e2b384;height:20px;width:%d%%;border-radius:4px;min-width:20px;"></div>
      <span style="color:#a0a0a0;font-size:12px;margin-left:8px;">%d views</span>
    </div>
  </div>
`, html.EscapeString(p.PagePath), pct, p.TotalViews))
	}
	b.WriteString(`</div>`)

	// Daily visitors bar chart (vertical bars — email-client-safe).
	b.WriteString(`<h2 style="color:#e2b384;font-size:18px;margin-bottom:12px;">Daily Visitors</h2>
<div style="background:#1a1a1c;border-radius:8px;padding:16px;margin-bottom:32px;">
<table style="width:100%;border-collapse:collapse;height:150px;" cellpadding="0" cellspacing="0">
<tr style="vertical-align:bottom;">
`)
	maxDaily := 0
	for _, d := range report.DailyVisitors {
		if d.Count > maxDaily {
			maxDaily = d.Count
		}
	}
	for _, d := range report.DailyVisitors {
		barHeight := 10
		if maxDaily > 0 {
			barHeight = (d.Count * 130) / maxDaily
		}
		if barHeight < 10 {
			barHeight = 10
		}
		b.WriteString(fmt.Sprintf(`
  <td style="text-align:center;padding:0 4px;">
    <div style="color:#a0a0a0;font-size:11px;margin-bottom:4px;">%d</div>
    <div style="background:#e2b384;width:100%%;height:%dpx;border-radius:4px 4px 0 0;margin:0 auto;max-width:60px;"></div>
  </td>
`, d.Count, barHeight))
	}
	b.WriteString(`</tr><tr>`)
	for _, d := range report.DailyVisitors {
		// Show short date.
		label := d.Date
		if len(label) >= 10 {
			label = label[5:] // MM-DD
		}
		b.WriteString(fmt.Sprintf(`<td style="text-align:center;padding:4px;color:#a0a0a0;font-size:11px;">%s</td>`, label))
	}
	b.WriteString(`</tr></table></div>`)

	// Top visitors table.
	b.WriteString(`<h2 style="color:#e2b384;font-size:18px;margin-bottom:12px;">Top Visitors</h2>
<table style="width:100%%;border-collapse:collapse;background:#1a1a1c;border-radius:8px;margin-bottom:32px;">
<tr style="border-bottom:1px solid #2a2a2c;">
  <th style="padding:10px 12px;text-align:left;color:#a0a0a0;font-size:12px;font-weight:normal;">#</th>
  <th style="padding:10px 12px;text-align:left;color:#a0a0a0;font-size:12px;font-weight:normal;">IP Address</th>
  <th style="padding:10px 12px;text-align:left;color:#a0a0a0;font-size:12px;font-weight:normal;">City</th>
  <th style="padding:10px 12px;text-align:left;color:#a0a0a0;font-size:12px;font-weight:normal;">Pages</th>
  <th style="padding:10px 12px;text-align:right;color:#a0a0a0;font-size:12px;font-weight:normal;">Views</th>
</tr>
`)

	// Sort by total views descending.
	sorted := make([]VisitorStats, len(report.Stats))
	copy(sorted, report.Stats)
	sort.Slice(sorted, func(i, j int) bool {
		ti, tj := 0, 0
		for _, pv := range sorted[i].PageViews {
			ti += pv.TotalViews
		}
		for _, pv := range sorted[j].PageViews {
			tj += pv.TotalViews
		}
		return ti > tj
	})

	for i, v := range sorted {
		total := 0
		var pages []string
		for _, pv := range v.PageViews {
			total += pv.TotalViews
			pages = append(pages, pv.PagePath)
		}
		pageList := strings.Join(pages, ", ")
		b.WriteString(fmt.Sprintf(`
<tr style="border-bottom:1px solid #222224;">
  <td style="padding:8px 12px;color:#d8d8d8;font-size:13px;">%d</td>
  <td style="padding:8px 12px;color:#d8d8d8;font-size:13px;">%s</td>
  <td style="padding:8px 12px;color:#d8d8d8;font-size:13px;">%s</td>
  <td style="padding:8px 12px;color:#a0a0a0;font-size:12px;">%s</td>
  <td style="padding:8px 12px;color:#e2b384;font-size:13px;text-align:right;font-weight:bold;">%d</td>
</tr>
`, i+1, html.EscapeString(v.IP), html.EscapeString(v.City), html.EscapeString(pageList), total))
	}
	b.WriteString(`</table>`)

	b.WriteString(`
<p style="color:#666;font-size:12px;text-align:center;margin-top:32px;">
  This is an automated weekly analytics report. An Excel file with detailed data is attached.
</p>
</div>
</body>
</html>`)

	return b.String()
}

func generateExcel(report WeeklyReport) ([]byte, error) {
	f := excelize.NewFile()

	// Style for headers.
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#e2b384"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	// Sheet 1: Visitors
	sheet1 := "Visitors"
	f.SetSheetName("Sheet1", sheet1)
	headers1 := []string{"IP Address", "City", "First Visit", "Last Visit", "Total Page Views"}
	for i, h := range headers1 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet1, cell, h)
		f.SetCellStyle(sheet1, cell, cell, headerStyle)
	}
	for i, v := range report.Stats {
		row := i + 2
		total := 0
		for _, pv := range v.PageViews {
			total += pv.TotalViews
		}
		f.SetCellValue(sheet1, cellName(1, row), v.IP)
		f.SetCellValue(sheet1, cellName(2, row), v.City)
		f.SetCellValue(sheet1, cellName(3, row), v.FirstVisit)
		f.SetCellValue(sheet1, cellName(4, row), v.LastVisit)
		f.SetCellValue(sheet1, cellName(5, row), total)
	}
	f.SetColWidth(sheet1, "A", "A", 18)
	f.SetColWidth(sheet1, "B", "B", 15)
	f.SetColWidth(sheet1, "C", "D", 20)
	f.SetColWidth(sheet1, "E", "E", 16)

	// Sheet 2: Page Views
	sheet2 := "Page Views"
	f.NewSheet(sheet2)
	headers2 := []string{"IP Address", "Page Path", "View Count"}
	for i, h := range headers2 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet2, cell, h)
		f.SetCellStyle(sheet2, cell, cell, headerStyle)
	}
	pvRow := 2
	for _, v := range report.Stats {
		for _, pv := range v.PageViews {
			f.SetCellValue(sheet2, cellName(1, pvRow), v.IP)
			f.SetCellValue(sheet2, cellName(2, pvRow), pv.PagePath)
			f.SetCellValue(sheet2, cellName(3, pvRow), pv.TotalViews)
			pvRow++
		}
	}
	f.SetColWidth(sheet2, "A", "A", 18)
	f.SetColWidth(sheet2, "B", "B", 25)
	f.SetColWidth(sheet2, "C", "C", 12)

	// Sheet 3: Daily Summary
	sheet3 := "Daily Summary"
	f.NewSheet(sheet3)
	headers3 := []string{"Date", "Unique Visitors"}
	for i, h := range headers3 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet3, cell, h)
		f.SetCellStyle(sheet3, cell, cell, headerStyle)
	}
	for i, d := range report.DailyVisitors {
		row := i + 2
		f.SetCellValue(sheet3, cellName(1, row), d.Date)
		f.SetCellValue(sheet3, cellName(2, row), d.Count)
	}
	f.SetColWidth(sheet3, "A", "A", 14)
	f.SetColWidth(sheet3, "B", "B", 16)

	// Sheet 4: Page Summary
	sheet4 := "Page Summary"
	f.NewSheet(sheet4)
	headers4 := []string{"Page Path", "Total Views"}
	for i, h := range headers4 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet4, cell, h)
		f.SetCellStyle(sheet4, cell, cell, headerStyle)
	}
	for i, p := range report.PagePopularity {
		row := i + 2
		f.SetCellValue(sheet4, cellName(1, row), p.PagePath)
		f.SetCellValue(sheet4, cellName(2, row), p.TotalViews)
	}
	f.SetColWidth(sheet4, "A", "A", 25)
	f.SetColWidth(sheet4, "B", "B", 14)

	// Write to buffer.
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}
