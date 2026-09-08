package parser

import (
	"dschool-attendance-api/internal/models"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseStudentSearchResults parses student_360_search_student.php
func ParseStudentSearchResults(doc *goquery.Document, baseURL string) []models.StudentSearchResult {
	results := make([]models.StudentSearchResult, 0)

	doc.Find("tr[onclick]").Each(func(i int, row *goquery.Selection) {
		onclick, _ := row.Attr("onclick")
		// e.g. location.href='student360.php?menu=student360&sd_no=18948';
		if strings.Contains(onclick, "sd_no=") {
			var sdNo string
			idx := strings.Index(onclick, "sd_no=")
			if idx != -1 {
				sub := onclick[idx+6:]
				for _, ch := range sub {
					if ch >= '0' && ch <= '9' {
						sdNo += string(ch)
					} else {
						break
					}
				}
			}

			cells := row.Find("td")
			if cells.Length() >= 3 {
				noVal := ParseInt(cells.Eq(0).Text())
				nameVal := CleanText(cells.Eq(2).Text())

				photoURL, _ := cells.Eq(1).Find("img").Attr("src")
				if photoURL != "" && !strings.HasPrefix(photoURL, "http") {
					photoURL = baseURL + "/" + strings.TrimPrefix(photoURL, "/")
				}

				if sdNo != "" && nameVal != "" {
					results = append(results, models.StudentSearchResult{
						No:       noVal,
						SdNo:     sdNo,
						Name:     nameVal,
						PhotoURL: photoURL,
					})
				}
			}
		}
	})

	return results
}

// ParseStudentAttendanceProfile parses menu_113.php
func ParseStudentAttendanceProfile(doc *goquery.Document, sdNo string, term int) *models.StudentAttendanceProfile {
	profile := &models.StudentAttendanceProfile{
		SdNo:    sdNo,
		Term:    term,
		History: make([]models.StudentAttendanceDayLog, 0),
	}

	tables := doc.Find("table")
	tables.Each(func(tblIdx int, tbl *goquery.Selection) {
		text := tbl.Text()

		// 1. Metric summary table (contains "มา", "สาย", "วัน")
		if strings.Contains(text, "วัน") && strings.Contains(text, "มา") && strings.Contains(text, "สาย") && !strings.Contains(text, "วันที่") {
			tbl.Find("tr").Each(func(_ int, row *goquery.Selection) {
				cells := row.Find("td")
				if cells.Length() >= 2 {
					label := CleanText(cells.Eq(0).Text())
					count := ParseInt(cells.Eq(1).Text())
					switch label {
					case "มา":
						profile.Summary.Present = count
					case "สาย":
						profile.Summary.Late = count
					case "ไม่ลงเวลา":
						profile.Summary.Unrecorded = count
					case "ลาป่วย":
						profile.Summary.SickLeave = count
					case "ลากิจ":
						profile.Summary.PersonalLeave = count
					case "ขาด":
						profile.Summary.Absent = count
					}
				}
			})
		}

		// 2. Day-by-day log table (contains "วันที่" and "สถานะ")
		if strings.Contains(text, "วันที่") && strings.Contains(text, "สถานะ") {
			tbl.Find("tr").Each(func(rIdx int, row *goquery.Selection) {
				cells := row.Find("td")
				if cells.Length() == 2 {
					dateStr := CleanText(cells.Eq(0).Text())
					statusStr := CleanText(cells.Eq(1).Text())
					if dateStr != "" && dateStr != "วันที่" && statusStr != "" && statusStr != "สถานะ" {
						profile.History = append(profile.History, models.StudentAttendanceDayLog{
							Date:   dateStr,
							Status: statusStr,
						})
					}
				}
			})
		}
	})

	return profile
}
