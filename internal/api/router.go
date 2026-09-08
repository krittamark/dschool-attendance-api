package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"dschool-attendance-api/internal/models"
)

// SetupRouter sets up http.Handler with routing and middlewares
func SetupRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	// OpenAPI and Interactive Docs (Public)
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		http.ServeFile(w, r, "openapi.json")
	})

	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml; charset=utf-8")
		http.ServeFile(w, r, "openapi.yaml")
	})

	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(swaggerUIHTML))
	})

	// Public Health Check
	mux.HandleFunc("/api/health", h.HealthCheck)

	// Admin API Endpoints for Key Management (Requires Admin Privileges)
	mux.HandleFunc("/api/admin/keys/revoke", h.RevokeApiKey)
	mux.HandleFunc("/api/admin/keys/activate", h.ActivateApiKey)
	mux.HandleFunc("/api/admin/keys", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListApiKeys(w, r)
		case http.MethodPost:
			h.CreateApiKey(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Dispatcher for /api/admin/keys/{key} (DELETE)
	mux.HandleFunc("/api/admin/keys/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")
		// Expect: ["api", "admin", "keys", "{key}"]
		if len(parts) == 4 && r.Method == http.MethodDelete {
			targetKey := parts[3]
			if targetKey == h.cfg.AdminApiKey {
				writeError(w, http.StatusForbidden, "Master Admin Key cannot be revoked")
				return
			}
			revoked, err := h.keyStore.RevokeKey(targetKey)
			if err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeSuccess(w, revoked, "API Key revoked successfully")
			return
		}
		http.NotFound(w, r)
	})

	// Protected Data Endpoints
	mux.HandleFunc("/api/attendance/daily", h.DailyAttendance)
	mux.HandleFunc("/api/attendance/overview", h.OverviewAttendance)
	mux.HandleFunc("/api/attendance/monthly", h.MonthlyAttendance)
	mux.HandleFunc("/api/attendance/semester", h.SemesterAttendance)
	mux.HandleFunc("/api/attendance/flag-ceremony", h.FlagCeremonyAttendance)
	mux.HandleFunc("/api/students/search", h.SearchStudents)
	mux.HandleFunc("/api/students/attendance", h.StudentAttendance)

	// Custom dispatcher for /api/students/ to support /api/students/{sd_no}/attendance
	mux.HandleFunc("/api/students/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")
		// Expect: ["api", "students", "{sd_no}", "attendance"]
		if len(parts) >= 4 && parts[3] == "attendance" {
			h.StudentAttendance(w, r)
			return
		}
		if len(parts) == 3 && parts[2] == "search" {
			h.SearchStudents(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// Wrap with middlewares: Auth (Admin + API Key) -> CORS -> Logger -> Recovery
	handler := authMiddleware(h)(mux)
	handler = corsMiddleware(handler)
	handler = loggingMiddleware(handler)
	handler = recoveryMiddleware(handler)

	return handler
}

func extractKey(r *http.Request) string {
	// 1. Header: X-API-Key
	key := r.Header.Get("X-API-Key")

	// 2. Header: Authorization: Bearer <key> or ApiKey <key>
	if key == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			key = strings.TrimPrefix(authHeader, "Bearer ")
		} else if strings.HasPrefix(authHeader, "ApiKey ") {
			key = strings.TrimPrefix(authHeader, "ApiKey ")
		}
	}

	// 3. Query Parameter: ?api_key=<key>
	if key == "" {
		key = r.URL.Query().Get("api_key")
	}

	return strings.TrimSpace(key)
}

func authMiddleware(h *Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Allow CORS Preflight and Public documentation / health endpoints
			if r.Method == http.MethodOptions ||
				path == "/api/health" ||
				path == "/openapi.json" ||
				path == "/openapi.yaml" ||
				path == "/docs" {
				next.ServeHTTP(w, r)
				return
			}

			key := extractKey(r)

			// 1. Admin Endpoints: Require Master Admin Key or an Admin Role Key
			if strings.HasPrefix(path, "/api/admin/") {
				isAdmin := (key == h.cfg.AdminApiKey) || (h.keyStore != nil && h.keyStore.IsAdminKey(key))
				if !isAdmin {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusForbidden)
					_ = json.NewEncoder(w).Encode(models.ApiResponse{
						Success: false,
						Error:   "Forbidden: Admin privileges required. Please provide valid Admin API Key via 'X-API-Key' or 'Authorization: Bearer <key>'",
					})
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// 2. Regular Protected Data Endpoints
			if h.keyStore != nil {
				keyInfo, valid := h.keyStore.ValidateKey(key)
				// Also allow Master Admin Key for data endpoints
				if !valid && key != h.cfg.AdminApiKey {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusUnauthorized)
					_ = json.NewEncoder(w).Encode(models.ApiResponse{
						Success: false,
						Error:   "Unauthorized: Invalid or revoked API Key. Please provide active key via 'X-API-Key' header, 'Authorization: Bearer <key>', or '?api_key=<key>'",
					})
					return
				}
				_ = keyInfo
			} else if key != h.cfg.ApiKey && key != h.cfg.AdminApiKey {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(models.ApiResponse{
					Success: false,
					Error:   "Unauthorized: Invalid API Key",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s took %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, `{"success":false,"error":"Internal Server Error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="th">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>dschool Student Attendance API - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
  <style>
    body { margin: 0; background: #fafafa; }
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/openapi.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIBundle.SwaggerUIStandalonePreset
        ],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>
`
