// server/config/config.go
package config

import (
	"log"
	"os"
	"server/internal/config/constant"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Mail     MailConfig
	Minio    MinioConfig
	Auth     AuthConfig
	OAuth    OAuthConfig
}

type AppConfig struct {
	Port      string
	Mode      string // "development" | "production"
	ClientURL string
	ServerURL string
	GlobalRateLimit constant.RateLimitConfig // max requests per minute per IP
}

type DatabaseConfig struct {
	DSN string
}

type RedisConfig struct {
	URL              string
	TokenPrefixDefault string
	AccessTokenPrefix  string
	RefreshTokenPrefix string
	RateLimitPrefix    string
}

type MailConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	EmailVerificationSubject string
	PasswordResetSubject     string
}

type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	UseSSL          bool
}

type AuthConfig struct {
	// Access token — short lived, dipakai di Authorization header
	AccessTokenSecret  string
	AccessTokenExpSec  int // default: 900 = 15 menit

	// Refresh token — long lived, opaque, disimpan di DB (hashed)
	RefreshTokenExpSec int // default: 604800 = 7 hari

	// Email verification
	VerifyEmailExpSec int // default: 86400 = 24 jam

	// Password reset
	ResetPasswordExpSec int // default: 3600 = 1 jam

	// OAuth state TTL
	OAuthStateTTLSec int // default: 600 = 10 menit
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	GithubClientID     string
	GithubClientSecret string
	GithubRedirectURL  string
}

var C *Config

// Load membaca .env lalu populate Config. Panggil sekali di main.go.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env not found, reading from OS environment")
	}

	C = &Config{
		App: AppConfig{
			Port:      getEnv("APP_PORT", "8080"),
			Mode:      getEnv("APP_MODE", "development"),
			ClientURL: getEnv("CLIENT_URL", "http://localhost:3000"),
			ServerURL: getEnv("SERVER_URL", "http://localhost:8080"),
			GlobalRateLimit: constant.RateLimitConfig{
				Limit:     getEnvInt("GLOBAL_RATE_LIMIT", 60),
				WindowSec: getEnvInt("GLOBAL_RATE_LIMIT_WINDOW", 60),
			},
		},

		Database: DatabaseConfig{
			// format: host=localhost user=postgres password=password dbname=mydb port=5432 sslmode=disable TimeZone=UTC
			DSN: getEnv("DATABASE_DSN", "host=localhost user=postgres password=password dbname=go_auth port=5432 sslmode=disable TimeZone=UTC"),
		},

		Redis: RedisConfig{
			URL:                getEnv("REDIS_URL", "redis://localhost:6379"),
			TokenPrefixDefault: getEnv("REDIS_TOKEN_PREFIX", "auth"),
			AccessTokenPrefix:  "at",
			RefreshTokenPrefix: "rt",
			RateLimitPrefix:    "rl",
		},

		Mail: MailConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnvInt("SMTP_PORT", 587),
			User:     getEnv("SMTP_USER", "user@example.com"),
			Password: getEnv("SMTP_PASS", "password"),
			From:     getEnv("SMTP_FROM", "no-reply@example.com"),
			EmailVerificationSubject: "Verify your email address",
			PasswordResetSubject:     "Reset your password",
		},

		Minio: MinioConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:          getEnv("MINIO_BUCKET", "my-bucket"),
			UseSSL:          getEnvBool("MINIO_USE_SSL", false),
		},

		Auth: AuthConfig{
			AccessTokenSecret:   getEnv("ACCESS_TOKEN_SECRET", "your-access-token-secret"),
			AccessTokenExpSec:   getEnvInt("ACCESS_TOKEN_EXP_SEC", 900),       // 15 menit
			RefreshTokenExpSec:  getEnvInt("REFRESH_TOKEN_EXP_SEC", 604800),   // 7 hari
			VerifyEmailExpSec:   getEnvInt("VERIFY_EMAIL_EXP_SEC", 86400),     // 24 jam
			ResetPasswordExpSec: getEnvInt("RESET_PASSWORD_EXP_SEC", 3600),    // 1 jam
			OAuthStateTTLSec:    getEnvInt("OAUTH_STATE_TTL_SEC", 600),        // 10 menit
		},

		OAuth: OAuthConfig{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/v1/auth/google/callback"),

			GithubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			GithubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
			GithubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/v1/auth/github/callback"),
		},
	}

	return C
}

 
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}