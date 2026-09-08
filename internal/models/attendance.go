package models

// ApiResponse represents a standardized JSON response
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// DailyStudentItem represents individual student attendance on a specific day
type DailyStudentItem struct {
	No           int    `json:"no"`
	Name         string `json:"name"`
	Gate         string `json:"gate"`
	Status       string `json:"status"` // มา, สาย, ไม่ลงเวลา, ลาป่วย, ลากิจ, ขาด
	CheckInTime  string `json:"check_in_time"`
	CheckOutTime string `json:"check_out_time"`
}

// DailyClassSummaryStat represents summary counts for a classroom
type DailyClassSummaryStat struct {
	Category string `json:"category"`
	Male     int    `json:"male"`
	Female   int    `json:"female"`
	Total    int    `json:"total"`
}

// DailyClassAttendance represents classroom daily attendance result
type DailyClassAttendance struct {
	Date      string                  `json:"date"`
	Classroom string                  `json:"classroom"`
	Summary   []DailyClassSummaryStat `json:"summary"`
	Students  []DailyStudentItem      `json:"students"`
}

// AttendanceStatCounts represents attendance metrics
type AttendanceStatCounts struct {
	Total         int `json:"total"`
	Present       int `json:"present"`
	OffCampus     int `json:"off_campus"`
	Unrecorded    int `json:"unrecorded"`
	SickLeave     int `json:"sick_leave"`
	PersonalLeave int `json:"personal_leave"`
	Absent        int `json:"absent"`
}

// GradeLevelAttendance represents stats for a whole grade (ม.1, ม.2, ...)
type GradeLevelAttendance struct {
	Grade string `json:"grade"`
	AttendanceStatCounts
}

// ClassroomAttendance represents stats for a single classroom in overview
type ClassroomAttendance struct {
	Classroom string `json:"classroom"`
	AttendanceStatCounts
}

// SchoolOverviewAttendance represents school-wide daily overview
type SchoolOverviewAttendance struct {
	Date        string                           `json:"date"`
	Summary     AttendanceStatCounts             `json:"summary"`
	GradeLevels []GradeLevelAttendance           `json:"grade_levels"`
	Classrooms  map[string][]ClassroomAttendance `json:"classrooms"`
}

// StudentMonthlyItem represents student monthly attendance totals
type StudentMonthlyItem struct {
	No            int    `json:"no"`
	Name          string `json:"name"`
	Present       int    `json:"present"`
	Late          int    `json:"late"`
	Unrecorded    int    `json:"unrecorded"`
	SickLeave     int    `json:"sick_leave"`
	PersonalLeave int    `json:"personal_leave"`
	Absent        int    `json:"absent"`
}

// MonthlyClassAttendance represents monthly attendance report for a classroom
type MonthlyClassAttendance struct {
	YearMonth string               `json:"year_month"`
	Classroom string               `json:"classroom"`
	Students  []StudentMonthlyItem `json:"students"`
}

// StudentSemesterItem represents student semester attendance totals
type StudentSemesterItem struct {
	No            int    `json:"no"`
	Name          string `json:"name"`
	Present       int    `json:"present"`
	Late          int    `json:"late"`
	Unrecorded    int    `json:"unrecorded"`
	SickLeave     int    `json:"sick_leave"`
	PersonalLeave int    `json:"personal_leave"`
	Absent        int    `json:"absent"`
}

// SemesterClassAttendance represents semester attendance report for a classroom
type SemesterClassAttendance struct {
	Term      int                   `json:"term"`
	Classroom string                `json:"classroom"`
	Students  []StudentSemesterItem `json:"students"`
}

// StudentSearchResult represents a student found via search
type StudentSearchResult struct {
	No       int    `json:"no"`
	SdNo     string `json:"sd_no"`
	Name     string `json:"name"`
	PhotoURL string `json:"photo_url"`
}

// StudentAttendanceDayLog represents an individual student's status on a date
type StudentAttendanceDayLog struct {
	Date   string `json:"date"`
	Status string `json:"status"`
}

// StudentAttendanceStats represents individual student aggregate metrics
type StudentAttendanceStats struct {
	Present       int `json:"present"`
	Late          int `json:"late"`
	Unrecorded    int `json:"unrecorded"`
	SickLeave     int `json:"sick_leave"`
	PersonalLeave int `json:"personal_leave"`
	Absent        int `json:"absent"`
}

// StudentAttendanceProfile represents student 360 attendance profile
type StudentAttendanceProfile struct {
	SdNo    string                    `json:"sd_no"`
	Term    int                       `json:"term"`
	Summary StudentAttendanceStats    `json:"summary"`
	History []StudentAttendanceDayLog `json:"history"`
}

// FlagCeremonyCount represents total flag ceremony stats
type FlagCeremonyCount struct {
	Total      int `json:"total"`
	Attended   int `json:"attended"`
	Unattended int `json:"unattended"`
	Absent     int `json:"absent"`
}

// FlagCeremonyLevel represents flag ceremony breakdown per level/class
type FlagCeremonyLevel struct {
	Level        string `json:"level"`
	Total        int    `json:"total"`
	TotalMale    int    `json:"total_male"`
	TotalFemale  int    `json:"total_female"`
	AttendTotal  int    `json:"attend_total"`
	AttendMale   int    `json:"attend_male"`
	AttendFemale int    `json:"attend_female"`
}

// FlagCeremonyOverview represents flag ceremony report
type FlagCeremonyOverview struct {
	Date        string              `json:"date"`
	Summary     FlagCeremonyCount   `json:"summary"`
	GradeLevels []FlagCeremonyLevel `json:"grade_levels"`
	Classrooms  []FlagCeremonyLevel `json:"classrooms"`
}
