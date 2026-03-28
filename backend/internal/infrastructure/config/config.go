package config

import (
	"net/url"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config はアプリケーション全体の設定を保持する。
type Config struct {
	// Server
	Port           int    `env:"PORT" envDefault:"8080"`
	AllowedOrigins string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:3000"`

	// Database
	DatabaseURL string `env:"DATABASE_URL,required"`

	// JWT
	JWTSecret string        `env:"JWT_SECRET,required"`
	JWTExpiry time.Duration // Load() 内でパース

	JWTExpiryRaw string `env:"JWT_EXPIRY" envDefault:"720h"`

	// X OAuth
	XClientID     string `env:"X_CLIENT_ID,required"`
	XClientSecret string `env:"X_CLIENT_SECRET,required"`
	XRedirectURI  string `env:"X_REDIRECT_URI,required"`

	// Frontend
	FrontendURL string `env:"FRONTEND_URL" envDefault:"http://localhost:3000"`

	// Cookie
	CookieSecure bool   `env:"COOKIE_SECURE" envDefault:"false"`
	CookieDomain string // Load() 内で FrontendURL から導出

	// Anthropic (Claude API)
	AnthropicAPIKey string `env:"ANTHROPIC_API_KEY,required"`

	// X API (Bearer Token for app-only endpoints: tweet counts, search)
	XBearerToken string `env:"X_BEARER_TOKEN" envDefault:""`
}

// BatchConfig はバッチジョブ専用の設定を保持する。
// サーバー用 Config の required フィールド (JWT 等) を除いたサブセット。
type BatchConfig struct {
	DatabaseURL       string `env:"DATABASE_URL,required"`
	XBearerToken      string `env:"X_BEARER_TOKEN"`
	XClientID         string `env:"X_CLIENT_ID"`
	XClientSecret     string `env:"X_CLIENT_SECRET"`
	XRedirectURI      string `env:"X_REDIRECT_URI"`

	// Anthropic (Claude API)
	AnthropicAPIKey string `env:"ANTHROPIC_API_KEY"`
}

// LoadBatch は環境変数を読み込み BatchConfig を返す。
func LoadBatch() (*BatchConfig, error) {
	cfg := &BatchConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Load は環境変数を読み込み Config を返す。
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	d, err := time.ParseDuration(cfg.JWTExpiryRaw)
	if err != nil {
		return nil, err
	}
	cfg.JWTExpiry = d
	cfg.CookieDomain = deriveCookieDomain(cfg.FrontendURL)

	return cfg, nil
}

// deriveCookieDomain は FrontendURL からホスト名を抽出して Cookie の Domain 属性値を返す。
// localhost の場合は空文字を返す（ローカル開発では Domain 不要）。
func deriveCookieDomain(frontendURL string) string {
	u, err := url.Parse(frontendURL)
	if err != nil {
		return ""
	}
	host := u.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "" {
		return ""
	}
	return host
}
