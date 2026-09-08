package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"dschool-attendance-api/internal/auth"
	"dschool-attendance-api/internal/cache"
	"dschool-attendance-api/internal/client"
	"dschool-attendance-api/internal/config"
	"dschool-attendance-api/internal/models"
	"dschool-attendance-api/internal/parser"
)

var (
	classroomRegex = regexp.MustCompile(`^(ม\.?\s*)?([1-6]\d{2})$`)
	numericRegex   = regexp.MustCompile(`^\d+$`)
)

type Handler struct {
	client   *client.Client
	cfg      *config.Config
	keyStore *auth.KeyStore
	cache    *cache.MemoryCache
}

func NewHandler(c *client.Client, cfg *config.Config, ks *auth.KeyStore, mc *cache.MemoryCache) *Handler {
	return &Handler{
		client:   c,
		cfg:      cfg,
		keyStore: ks,
		cache:    mc,
	}
}

func writeJSON(w http.ResponseWriter, status int, resp models.ApiResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func writeSuccess(w http.ResponseWriter, data interface{}, message string) {
	writeJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func writeError(w http.ResponseWriter, status int, errMsg string) {
	writeJSON(w, status, models.ApiResponse{
		Success: false,
		Error:   errMsg,
	})
}

// validateDate checks if date string is a valid calendar date in YYYY-MM-DD, DD/MM/YYYY, or YY/MM/DD
func validateDate(dateStr string) error {
	if dateStr == "" || dateStr == "today" {
		return nil
	}

	norm := strings.ReplaceAll(dateStr, "-", "/")
	parts := strings.Split(norm, "/")
	if len(parts) != 3 {
		return fmt.Errorf("invalid date format %q; expected YYYY-MM-DD (e.g. 2026-09-08)", dateStr)
	}

	p0, err0 := strconv.Atoi(parts[0])
	p1, err1 := strconv.Atoi(parts[1])
	p2, err2 := strconv.Atoi(parts[2])
	if err0 != nil || err1 != nil || err2 != nil {
		return fmt.Errorf("invalid date format %q; numeric values expected", dateStr)
	}

	var y, m, d int
	if len(parts[0]) == 4 || p0 > 1900 {
		y, m, d = p0, p1, p2
	} else if len(parts[2]) == 4 || p2 > 1900 {
		d, m, y = p0, p1, p2
	} else {
		y, m, d = p0+2500, p1, p2
	}

	if y > 2400 {
		y -= 543 // convert BE to CE for standard calendar validation
	}

	if m < 1 || m > 12 || d < 1 || d > 31 {
		return fmt.Errorf("date out of range: month %d, day %d", m, d)
	}

	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	if t.Year() != y || int(t.Month()) != m || t.Day() != d {
		return fmt.Errorf("impossible calendar date %q (e.g. day exceeds days in month)", dateStr)
	}

	return nil
}

// HealthCheck handles GET /api/health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "dschool-attendance-api",
	}, "Service is operational")
}

// DailyAttendance handles GET /api/attendance/daily?classroom=101&date=YYYY-MM-DD
func (h *Handler) DailyAttendance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	classroom := strings.TrimSpace(r.URL.Query().Get("classroom"))
	if classroom == "" {
		classroom = "101"
	} else if !classroomRegex.MatchString(classroom) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid classroom %q; expected format like '101', '401', 'ม.101'", classroom))
		return
	}

	dateStr := strings.TrimSpace(r.URL.Query().Get("date"))
	if err := validateDate(dateStr); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cacheKey := fmt.Sprintf("daily:%s:%s", classroom, dateStr)
	if h.cache != nil {
		if cached, ok := h.cache.Get(cacheKey); ok {
			w.Header().Set("X-Cache", "HIT")
			writeSuccess(w, cached, "Daily attendance retrieved successfully (cached)")
			return
		}
	}

	doc, err := h.client.GetDailyClassAttendance(ctx, classroom, dateStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch daily attendance: "+err.Error())
		return
	}

	result := parser.ParseDailyClassAttendance(doc, classroom)
	if h.cache != nil {
		h.cache.Set(cacheKey, result, 60*time.Second) // 1 minute TTL
	}
	w.Header().Set("X-Cache", "MISS")
	writeSuccess(w, result, "Daily attendance retrieved successfully")
}

// OverviewAttendance handles GET /api/attendance/overview?date=YYYY-MM-DD
func (h *Handler) OverviewAttendance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dateStr := strings.TrimSpace(r.URL.Query().Get("date"))
	if err := validateDate(dateStr); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cacheKey := fmt.Sprintf("overview:%s", dateStr)
	if h.cache != nil {
		if cached, ok := h.cache.Get(cacheKey); ok {
			w.Header().Set("X-Cache", "HIT")
			writeSuccess(w, cached, "School overview attendance retrieved successfully (cached)")
			return
		}
	}

	doc, err := h.client.GetOverviewAttendance(ctx, dateStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch overview attendance: "+err.Error())
		return
	}

	result := parser.ParseOverviewAttendance(doc)
	if h.cache != nil {
		h.cache.Set(cacheKey, result, 60*time.Second) // 1 minute TTL
	}
	w.Header().Set("X-Cache", "MISS")
	writeSuccess(w, result, "School overview attendance retrieved successfully")
}

// MonthlyAttendance handles GET /api/attendance/monthly?classroom=101&month=9&year=2569
func (h *Handler) MonthlyAttendance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	classroom := strings.TrimSpace(r.URL.Query().Get("classroom"))
	if classroom == "" {
		classroom = "101"
	} else if !classroomRegex.MatchString(classroom) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid classroom %q; expected format like '101', '401'", classroom))
		return
	}

	month := strings.TrimSpace(r.URL.Query().Get("month"))
	year := strings.TrimSpace(r.URL.Query().Get("year"))

	if month != "" {
		mVal, err := strconv.Atoi(month)
		if err != nil || mVal < 1 || mVal > 12 {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid month %q; month must be between 1 and 12", month))
			return
		}
		if mVal < 10 && len(month) == 1 {
			month = "0" + month
		}
	}

	if year != "" {
		yVal, err := strconv.Atoi(year)
		if err != nil || yVal < 2500 || yVal > 2600 {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid year %q; expected Buddhist era year like 2569", year))
			return
		}
	}

	cacheKey := fmt.Sprintf("monthly:%s:%s:%s", classroom, month, year)
	if h.cache != nil {
		if cached, ok := h.cache.Get(cacheKey); ok {
			w.Header().Set("X-Cache", "HIT")
			writeSuccess(w, cached, "Monthly classroom attendance retrieved successfully (cached)")
			return
		}
	}

	doc, err := h.client.GetMonthlyClassAttendance(ctx, classroom, month, year)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch monthly attendance: "+err.Error())
		return
	}

	result := parser.ParseMonthlyClassAttendance(doc, classroom)
	if h.cache != nil {
		h.cache.Set(cacheKey, result, 5*time.Minute) // 5 minutes TTL
	}
	w.Header().Set("X-Cache", "MISS")
	writeSuccess(w, result, "Monthly classroom attendance retrieved successfully")
}

// SemesterAttendance handles GET /api/attendance/semester?classroom=101&term=1
func (h *Handler) SemesterAttendance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	classroom := strings.TrimSpace(r.URL.Query().Get("classroom"))
	if classroom == "" {
		classroom = "101"
	} else if !classroomRegex.MatchString(classroom) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid classroom %q; expected format like '101', '401'", classroom))
		return
	}

	termStr := strings.TrimSpace(r.URL.Query().Get("term"))
	term := 1
	if termStr != "" {
		if termStr == "1" {
			term = 1
		} else if termStr == "2" {
			term = 2
		} else {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid term %q; expected 1 or 2", termStr))
			return
		}
	}

	cacheKey := fmt.Sprintf("semester:%s:%d", classroom, term)
	if h.cache != nil {
		if cached, ok := h.cache.Get(cacheKey); ok {
			w.Header().Set("X-Cache", "HIT")
			writeSuccess(w, cached, "Semester classroom attendance retrieved successfully (cached)")
			return
		}
	}

	doc, err := h.client.GetSemesterClassAttendance(ctx, classroom, term)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch semester attendance: "+err.Error())
		return
	}

	result := parser.ParseSemesterClassAttendance(doc, classroom, term)
	if h.cache != nil {
		h.cache.Set(cacheKey, result, 5*time.Minute) // 5 minutes TTL
	}
	w.Header().Set("X-Cache", "MISS")
	writeSuccess(w, result, "Semester classroom attendance retrieved successfully")
}

// SearchStudents handles GET /api/students/search?q=กิตติพิชญ์
func (h *Handler) SearchStudents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeError(w, http.StatusBadRequest, "Search query 'q' parameter is required")
		return
	}
	if len(q) > 100 {
		writeError(w, http.StatusBadRequest, "Search query is too long (maximum 100 characters)")
		return
	}

	cacheKey := fmt.Sprintf("search:%s", q)
	if h.cache != nil {
		if cached, ok := h.cache.Get(cacheKey); ok {
			w.Header().Set("X-Cache", "HIT")
			writeSuccess(w, cached, "Students search completed (cached)")
			return
		}
	}

	doc, err := h.client.SearchStudents(ctx, q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to search students: "+err.Error())
		return
	}

	results := parser.ParseStudentSearchResults(doc, h.cfg.BaseURL)
	if h.cache != nil {
		h.cache.Set(cacheKey, results, 10*time.Minute) // 10 minutes TTL
	}
	w.Header().Set("X-Cache", "MISS")
	writeSuccess(w, results, "Students search completed")
}

// StudentAttendance handles GET /api/students/{sd_no}/attendance?term=1
func (h *Handler) StudentAttendance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sdNo := strings.TrimSpace(r.URL.Query().Get("sd_no"))
	if sdNo == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[3] == "attendance" {
			sdNo = pathParts[2]
		}
	}

	if sdNo == "" {
		writeError(w, http.StatusBadRequest, "Student ID 'sd_no' is required")
		return
	}
	if !numericRegex.MatchString(sdNo) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid student ID %q; must contain only numbers", sdNo))
		return
	}

	termStr := strings.TrimSpace(r.URL.Query().Get("term"))
	term := 1
	if termStr != "" {
		if termStr == "1" {
			term = 1
		} else if termStr == "2" {
			term = 2
		} else {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid term %q; expected 1 or 2", termStr))
			return
		}
	}

	cacheKey := fmt.Sprintf("student:%s:%d", sdNo, term)
	if h.cache != nil {
		if cached, ok := h.cache.Get(cacheKey); ok {
			w.Header().Set("X-Cache", "HIT")
			writeSuccess(w, cached, "Student attendance profile retrieved successfully (cached)")
			return
		}
	}

	doc, err := h.client.GetStudentAttendance(ctx, sdNo, term)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch student attendance: "+err.Error())
		return
	}

	result := parser.ParseStudentAttendanceProfile(doc, sdNo, term)
	if h.cache != nil {
		h.cache.Set(cacheKey, result, 2*time.Minute) // 2 minutes TTL
	}
	w.Header().Set("X-Cache", "MISS")
	writeSuccess(w, result, "Student attendance profile retrieved successfully")
}

// FlagCeremonyAttendance handles GET /api/attendance/flag-ceremony?date=YYYY-MM-DD
func (h *Handler) FlagCeremonyAttendance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dateStr := strings.TrimSpace(r.URL.Query().Get("date"))
	if err := validateDate(dateStr); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cacheKey := fmt.Sprintf("flag:%s", dateStr)
	if h.cache != nil {
		if cached, ok := h.cache.Get(cacheKey); ok {
			w.Header().Set("X-Cache", "HIT")
			writeSuccess(w, cached, "Flag ceremony attendance retrieved successfully (cached)")
			return
		}
	}

	doc, err := h.client.GetFlagCeremonyOverview(ctx, dateStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch flag ceremony attendance: "+err.Error())
		return
	}

	result := parser.ParseFlagCeremonyOverview(doc)
	if h.cache != nil {
		h.cache.Set(cacheKey, result, 60*time.Second) // 1 minute TTL
	}
	w.Header().Set("X-Cache", "MISS")
	writeSuccess(w, result, "Flag ceremony attendance retrieved successfully")
}
