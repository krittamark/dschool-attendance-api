package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"dschool-attendance-api/internal/models"
)

type CreateKeyRequest struct {
	Name string `json:"name"`
	Key  string `json:"key,omitempty"`
	Role string `json:"role,omitempty"` // "client" or "admin"
}

type KeyActionRequest struct {
	Key string `json:"key"`
}

// ListApiKeys handles GET /api/admin/keys
func (h *Handler) ListApiKeys(w http.ResponseWriter, r *http.Request) {
	if h.keyStore == nil {
		writeError(w, http.StatusInternalServerError, "KeyStore is not configured")
		return
	}

	keys := h.keyStore.ListKeys()
	writeSuccess(w, keys, "API keys retrieved successfully")
}

// CreateApiKey handles POST /api/admin/keys
func (h *Handler) CreateApiKey(w http.ResponseWriter, r *http.Request) {
	if h.keyStore == nil {
		writeError(w, http.StatusInternalServerError, "KeyStore is not configured")
		return
	}

	var req CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		writeError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	created, err := h.keyStore.CreateKey(strings.TrimSpace(req.Name), strings.TrimSpace(req.Key), strings.TrimSpace(req.Role))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "API Key created successfully",
		Data:    created,
	})
}

// RevokeApiKey handles POST /api/admin/keys/revoke
func (h *Handler) RevokeApiKey(w http.ResponseWriter, r *http.Request) {
	if h.keyStore == nil {
		writeError(w, http.StatusInternalServerError, "KeyStore is not configured")
		return
	}

	var req KeyActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: expected JSON with 'key'")
		return
	}

	key := strings.TrimSpace(req.Key)
	if key == "" {
		writeError(w, http.StatusBadRequest, "Parameter 'key' is required")
		return
	}

	// Protect master admin key from revocation
	if key == h.cfg.AdminApiKey {
		writeError(w, http.StatusForbidden, "Master Admin Key cannot be revoked")
		return
	}

	revoked, err := h.keyStore.RevokeKey(key)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, revoked, "API Key revoked successfully")
}

// ActivateApiKey handles POST /api/admin/keys/activate
func (h *Handler) ActivateApiKey(w http.ResponseWriter, r *http.Request) {
	if h.keyStore == nil {
		writeError(w, http.StatusInternalServerError, "KeyStore is not configured")
		return
	}

	var req KeyActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: expected JSON with 'key'")
		return
	}

	key := strings.TrimSpace(req.Key)
	if key == "" {
		writeError(w, http.StatusBadRequest, "Parameter 'key' is required")
		return
	}

	activated, err := h.keyStore.ActivateKey(key)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, activated, "API Key activated successfully")
}
