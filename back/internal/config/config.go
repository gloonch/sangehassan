package config

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	SMSProvider                  string
	WorkerPollSeconds            int
	NotificationRetrySchedule    []time.Duration
	DBMaxOpenConns               int
	DBMaxIdleConns               int
	DBConnMaxLifetime            time.Duration
	HTTPReadTimeout              time.Duration
	HTTPWriteTimeout             time.Duration
	HTTPIdleTimeout              time.Duration
	HTTPReadHeaderTimeout        time.Duration
	MaxUploadSizeMB              int
	AppVersion                   string
	GitCommit                    string
	BuildTime                    string
}

func Load() (Config, error) {
	appEnv := getEnv("APP_ENV", "development")
	defaultUploadDir, defaultWorkflowDir := "./storage/images", "./storage/workflow-files"
	if strings.EqualFold(appEnv, "production") {
		defaultUploadDir, defaultWorkflowDir = "/app/storage/images", "/app/storage/workflow-files"
	}
	cfg := Config{
		AppEnv:                       appEnv,
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
		UploadDir:                    getEnv("UPLOAD_DIR", defaultUploadDir),
		WorkflowFileDir:              getEnv("WORKFLOW_FILE_DIR", defaultWorkflowDir),
		CatalogMinProducts:           atoiDefault(getEnv("CATALOG_MIN_PRODUCTS", "6"), 6),
		BootstrapSuperAdminPhone:     getEnv("BOOTSTRAP_SUPER_ADMIN_PHONE", ""),
		BootstrapSuperAdminPassword:  getEnv("BOOTSTRAP_SUPER_ADMIN_PASSWORD", ""),
		BootstrapSuperAdminFirstName: getEnv("BOOTSTRAP_SUPER_ADMIN_FIRST_NAME", ""),
		BootstrapSuperAdminLastName:  getEnv("BOOTSTRAP_SUPER_ADMIN_LAST_NAME", ""),
		SMSProvider:                  strings.ToLower(getEnv("SMS_PROVIDER", "disabled")),
		WorkerPollSeconds:            atoiDefault(getEnv("WORKER_POLL_SECONDS", "30"), 30),
		DBMaxOpenConns:               atoiDefault(getEnv("DB_MAX_OPEN_CONNS", "20"), 20),
		DBMaxIdleConns:               atoiDefault(getEnv("DB_MAX_IDLE_CONNS", "10"), 10),
		DBConnMaxLifetime:            durationDefault(getEnv("DB_CONN_MAX_LIFETIME", "30m"), 30*time.Minute),
		HTTPReadTimeout:              durationDefault(getEnv("HTTP_READ_TIMEOUT", "30s"), 30*time.Second),
		HTTPWriteTimeout:             durationDefault(getEnv("HTTP_WRITE_TIMEOUT", "120s"), 120*time.Second),
		HTTPIdleTimeout:              durationDefault(getEnv("HTTP_IDLE_TIMEOUT", "120s"), 120*time.Second),
		HTTPReadHeaderTimeout:        durationDefault(getEnv("HTTP_READ_HEADER_TIMEOUT", "10s"), 10*time.Second),
		MaxUploadSizeMB:              atoiDefault(getEnv("MAX_UPLOAD_SIZE_MB", "15"), 15),
		AppVersion:                   getEnv("APP_VERSION", "dev"),
		GitCommit:                    getEnv("GIT_COMMIT", "unknown"),
		BuildTime:                    getEnv("BUILD_TIME", "unknown"),
	}
	cfg.NotificationRetrySchedule = parseDurations(getEnv("NOTIFICATION_RETRY_SCHEDULE", "1m,5m,15m,1h"))

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
	if cfg.DBMaxOpenConns < 1 || cfg.DBMaxIdleConns < 0 || cfg.DBMaxIdleConns > cfg.DBMaxOpenConns {
		return Config{}, errors.New("database connection pool configuration is invalid")
	}
	if cfg.MaxUploadSizeMB < 1 || cfg.MaxUploadSizeMB > 100 {
		return Config{}, errors.New("MAX_UPLOAD_SIZE_MB must be between 1 and 100")
	}
	if strings.EqualFold(cfg.AppEnv, "production") {
		if len(cfg.JWTSecret) < 32 {
			return Config{}, errors.New("JWT_SECRET must contain at least 32 characters in production")
		}
		if cfg.DBPassword == "" {
			return Config{}, errors.New("DB_PASSWORD is required in production")
		}
		if !cfg.CookieSecure {
			return Config{}, errors.New("COOKIE_SECURE must be true in production")
		}
		if len(cfg.AllowedOrigins) == 0 {
			return Config{}, errors.New("ALLOWED_ORIGINS is required in production")
		}
		for _, origin := range cfg.AllowedOrigins {
			parsed, parseErr := url.Parse(origin)
			if parseErr != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
				return Config{}, errors.New("production ALLOWED_ORIGINS must contain exact HTTPS origins")
			}
		}
		if cfg.SMSProvider == "fake" {
			return Config{}, errors.New("SMS_PROVIDER=fake is not allowed in production")
		}
		publicDir, publicErr := filepath.Abs(cfg.UploadDir)
		privateDir, privateErr := filepath.Abs(cfg.WorkflowFileDir)
		if publicErr != nil || privateErr != nil || !filepath.IsAbs(cfg.UploadDir) || !filepath.IsAbs(cfg.WorkflowFileDir) || publicDir == string(filepath.Separator) || privateDir == string(filepath.Separator) {
			return Config{}, errors.New("production storage paths must be absolute non-root directories")
		}
		relative, relErr := filepath.Rel(publicDir, privateDir)
		if relErr != nil || relative == "." || (!strings.HasPrefix(relative, ".."+string(filepath.Separator)) && relative != "..") {
			return Config{}, errors.New("WORKFLOW_FILE_DIR must be outside the public UPLOAD_DIR")
		}
	}
	if cfg.SMSProvider != "disabled" && cfg.SMSProvider != "fake" {
		return Config{}, errors.New("SMS_PROVIDER must be disabled or fake")
	}

	return cfg, nil
}

func durationDefault(value string, fallback time.Duration) time.Duration {
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseDurations(raw string) []time.Duration {
	result := make([]time.Duration, 0, 4)
	for _, item := range strings.Split(raw, ",") {
		if d, err := time.ParseDuration(strings.TrimSpace(item)); err == nil && d > 0 {
			result = append(result, d)
		}
	}
	if len(result) == 0 {
		return []time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute, time.Hour}
	}
	return result
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
