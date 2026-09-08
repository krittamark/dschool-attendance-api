package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CleanText trims spaces and NBSP
func CleanText(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	return strings.TrimSpace(s)
}

// ParseInt parses an integer from text, treating "-", "", or commas gracefully
func ParseInt(s string) int {
	s = CleanText(s)
	s = strings.ReplaceAll(s, ",", "")
	if s == "" || s == "-" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

// ClassToEdLevel maps a classroom string like "101", "ม.101", "401" to edlevel and clean class string
func ClassToEdLevel(classroom string) (edlevel string, classCode string) {
	c := strings.TrimSpace(classroom)
	c = strings.TrimPrefix(c, "ม.")
	c = strings.TrimPrefix(c, "ม. ")
	c = strings.TrimSpace(c)

	if len(c) == 0 {
		return "2", "101"
	}

	firstChar := c[0]
	if firstChar == '1' || firstChar == '2' || firstChar == '3' {
		return "2", c
	}
	return "3", c
}

// FormatDateParam converts various date formats (YYYY-MM-DD or DD/MM/YYYY or YY/MM/DD)
// into the dschool required format: "YY/MM/DD" (Buddhist Era 2-digit year)
func FormatDateParam(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		loc = time.Local
	}
	now := time.Now().In(loc)

	if dateStr == "" || dateStr == "today" {
		beYear2 := (now.Year() + 543) % 100
		return fmt.Sprintf("%02d/%02d/%02d", beYear2, int(now.Month()), now.Day())
	}

	// Normalize separators to slash
	norm := strings.ReplaceAll(dateStr, "-", "/")
	parts := strings.Split(norm, "/")

	if len(parts) == 3 {
		p0, _ := strconv.Atoi(parts[0])
		p1, _ := strconv.Atoi(parts[1])
		p2, _ := strconv.Atoi(parts[2])

		// Case A: First part is 4-digit year (YYYY/MM/DD)
		if len(parts[0]) == 4 || p0 > 1900 {
			year := p0
			month := p1
			day := p2
			var beYear2 int
			if year > 2400 {
				// Already Buddhist Era (e.g. 2569)
				beYear2 = year % 100
			} else {
				// Common Era (e.g. 2026) -> add 543
				beYear2 = (year + 543) % 100
			}
			return fmt.Sprintf("%02d/%02d/%02d", beYear2, month, day)
		}

		// Case B: Last part is 4-digit year (DD/MM/YYYY)
		if len(parts[2]) == 4 || p2 > 1900 {
			day := p0
			month := p1
			year := p2
			var beYear2 int
			if year > 2400 {
				beYear2 = year % 100
			} else {
				beYear2 = (year + 543) % 100
			}
			return fmt.Sprintf("%02d/%02d/%02d", beYear2, month, day)
		}

		// Case C: 2-digit year YY/MM/DD (already Buddhist era 2-digit)
		return fmt.Sprintf("%02d/%02d/%02d", p0%100, p1, p2)
	}

	return dateStr
}
