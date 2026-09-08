package config

import (
	"os"
)

type Config struct {
	BaseURL      string
	LoginURL     string
	Port         string
	ApiKey       string
	AdminApiKey  string
	KeysFilePath string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := os.Getenv("DSCHOOL_BASE_URL")
	if baseURL == "" {
		baseURL = "https://dschool-l1.gp-education.com/dschool_app_v2020"
	}

	loginURL := os.Getenv("DSCHOOL_LOGIN_URL")
	if loginURL == "" {
		loginURL = "http://dschool-l1.gp-education.com/dschool_app_v2020/index.php?app=m&mobile_id=e4K2ixcwaVI:APA91bFqNdfS7jbA_WeiD9Zfo3YMjfgEtz_ELKerBjZAyTIFdztCFhAMM0U2oI4IneqAjuUYMexPFZPDn3nc0ZXjQs-tt5PxOPBGAMn9v1wZcWSd0T6VtKRL1K2cD8f_5HlXIFaA1qfP&gcm_regid=e4K2ixcwaVI:APA91bFqNdfS7jbA_WeiD9Zfo3YMjfgEtz_ELKerBjZAyTIFdztCFhAMM0U2oI4IneqAjuUYMexPFZPDn3nc0ZXjQs-tt5PxOPBGAMn9v1wZcWSd0T6VtKRL1K2cD8f_5HlXIFaA1qfP&type=m&school_id=1013270183&change_stat=1&latitude=0&longitude=0&v2022=8%3Bp8%3Bp8%3Bp33195susv%2C&user_id="
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "dschool-secret-key-2026"
	}

	adminApiKey := os.Getenv("ADMIN_API_KEY")
	if adminApiKey == "" {
		adminApiKey = "dschool-admin-master-key-2026"
	}

	keysFilePath := os.Getenv("KEYS_FILE_PATH")
	if keysFilePath == "" {
		keysFilePath = "data/api_keys.json"
	}

	return &Config{
		BaseURL:      baseURL,
		LoginURL:     loginURL,
		Port:         port,
		ApiKey:       apiKey,
		AdminApiKey:  adminApiKey,
		KeysFilePath: keysFilePath,
	}
}
