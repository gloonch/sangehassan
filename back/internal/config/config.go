package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv                       string
	Port                         string
	DBHost                       string
	DBPort                       string
	DBUser                       string
	DBPassword                   string
	DBName                       string
	DBSSLMode                    string
	JWTSecret                    string
	JWTTTLHours                  int
	AccessTokenMinutes           int
	RefreshTokenDays             int
	CookieSecure                 bool
	AllowedOrigins               []string
	UploadDir                    string
	WorkflowFileDir              string
	CatalogMinProducts           int
	BootstrapSuperAdminPhone     string
	BootstrapSuperAdminPassword  string
	BootstrapSuperAdminFirstName string
	BootstrapSuperAdminLastName  string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:                       getEnv("APP_ENV", "development"),
		Port:                         getEnv("PORT", "8080"),
		DBHost:                       getEnv("DB_HOST", "localhost"),
		DBPort:                       getEnv("DB_PORT", "5432"),
		DBUser:                       getEnv("DB_USER", "postgres"),
		DBPassword:                   getEnv("DB_PASSWORD", ""),
		DBName:                       getEnv("DB_NAME", "sangehassan"),
		DBSSLMode:                    getEnv("DB_SSLMODE", "disable"),
		JWTSecret:                    getEnv("JWT_SECRET", ""),
		AccessTokenMinutes:           atoiDefault(getEnv("ACCESS_TOKEN_MINUTES", "15"), 15),
		RefreshTokenDays:             atoiDefault(getEnv("REFRESH_TOKEN_DAYS", "30"), 30),
		UploadDir:                    getEnv("UPLOAD_DIR", "./storage/images"),
		WorkflowFileDir:              getEnv("WORKFLOW_FILE_DIR", "./storage/workflow-files"),
		CatalogMinProducts:           atoiDefault(getEnv("CATALOG_MIN_PRODUCTS", "6"), 6),
		BootstrapSuperAdminPhone:     getEnv("BOOTSTRAP_SUPER_ADMIN_PHONE", ""),
		BootstrapSuperAdminPassword:  getEnv("BOOTSTRAP_SUPER_ADMIN_PASSWORD", ""),
		BootstrapSuperAdminFirstName: getEnv("BOOTSTRAP_SUPER_ADMIN_FIRST_NAME", ""),
		BootstrapSuperAdminLastName:  getEnv("BOOTSTRAP_SUPER_ADMIN_LAST_NAME", ""),
	}

	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}

	jwtHoursRaw := getEnv("JWT_TTL_HOURS", "24")
	jwtHours, err := strconv.Atoi(jwtHoursRaw)
	if err != nil {
		return Config{}, errors.New("JWT_TTL_HOURS must be an integer")
	}
	cfg.JWTTTLHours = jwtHours

	cookieSecureRaw := strings.ToLower(getEnv("COOKIE_SECURE", "false"))
	cfg.CookieSecure = cookieSecureRaw == "true" || cookieSecureRaw == "1"

	originsRaw := strings.TrimSpace(getEnv("ALLOWED_ORIGINS", ""))
	if originsRaw != "" {
		items := strings.Split(originsRaw, ",")
		for _, item := range items {
			value := strings.TrimSpace(item)
			if value != "" {
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, value)
			}
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func atoiDefault(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
