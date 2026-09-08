package parser

import (
	"dschool-attendance-api/internal/models"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseSemesterClassAttendance parses menu_103.php into models.SemesterClassAttendance
func ParseSemesterClassAttendance(doc *goquery.Document, defaultClass string, term int) *models.SemesterClassAttendance {
	result := &models.SemesterClassAttendance{
		Term:      term,
		Classroom: defaultClass,
		Students:  make([]models.StudentSemesterItem, 0),
	}

	doc.Find("div[align='center']").Each(func(i int, s *goquery.Selection) {
		text := CleanText(s.Text())
		if strings.Contains(text, "ภาคเรียนที่ 2") {
			result.Term = 2
		} else if strings.Contains(text, "ภาคเรียนที่ 1") {
			result.Term = 1
		}
		if strings.HasPrefix(text, "ม.") && len(text) <= 10 {
			result.Classroom = text
		}
	})

	doc.Find("table").Each(func(i int, tbl *goquery.Selection) {
		firstRow := CleanText(tbl.Find("tr").First().Text())
		if strings.Contains(firstRow, "ชื่อ") && strings.Contains(firstRow, "มา") && strings.Contains(firstRow, "สาย") {
			tbl.Find("tr").Each(func(rIdx int, row *goquery.Selection) {
				if rIdx == 0 {
					return
				}
				cells := row.Find("td")
				if cells.Length() >= 7 {
					rawName := CleanText(cells.Eq(0).Text())
					if rawName == "" || rawName == "ชื่อ" {
						return
					}

					noVal := 0
					nameVal := rawName
					parts := strings.SplitN(rawName, ".", 2)
					if len(parts) == 2 {
						noVal = ParseInt(parts[0])
						nameVal = strings.TrimSpace(parts[1])
					}

					result.Students = append(result.Students, models.StudentSemesterItem{
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
