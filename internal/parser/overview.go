package parser

import (
	"dschool-attendance-api/internal/models"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseOverviewAttendance parses menu_101.php into models.SchoolOverviewAttendance
func ParseOverviewAttendance(doc *goquery.Document) *models.SchoolOverviewAttendance {
	result := &models.SchoolOverviewAttendance{
		GradeLevels: make([]models.GradeLevelAttendance, 0),
		Classrooms:  make(map[string][]models.ClassroomAttendance),
	}

	// Extract Date
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

	// 1. Overall Summary Table
	doc.Find("table").Each(func(i int, tbl *goquery.Selection) {
		text := tbl.Text()
		if strings.Contains(text, "ทั้งหมด") && strings.Contains(text, "ลงเวลา") && strings.Contains(text, "ลาป่วย") && !strings.Contains(text, "ระดับชั้น") {
			tbl.Find("tr").Each(func(_ int, row *goquery.Selection) {
				cells := row.Find("td")
				if cells.Length() >= 2 {
					label := CleanText(cells.Eq(0).Text())
					val := ParseInt(cells.Eq(1).Text())
					switch {
					case strings.Contains(label, "ทั้งหมด"):
						result.Summary.Total = val
					case strings.Contains(label, "เรียนนอกสถานที่"):
						result.Summary.OffCampus = val
					case strings.Contains(label, "ไม่ลงเวลา"):
						result.Summary.Unrecorded = val
					case strings.Contains(label, "ลงเวลา"): // Must come after "ไม่ลงเวลา"
						result.Summary.Present = val
					case strings.Contains(label, "ลาป่วย"):
						result.Summary.SickLeave = val
					case strings.Contains(label, "ลากิจ"):
						result.Summary.PersonalLeave = val
					case strings.Contains(label, "ขาด"):
						result.Summary.Absent = val
					}
				}
			})
		}
	})

	// 2. Grade-level & Classroom tables
	doc.Find("table").Each(func(i int, tbl *goquery.Selection) {
		header := CleanText(tbl.Find("tr").First().Text())
		if strings.Contains(header, "ระดับชั้น") && strings.Contains(header, "ทั้งหมด") && strings.Contains(header, "ลงเวลา") {
			parentSection := tbl.Closest("div.main_box")
			heading := CleanText(parentSection.Find("h3").Text())

			if strings.Contains(heading, "รายระดับชั้น") {
				// High-level grade rows (ม.1 to ม.6)
				tbl.Find("tr").Each(func(rIdx int, row *goquery.Selection) {
					cells := row.Find("td")
					if cells.Length() >= 8 {
						gradeName := CleanText(cells.Eq(0).Text())
						if gradeName != "" && gradeName != "ระดับชั้น" && gradeName != "รวม" {
							result.GradeLevels = append(result.GradeLevels, models.GradeLevelAttendance{
								Grade: gradeName,
								AttendanceStatCounts: models.AttendanceStatCounts{
									Total:         ParseInt(cells.Eq(1).Text()),
									Present:       ParseInt(cells.Eq(2).Text()),
									OffCampus:     ParseInt(cells.Eq(3).Text()),
									Unrecorded:    ParseInt(cells.Eq(4).Text()),
									SickLeave:     ParseInt(cells.Eq(5).Text()),
									PersonalLeave: ParseInt(cells.Eq(6).Text()),
									Absent:        ParseInt(cells.Eq(7).Text()),
								},
							})
						}
					}
				})
			} else if strings.HasPrefix(heading, "ม.") {
				// Classroom breakdown for this grade (e.g. ม.1 -> ม.101, ม.102...)
				gradeKey := heading
				classrooms := make([]models.ClassroomAttendance, 0)

				tbl.Find("tr").Each(func(rIdx int, row *goquery.Selection) {
					cells := row.Find("td")
					if cells.Length() >= 8 {
						className := CleanText(cells.Eq(0).Text())
						if className != "" && className != "ระดับชั้น" && className != "รวม" {
							classrooms = append(classrooms, models.ClassroomAttendance{
								Classroom: className,
								AttendanceStatCounts: models.AttendanceStatCounts{
									Total:         ParseInt(cells.Eq(1).Text()),
									Present:       ParseInt(cells.Eq(2).Text()),
									OffCampus:     ParseInt(cells.Eq(3).Text()),
									Unrecorded:    ParseInt(cells.Eq(4).Text()),
									SickLeave:     ParseInt(cells.Eq(5).Text()),
									PersonalLeave: ParseInt(cells.Eq(6).Text()),
									Absent:        ParseInt(cells.Eq(7).Text()),
								},
							})
						}
					}
				})
				if len(classrooms) > 0 {
					result.Classrooms[gradeKey] = classrooms
				}
			}
		}
	})

	return result
}
