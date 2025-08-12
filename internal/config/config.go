package config

type HTTPConfig struct {
	Port string `env:"HTTP_PORT" default:"8080"`
	// Trusted proxy list (CIDR/IP). Empty slice = trust no proxies
	TrustedProxies []string `env:"HTTP_TRUSTED_PROXIES"`
	// CORS allowed origins (comma-separated). Use "*" for all in dev.
	AllowedOrigins []string `env:"HTTP_CORS_ALLOWED_ORIGINS" default:"*" envSeparator:","`
	// Security headers toggle
	SecurityHeaders bool `env:"HTTP_SECURITY_HEADERS" default:"true"`
	// Rate limit for login endpoint (RPS and burst)
	LoginRateLimitRPS   float64 `env:"HTTP_LOGIN_RATELIMIT_RPS" default:"1"`
	LoginRateLimitBurst int     `env:"HTTP_LOGIN_RATELIMIT_BURST" default:"5"`
	// When true, Redis rate limiter will deny requests on Redis errors (fail-closed). Default false (fail-open).
	LoginRateLimitFailClosed bool `env:"HTTP_LOGIN_RATELIMIT_FAIL_CLOSED" default:"false"`
	// Max body size for JSON requests (bytes)
	MaxBodyBytes int64 `env:"HTTP_MAX_BODY_BYTES" default:"1048576"`
}

type DBConfig struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Name     string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSLMODE" default:"disable"`
	// Connection pool tuning
	MaxOpenConns    int `env:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int `env:"DB_MAX_IDLE_CONNS" default:"25"`
	ConnMaxLifetime int `env:"DB_CONN_MAX_LIFETIME_SEC" default:"900"`
	ConnMaxIdleTime int `env:"DB_CONN_MAX_IDLE_TIME_SEC" default:"300"`
}

type JWTConfig struct {
	Secret    string `env:"JWT_SECRET"`
	ExpireSec int    `env:"JWT_EXPIRE_SEC" default:"3600"`
	Issuer    string `env:"JWT_ISSUER" default:"app"`
	Audience  string `env:"JWT_AUDIENCE" default:"app-clients"`
	LeewaySec int    `env:"JWT_LEEWAY_SEC" default:"30"`
	// Algorithm: HS256 (default), RS256, EdDSA
	Alg string `env:"JWT_ALG" default:"HS256"`
	// Optional key id to attach in JWT header when signing (useful for rotation)
	KID string `env:"JWT_KID"`
	// Private key for RS256/EdDSA (PEM content or file path). Prefer path in production.
	PrivateKeyPath string `env:"JWT_PRIVATE_KEY_PATH"`
	PrivateKeyPEM  string `env:"JWT_PRIVATE_KEY_PEM"`
	// Directory containing public key PEM files for verification and rotation. Filename (without extension) is treated as kid.
	PublicKeysDir string `env:"JWT_PUBLIC_KEYS_DIR"`
}

type RBACConfig struct {
	// Optional path to YAML file defining role -> permissions mapping
	PolicyPath string `env:"RBAC_POLICY_PATH"`
}

type SeedConfig struct {
	Enable    bool   `env:"SEED_ENABLE" default:"false"`
	Email     string `env:"SEED_USER_EMAIL"`
	Password  string `env:"SEED_USER_PASSWORD"`
	FirstName string `env:"SEED_USER_FIRST_NAME" default:"Admin"`
	LastName  string `env:"SEED_USER_LAST_NAME" default:"User"`
	Role      string `env:"SEED_USER_ROLE" default:"admin"`
}

type SecurityConfig struct {
	// BcryptCost allows tuning password hashing cost per environment (4-31). 0 = use library default
	BcryptCost int `env:"BCRYPT_COST" default:"0"`
	// Refresh token TTL in seconds
	RefreshTTLSeconds int `env:"REFRESH_TTL_SEC" default:"604800"`
	// Enable refresh token flow and endpoints
	RefreshEnabled bool `env:"AUTH_REFRESH_ENABLED" default:"false"`
}

type Config struct {
	Env      string `env:"ENV" default:"dev"`
	HTTP     HTTPConfig
	DB       DBConfig
	JWT      JWTConfig
	RBAC     RBACConfig
	Seed     SeedConfig
	LogLevel string `env:"LOG_LEVEL" default:"debug"`
	// Directory path for SQL migrations (default: "migrations")
	MigrationsPath string `env:"MIGRATIONS_PATH" default:"migrations"`
	// Security-related tunables
	Security SecurityConfig
	// Optional Redis for distributed features (rate limit, refresh tokens)
	RedisAddr     string `env:"REDIS_ADDR"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisDB       int    `env:"REDIS_DB" default:"0"`

	// i18n configuration
	I18nLocalesDir    string `env:"I18N_LOCALES_DIR" default:"configs/locales"`
	I18nDefaultLocale string `env:"I18N_DEFAULT_LOCALE" default:"en"`

	// pgxpool tuning (optional). If unset/non-positive, defaults are applied.
	PGXMaxConns             int `env:"PGX_MAX_CONNS" default:"0"`
	PGXMinConns             int `env:"PGX_MIN_CONNS" default:"0"`
	PGXMaxConnLifetime      int `env:"PGX_CONN_MAX_LIFETIME_SEC" default:"0"`
	PGXMaxConnIdleTime      int `env:"PGX_CONN_MAX_IDLE_TIME_SEC" default:"0"`
	PGXHealthCheckPeriodSec int `env:"PGX_HEALTH_CHECK_PERIOD_SEC" default:"0"`
}
