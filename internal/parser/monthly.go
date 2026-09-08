package parser

import (
	"dschool-attendance-api/internal/models"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseMonthlyClassAttendance parses menu_102.php into models.MonthlyClassAttendance
func ParseMonthlyClassAttendance(doc *goquery.Document, defaultClass string) *models.MonthlyClassAttendance {
	result := &models.MonthlyClassAttendance{
		Classroom: defaultClass,
		Students:  make([]models.StudentMonthlyItem, 0),
	}

	// Header extraction
	doc.Find("div[align='center']").Each(func(i int, s *goquery.Selection) {
		text := CleanText(s.Text())
		if strings.Contains(text, "256") || strings.Contains(text, "257") {
			if result.YearMonth == "" {
				result.YearMonth = text
			}
		}
		if strings.HasPrefix(text, "ม.") && len(text) <= 10 {
			result.Classroom = text
		}
	})

	// Student table
	doc.Find("table").Each(func(i int, tbl *goquery.Selection) {
		firstRow := CleanText(tbl.Find("tr").First().Text())
		if strings.Contains(firstRow, "ชื่อ") && strings.Contains(firstRow, "มา") && strings.Contains(firstRow, "สาย") {
			tbl.Find("tr").Each(func(rIdx int, row *goquery.Selection) {
				if rIdx == 0 {
					return
				}
				cells := row.Find("td")
				// Columns: ชื่อ, มา, สาย, ไม่ลงเวลา, ลาป่วย, ลากิจ, ขาด
				if cells.Length() >= 7 {
					rawName := CleanText(cells.Eq(0).Text())
					if rawName == "" || rawName == "ชื่อ" {
						return
					}

					noVal := 0
					nameVal := rawName
					// Handle names formatted like "1. นายชื่อ นามสกุล"
					parts := strings.SplitN(rawName, ".", 2)
					if len(parts) == 2 {
						noVal = ParseInt(parts[0])
						nameVal = strings.TrimSpace(parts[1])
					}

					result.Students = append(result.Students, models.StudentMonthlyItem{
						No:            noVal,
						Name:          nameVal,
						Present:       ParseInt(cells.Eq(1).Text()),
						Late:          ParseInt(cells.Eq(2).Text()),
						Unrecorded:    ParseInt(cells.Eq(3).Text()),
						SickLeave:     ParseInt(cells.Eq(4).Text()),
						PersonalLeave: ParseInt(cells.Eq(5).Text()),
						Absent:        ParseInt(cells.Eq(6).Text()),
					})
				}
			})
		}
	})

	return result
}
