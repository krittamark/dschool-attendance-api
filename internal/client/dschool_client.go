package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"dschool-attendance-api/internal/config"
	"dschool-attendance-api/internal/parser"

	"github.com/PuerkitoBio/goquery"
)

const (
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

type Client struct {
	httpClient *http.Client
	cfg        *config.Config
	mu         sync.Mutex
	lastLogin  time.Time
}

func NewClient(cfg *config.Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 30,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
	}

	httpClient := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   30 * time.Second,
	}

	c := &Client{
		httpClient: httpClient,
		cfg:        cfg,
	}

	// Initial login
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := c.login(ctx); err != nil {
		return nil, fmt.Errorf("initial login failed: %w", err)
	}

	return c, nil
}

func (c *Client) login(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.LoginURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("login returned status code %d", resp.StatusCode)
	}

	c.lastLogin = time.Now()
	return nil
}

// EnsureLogin refreshes login if needed
func (c *Client) EnsureLogin(ctx context.Context) error {
	if time.Since(c.lastLogin) > 20*time.Minute {
		return c.login(ctx)
	}
	return nil
}

// RequestDocument makes a GET request and returns a goquery.Document
func (c *Client) RequestDocument(ctx context.Context, targetURL string) (*goquery.Document, error) {
	if err := c.EnsureLogin(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		if err := c.login(ctx); err == nil {
			req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
			req2.Header.Set("User-Agent", defaultUserAgent)
			resp2, err2 := c.httpClient.Do(req2)
			if err2 == nil {
				defer resp2.Body.Close()
				return goquery.NewDocumentFromReader(resp2.Body)
			}
		}
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

// SetSessionAndGet calls set_sesstion.php and follows redirect to the target page
func (c *Client) SetSessionAndGet(ctx context.Context, targetPage string, sname, sdata, sname2, sdata2 string) (*goquery.Document, error) {
	escapedTarget := strings.ReplaceAll(targetPage, "&", "[and]")
	vals := url.Values{}
	vals.Set("url", escapedTarget)
	vals.Set("sname", sname)
	vals.Set("sdata", sdata)
	if sname2 != "" {
		vals.Set("sname2", sname2)
		vals.Set("sdata2", sdata2)
	}

	setSessionURL := fmt.Sprintf("%s/set_sesstion.php?%s", c.cfg.BaseURL, vals.Encode())
	return c.RequestDocument(ctx, setSessionURL)
}

// GetOverviewAttendance gets school overview attendance for a given date
func (c *Client) GetOverviewAttendance(ctx context.Context, dateStr string) (*goquery.Document, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	targetPage := "menu_101.php?menu=system&submenu_id=101"
	formattedDate := parser.FormatDateParam(dateStr)

	return c.SetSessionAndGet(ctx, targetPage, "date0", formattedDate, "", "")
}

// GetDailyClassAttendance gets daily classroom attendance
func (c *Client) GetDailyClassAttendance(ctx context.Context, classroom, dateStr string) (*goquery.Document, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	edlevel, classCode := parser.ClassToEdLevel(classroom)
	targetPage := "menu_105.php?menu=system&submenu_id=105"

	// Always ensure date0 is set (defaults to today if dateStr is "")
	formattedDate := parser.FormatDateParam(dateStr)
	if _, err := c.SetSessionAndGet(ctx, targetPage, "date0", formattedDate, "", ""); err != nil {
		return nil, err
	}

	// Set classroom and fetch menu_105
	return c.SetSessionAndGet(ctx, targetPage, "edlevel", edlevel, "class_input", classCode)
}

// GetMonthlyClassAttendance gets monthly classroom attendance
func (c *Client) GetMonthlyClassAttendance(ctx context.Context, classroom, monthStr, yearStr string) (*goquery.Document, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	edlevel, classCode := parser.ClassToEdLevel(classroom)
	targetPage := "menu_102.php?menu=system&submenu_id=102"

	// 1. Set classroom
	if _, err := c.SetSessionAndGet(ctx, targetPage, "edlevel", edlevel, "class_input", classCode); err != nil {
		return nil, err
	}

	// 2. Set month and year if provided
	if monthStr != "" && yearStr != "" {
		return c.SetSessionAndGet(ctx, targetPage, "month0", monthStr, "year0", yearStr)
	}

	return c.RequestDocument(ctx, fmt.Sprintf("%s/%s", c.cfg.BaseURL, targetPage))
}

// GetSemesterClassAttendance gets semester classroom attendance
func (c *Client) GetSemesterClassAttendance(ctx context.Context, classroom string, term int) (*goquery.Document, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	edlevel, classCode := parser.ClassToEdLevel(classroom)
	targetPage := "menu_103.php?menu=system&submenu_id=103"

	// 1. Set classroom
	if _, err := c.SetSessionAndGet(ctx, targetPage, "edlevel", edlevel, "class_input", classCode); err != nil {
		return nil, err
	}

	// 2. Set term (0 = Term 1, 1 = Term 2)
	termVal := "0"
	if term == 2 {
		termVal = "1"
	}

	return c.SetSessionAndGet(ctx, targetPage, "term0", termVal, "", "")
}

// SearchStudents searches for a student by query
func (c *Client) SearchStudents(ctx context.Context, query string) (*goquery.Document, error) {
	searchURL := fmt.Sprintf("%s/student_360_search_student.php?search_query=%s", c.cfg.BaseURL, url.QueryEscape(query))
	return c.RequestDocument(ctx, searchURL)
}

// GetStudentAttendance gets individual student attendance report
func (c *Client) GetStudentAttendance(ctx context.Context, sdNo string, term int) (*goquery.Document, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Initialize student context in dschool session by visiting student360 profile
	initProfileURL := fmt.Sprintf("%s/student360.php?menu=student360&sd_no=%s", c.cfg.BaseURL, sdNo)
	if _, err := c.RequestDocument(ctx, initProfileURL); err != nil {
		return nil, err
	}

	// 2. Set term and fetch menu_113
	targetPage := fmt.Sprintf("menu_113.php?menu=student360&submenu_id=113&sd_no=%s", sdNo)
	termVal := "0"
	if term == 2 {
		termVal = "1"
	}

	return c.SetSessionAndGet(ctx, targetPage, "term0", termVal, "", "")
}

// GetFlagCeremonyOverview gets flag ceremony overview
func (c *Client) GetFlagCeremonyOverview(ctx context.Context, dateStr string) (*goquery.Document, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	targetPage := "menu_201.php?menu=system&submenu_id=201"
	formattedDate := parser.FormatDateParam(dateStr)
	return c.SetSessionAndGet(ctx, targetPage, "date0", formattedDate, "", "")
}
