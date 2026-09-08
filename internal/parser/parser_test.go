package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseDailyClassAttendance(t *testing.T) {
	htmlContent := `
	<div align="center">8 ก.ย. 2569</div>
	<div align="center">ม.401</div>
	<table>
		<tr><td></td><td>ชาย</td><td>หญิง</td><td>รวม</td></tr>
		<tr><td>จำนวนเต็ม</td><td>16</td><td>14</td><td>30</td></tr>
		<tr><td>มา</td><td>14</td><td>13</td><td>27</td></tr>
		<tr><td>ไม่ลงเวลา</td><td>2</td><td>1</td><td>3</td></tr>
	</table>
	<table>
		<tr><th>#</th><th>รูป</th><th>ชื่อ</th><th>ประตู</th><th>สถานะ</th><th>เวลามา</th><th>เวลากลับ</th></tr>
		<tr><td>1</td><td></td><td>กิตติพิชญ์ ปัชชามูล</td><td>Face AI</td><td>มา</td><td>07:05</td><td>-</td></tr>
		<tr><td>2</td><td></td><td>ธนกร บุญทะจิตต์</td><td>Face AI</td><td>สาย</td><td>08:01</td><td>-</td></tr>
	</table>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := ParseDailyClassAttendance(doc, "401")
	if result.Classroom != "ม.401" {
		t.Errorf("Expected classroom 'ม.401', got '%s'", result.Classroom)
	}
	if len(result.Summary) != 3 {
		t.Errorf("Expected 3 summary items, got %d", len(result.Summary))
	}
	if len(result.Students) != 2 {
		t.Fatalf("Expected 2 students, got %d", len(result.Students))
	}
	if result.Students[0].Name != "กิตติพิชญ์ ปัชชามูล" || result.Students[0].Status != "มา" || result.Students[0].CheckInTime != "07:05" {
		t.Errorf("Student 0 data mismatch: %+v", result.Students[0])
	}
	if result.Students[1].Status != "สาย" || result.Students[1].CheckInTime != "08:01" {
		t.Errorf("Student 1 data mismatch: %+v", result.Students[1])
	}
}

func TestParseStudentSearchResults(t *testing.T) {
	htmlContent := `
	<table>
		<tr onclick="location.href='student360.php?menu=student360&sd_no=18948';">
			<td>1</td>
			<td><img src="card_template/img_proxy.php?sd_no=18948"></td>
			<td>กิตติพิชญ์ ปัชชามูล</td>
		</tr>
	</table>
	`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	results := ParseStudentSearchResults(doc, "https://dschool-l1.gp-education.com/dschool_app_v2020")
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].SdNo != "18948" || results[0].Name != "กิตติพิชญ์ ปัชชามูล" {
		t.Errorf("Result mismatch: %+v", results[0])
	}
}

func TestParseStudentAttendanceProfile(t *testing.T) {
	htmlContent := `
	<table>
		<tr><td>มา</td><td>57</td><td>วัน</td></tr>
		<tr><td>สาย</td><td>6</td><td>วัน</td></tr>
		<tr><td>ไม่ลงเวลา</td><td>52</td><td>วัน</td></tr>
		<tr><td>ลาป่วย</td><td>2</td><td>วัน</td></tr>
		<tr><td>ลากิจ</td><td>-</td><td>วัน</td></tr>
		<tr><td>ขาด</td><td>-</td><td>วัน</td></tr>
	</table>
	<table>
		<tr><td>วันที่</td><td>สถานะ</td></tr>
		<tr><td>08 ก.ย. 69</td><td>มา</td></tr>
		<tr><td>07 ก.ย. 69</td><td>สาย</td></tr>
	</table>
	`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	profile := ParseStudentAttendanceProfile(doc, "18948", 1)
	if profile.Summary.Present != 57 || profile.Summary.Late != 6 || profile.Summary.SickLeave != 2 {
		t.Errorf("Summary mismatch: %+v", profile.Summary)
	}
	if len(profile.History) != 2 {
		t.Fatalf("Expected 2 history items, got %d", len(profile.History))
	}
	if profile.History[0].Date != "08 ก.ย. 69" || profile.History[0].Status != "มา" {
		t.Errorf("History 0 mismatch: %+v", profile.History[0])
	}
}
