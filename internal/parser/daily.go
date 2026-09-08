package parser

import (
	"dschool-attendance-api/internal/models"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseDailyClassAttendance parses menu_105.php into models.DailyClassAttendance
func ParseDailyClassAttendance(doc *goquery.Document, defaultClass string) *models.DailyClassAttendance {
	result := &models.DailyClassAttendance{
		Classroom: defaultClass,
		Summary:   make([]models.DailyClassSummaryStat, 0),
		Students:  make([]models.DailyStudentItem, 0),
	}

	// Extract Date and Classroom from header divs
	doc.Find("div[align='center']").Each(func(i int, s *goquery.Selection) {
		text := CleanText(s.Text())
		if strings.Contains(text, "ก.ย.") || strings.Contains(text, "ม.ค.") || strings.Contains(text, "256") || strings.Contains(text, "257") {
			if result.Date == "" {
				// Clean up extra calendar icon text if present
				dateParts := strings.Fields(text)
				if len(dateParts) >= 3 {
					result.Date = strings.Join(dateParts[:3], " ")
				} else {
					result.Date = text
				}
			}
		}
		if strings.HasPrefix(text, "ม.") && len(text) <= 10 {
			result.Classroom = text
		}
	})

	// Find tables
	tables := doc.Find("table")
	tables.Each(func(i int, tbl *goquery.Selection) {
		headerTexts := make([]string, 0)
		tbl.Find("tr").First().Find("td, th").Each(func(_ int, cell *goquery.Selection) {
			headerTexts = append(headerTexts, CleanText(cell.Text()))
		})
		headerLine := strings.Join(headerTexts, " ")

		// 1. Summary Table: contains "ชาย" and "หญิง"
		if strings.Contains(headerLine, "ชาย") && strings.Contains(headerLine, "หญิง") {
			tbl.Find("tr").Each(func(rowIdx int, row *goquery.Selection) {
				cells := row.Find("td, th")
				if cells.Length() >= 4 {
					cat := CleanText(cells.Eq(0).Text())
					if cat != "" && cat != "ชาย" && cat != "หญิง" {
						stat := models.DailyClassSummaryStat{
							Category: cat,
							Male:     ParseInt(cells.Eq(1).Text()),
							Female:   ParseInt(cells.Eq(2).Text()),
							Total:    ParseInt(cells.Eq(3).Text()),
						}
						result.Summary = append(result.Summary, stat)
					}
				}
			})
		}

		// 2. Student List Table: contains "ชื่อ" and ("ประตู" or "สถานะ")
		if strings.Contains(headerLine, "ชื่อ") && (strings.Contains(headerLine, "สถานะ") || strings.Contains(headerLine, "ประตู")) {
			tbl.Find("tr").Each(func(rowIdx int, row *goquery.Selection) {
				// Skip header
				if rowIdx == 0 {
					return
				}
				cells := row.Find("td")
				// Columns: # (0), รูป (1), ชื่อ (2), ประตู (3), สถานะ (4), เวลามา (5), เวลากลับ (6)
				if cells.Length() >= 5 {
					noStr := CleanText(cells.Eq(0).Text())
					noVal := ParseInt(noStr)
					if noVal == 0 {
						return // skip non-student rows
					}

					var name, gate, status, inTime, outTime string
					if cells.Length() >= 7 {
						name = CleanText(cells.Eq(2).Text())
						gate = CleanText(cells.Eq(3).Text())
						status = CleanText(cells.Eq(4).Text())
						inTime = CleanText(cells.Eq(5).Text())
						outTime = CleanText(cells.Eq(6).Text())
					} else if cells.Length() >= 5 {
						name = CleanText(cells.Eq(1).Text())
						gate = CleanText(cells.Eq(2).Text())
						status = CleanText(cells.Eq(3).Text())
						inTime = CleanText(cells.Eq(4).Text())
					}

					if name != "" {
						result.Students = append(result.Students, models.DailyStudentItem{
							No:           noVal,
							Name:         name,
							Gate:         gate,
							Status:       status,
							CheckInTime:  inTime,
							CheckOutTime: outTime,
						})
					}
				}
			})
		}
	})

	return result
}
