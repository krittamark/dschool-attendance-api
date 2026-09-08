package parser

import (
	"dschool-attendance-api/internal/models"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseFlagCeremonyOverview parses menu_201.php into models.FlagCeremonyOverview
func ParseFlagCeremonyOverview(doc *goquery.Document) *models.FlagCeremonyOverview {
	result := &models.FlagCeremonyOverview{
		GradeLevels: make([]models.FlagCeremonyLevel, 0),
		Classrooms:  make([]models.FlagCeremonyLevel, 0),
	}

	doc.Find("div[align='center']").Each(func(i int, s *goquery.Selection) {
		text := CleanText(s.Text())
		if strings.Contains(text, "ก.ย.") || strings.Contains(text, "256") || strings.Contains(text, "257") {
			if result.Date == "" {
				parts := strings.Fields(text)
				if len(parts) >= 3 {
					result.Date = strings.Join(parts[:3], " ")
				}
			}
		}
	})

	// Summary stats table
	doc.Find("table").Each(func(i int, tbl *goquery.Selection) {
		text := tbl.Text()
		if strings.Contains(text, "เข้าร่วม") && strings.Contains(text, "ไม่เข้าร่วม") && !strings.Contains(text, "ระดับชั้น") {
			tbl.Find("tr").Each(func(_ int, row *goquery.Selection) {
				cells := row.Find("td")
				if cells.Length() >= 2 {
					label := CleanText(cells.Eq(0).Text())
					val := ParseInt(cells.Eq(1).Text())
					switch {
					case strings.Contains(label, "ทั้งหมด"):
						result.Summary.Total = val
					case strings.Contains(label, "เข้าร่วม") && !strings.Contains(label, "ไม่"):
						result.Summary.Attended = val
					case strings.Contains(label, "ไม่เข้าร่วม"):
						result.Summary.Unattended = val
					case strings.Contains(label, "ขาด"):
						result.Summary.Absent = val
					}
				}
			})
		}
	})

	// Grade and classroom breakdown tables
	doc.Find("table").Each(func(i int, tbl *goquery.Selection) {
		header := CleanText(tbl.Find("tr").First().Text())
		if strings.Contains(header, "ระดับชั้น") && strings.Contains(header, "เข้าร่วม") {
			// Find if this is grade-level or classroom-level
			parentSection := tbl.Closest("div.main_box")
			heading := CleanText(parentSection.Find("h3").Text())

			tbl.Find("tr").Each(func(rIdx int, row *goquery.Selection) {
				cells := row.Find("td")
				// Columns: ระดับชั้น, ทั้งหมด(รวม, ชาย, หญิง), เข้าร่วม(รวม, ชาย, หญิง)
				if cells.Length() >= 7 {
					name := CleanText(cells.Eq(0).Text())
					if name == "" || name == "ระดับชั้น" || name == "รวม" {
						return
					}

					item := models.FlagCeremonyLevel{
						Level:        name,
						Total:        ParseInt(cells.Eq(1).Text()),
						TotalMale:    ParseInt(cells.Eq(2).Text()),
						TotalFemale:  ParseInt(cells.Eq(3).Text()),
						AttendTotal:  ParseInt(cells.Eq(4).Text()),
						AttendMale:   ParseInt(cells.Eq(5).Text()),
						AttendFemale: ParseInt(cells.Eq(6).Text()),
					}

					if strings.Contains(heading, "รายระดับชั้น") || len(name) <= 4 {
						result.GradeLevels = append(result.GradeLevels, item)
					} else {
						result.Classrooms = append(result.Classrooms, item)
					}
				}
			})
		}
	})

	return result
}
