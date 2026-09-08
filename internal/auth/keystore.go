package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type KeyInfo struct {
	Key        string     `json:"key"`
	Name       string     `json:"name"`
	IsActive   bool       `json:"is_active"`
	Role       string     `json:"role"` // "admin" or "client"
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type KeyStore struct {
	filePath string
	mu       sync.RWMutex
	keys     map[string]KeyInfo
}

func GenerateRandomKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "dsk_live_" + hex.EncodeToString(bytes), nil
}

func NewKeyStore(filePath string, defaultClientKey, adminMasterKey string) (*KeyStore, error) {
	ks := &KeyStore{
		filePath: filePath,
		keys:     make(map[string]KeyInfo),
	}

	// Try loading existing file
	if err := ks.load(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Initialize with defaults
			now := time.Now()
			if defaultClientKey != "" {
				ks.keys[defaultClientKey] = KeyInfo{
					Key:       defaultClientKey,
					Name:      "Default Client Key",
					IsActive:  true,
					Role:      "client",
					CreatedAt: now,
				}
			}
			if adminMasterKey != "" {
				ks.keys[adminMasterKey] = KeyInfo{
					Key:       adminMasterKey,
					Name:      "Master Admin Key",
					IsActive:  true,
					Role:      "admin",
					CreatedAt: now,
				}
			}
			_ = ks.save()
		} else {
			return nil, fmt.Errorf("failed to load key store: %w", err)
		}
	} else {
		// Ensure master admin key exists
		if adminMasterKey != "" {
			if _, exists := ks.keys[adminMasterKey]; !exists {
				ks.keys[adminMasterKey] = KeyInfo{
					Key:       adminMasterKey,
					Name:      "Master Admin Key",
					IsActive:  true,
					Role:      "admin",
					CreatedAt: time.Now(),
				}
				_ = ks.save()
			}
		}
	}

	return ks, nil
}

func (ks *KeyStore) load() error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	data, err := os.ReadFile(ks.filePath)
	if err != nil {
		return err
	}

	var list []KeyInfo
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	ks.keys = make(map[string]KeyInfo)
	for _, item := range list {
		ks.keys[item.Key] = item
	}
	return nil
}

func (ks *KeyStore) save() error {
	// Must be called with lock held
	if err := os.MkdirAll(filepath.Dir(ks.filePath), 0755); err != nil {
		return err
	}

	list := make([]KeyInfo, 0, len(ks.keys))
	for _, v := range ks.keys {
		list = append(list, v)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ks.filePath, data, 0644)
}

func (ks *KeyStore) ValidateKey(key string) (*KeyInfo, bool) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	item, exists := ks.keys[key]
	if !exists || !item.IsActive {
		return nil, false
	}

	// Update last used timestamp
	now := time.Now()
	item.LastUsedAt = &now
	ks.keys[key] = item

	return &item, true
}

func (ks *KeyStore) IsAdminKey(key string) bool {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	item, exists := ks.keys[key]
	return exists && item.IsActive && item.Role == "admin"
}

func (ks *KeyStore) ListKeys() []KeyInfo {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	list := make([]KeyInfo, 0, len(ks.keys))
	for _, item := range ks.keys {
		list = append(list, item)
	}
	return list
}

func (ks *KeyStore) CreateKey(name string, customKey string, role string) (*KeyInfo, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	finalKey := customKey
	if finalKey == "" {
		generated, err := GenerateRandomKey()
		if err != nil {
			return nil, err
		}
		finalKey = generated
	}

	if _, exists := ks.keys[finalKey]; exists {
		return nil, fmt.Errorf("API Key already exists: %s", finalKey)
	}

	if role == "" {
		role = "client"
	}
	if name == "" {
		name = "API Client Key"
	}

	item := KeyInfo{
		Key:       finalKey,
		Name:      name,
		IsActive:  true,
		Role:      role,
		CreatedAt: time.Now(),
	}

	ks.keys[finalKey] = item
	if err := ks.save(); err != nil {
		return nil, fmt.Errorf("failed to persist key: %w", err)
	}

	return &item, nil
}

func (ks *KeyStore) RevokeKey(key string) (*KeyInfo, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	item, exists := ks.keys[key]
	if !exists {
		return nil, fmt.Errorf("API Key not found: %s", key)
	}

	now := time.Now()
	item.IsActive = false
	item.RevokedAt = &now
	ks.keys[key] = item

	if err := ks.save(); err != nil {
		return nil, fmt.Errorf("failed to persist revocation: %w", err)
	}

	return &item, nil
}

func (ks *KeyStore) ActivateKey(key string) (*KeyInfo, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	item, exists := ks.keys[key]
	if !exists {
		return nil, fmt.Errorf("API Key not found: %s", key)
	}

	item.IsActive = true
	item.RevokedAt = nil
	ks.keys[key] = item

	if err := ks.save(); err != nil {
		return nil, fmt.Errorf("failed to persist activation: %w", err)
	}

	return &item, nil
}
